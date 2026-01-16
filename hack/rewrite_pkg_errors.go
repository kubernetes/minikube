package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	filesProcessed := 0
	filesModified := 0
	skippedCalls := 0

	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "vendor/") {
			return nil
		}
		if info.IsDir() {
			if info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip self
		if strings.Contains(path, "rewrite_pkg_errors.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		usesPkgErrors := false
		for _, imp := range node.Imports {
			if imp.Path.Value == "\"github.com/pkg/errors\"" {
				usesPkgErrors = true
				imp.Path.Value = "\"errors\""
			}
		}

		if !usesPkgErrors {
			return nil
		}
		filesProcessed++

		rewritten := false

		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != "errors" {
				return true
			}

			if sel.Sel.Name == "Wrap" || sel.Sel.Name == "Wrapf" {
				if len(call.Args) < 2 {
					return true
				}
				errArg := call.Args[0]

				// Safety Check: If errArg is a function call, replacing with fmt.Errorf is UNSAFE
				if _, isCall := errArg.(*ast.CallExpr); isCall {
					fmt.Printf("WARNING: Skipping unsafe errors.%s rewrite in %s line %d: error arg is a function call\n",
						sel.Sel.Name, path, fset.Position(call.Pos()).Line)
					skippedCalls++
					// Reset import (partial fix might leave file broken if we don't fix this call)
					// But we already changed the import in the loop above.
					// We will manually fix these.
					return true
				}

				// errors.Wrap(err, "msg") -> fmt.Errorf("msg: %w", err)
				msgIndex := 1
				restArgs := call.Args[2:]

				ident.Name = "fmt"
				sel.Sel.Name = "Errorf"

				msgArg := call.Args[msgIndex]

				if lit, ok := msgArg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					val := lit.Value
					if len(val) >= 2 {
						inner := val[1 : len(val)-1]
						newVal := fmt.Sprintf("\"%s: %%w\"", inner)
						lit.Value = newVal

						newArgs := []ast.Expr{lit}
						newArgs = append(newArgs, restArgs...)
						newArgs = append(newArgs, errArg)
						call.Args = newArgs
						rewritten = true
						return true
					}
				}

				if len(restArgs) == 0 {
					fmtStr := &ast.BasicLit{Kind: token.STRING, Value: "\"%s: %w\""}
					call.Args = []ast.Expr{fmtStr, msgArg, errArg}
					rewritten = true
				} else {
					fmt.Printf("WARNING: Skipped complex errors.Wrapf in %s line %d\n", path, fset.Position(call.Pos()).Line)
					skippedCalls++
					ident.Name = "errors"
					sel.Sel.Name = "Wrapf"
				}
			} else if sel.Sel.Name == "Errorf" {
				ident.Name = "fmt"
				rewritten = true
			}

			return true
		})

		if rewritten || usesPkgErrors {
			// Even if we didn't rewrite any calls (e.g. only errors.New or skipped calls),
			// we changed the import path to "errors".
			// If we skipped calls, the file will fail to compile (errors.Wrap not defined).
			// This is desired so we find them.
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			if err := format.Node(f, fset, node); err != nil {
				f.Close()
				return err
			}
			f.Close()
			filesModified++
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Processed %d files, modified %d files, skipped %d calls\n", filesProcessed, filesModified, skippedCalls)
}
