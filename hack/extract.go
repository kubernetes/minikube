package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-collections/collections/stack"
)

var funcs map[string]struct{}
var funcStack *stack.Stack
var translations map[string]interface{}

func main() {
	translations = make(map[string]interface{})
	paths := []string{"../cmd", "../pkg"}
	//paths := []string{"../cmd/minikube/cmd/delete.go"}
	//paths := []string{"../pkg/minikube/cluster/cluster.go"}
	funcs = map[string]struct{}{"Translate": {}}
	funcStack = stack.New()
	for f := range funcs {
		funcStack.Push(f)
	}

	for funcStack.Len() > 0 {
		f := funcStack.Pop().(string)
		fmt.Printf("-----%s------\n", f)
		for _, root := range paths {
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if shouldCheckFile(path) {
					return inspectFile(path, f)
				}
				return nil
			})

			if err != nil {
				panic(err)
			}
		}
	}

	translationsFiles := "../pkg/minikube/translate/translations"
	err := filepath.Walk(translationsFiles, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsDir() {
			return nil
		}
		fmt.Println(path)
		var currentTranslations map[string]interface{}
		f, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		err = json.Unmarshal(f, &currentTranslations)
		if err != nil {
			return err
		}

		fmt.Println(currentTranslations)

		for k := range translations {
			fmt.Println(k)
			if _, ok := currentTranslations[k]; !ok {
				currentTranslations[k] = " "
			}
		}

		c, err := json.MarshalIndent(currentTranslations, "", "\t")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path, c, info.Mode())
		return err
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")
}

func addFuncToList(f string) {
	if _, ok := funcs[f]; !ok {
		funcs[f] = struct{}{}
		funcStack.Push(f)
	}
}

func shouldCheckFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

func inspectFile(filename string, f string) error {
	fset := token.NewFileSet()
	r, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	//fmt.Printf("    Parsing %s\n", filename)
	file, err := parser.ParseFile(fset, "", r, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(file, func(x ast.Node) bool {
		fd, ok := x.(*ast.FuncDecl)
		if !ok {
			/*gd, ok := x.(*ast.GenDecl)
			if !ok {
				return true
			}
			for _, spec := range gd.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, v := range vs.Values {
						if ue, ok := v.(*ast.UnaryExpr); ok {
							fmt.Printf("%s: %s\n", ue.X, reflect.TypeOf(ue.X))
						}
					}
				}
			}*/
			return true
		}

		for _, stmt := range fd.Body.List {
			checkStmt(stmt, fd.Name.String(), f)
		}
		return true
	})

	return nil
}

func checkStmt(stmt ast.Stmt, parentFunc string, f string) {
	// If this line is an expression, see if it's a function call
	if t, ok := stmt.(*ast.ExprStmt); ok {
		checkCallExpression(t, parentFunc, f)
	}

	// If this line is the beginning of an if statment, then check of the body of the block
	if b, ok := stmt.(*ast.IfStmt); ok {
		checkIfStmt(b, parentFunc, f)
	}

	// Same for loops
	if b, ok := stmt.(*ast.ForStmt); ok {
		for _, s := range b.Body.List {
			checkStmt(s, parentFunc, f)
		}
	}
}

func checkIfStmt(stmt *ast.IfStmt, parentFunc string, f string) {
	for _, s := range stmt.Body.List {
		checkStmt(s, parentFunc, f)
	}
	if stmt.Else != nil {
		// A straight else
		if block, ok := stmt.Else.(*ast.BlockStmt); ok {
			for _, s := range block.List {
				checkStmt(s, parentFunc, f)
			}
		}

		// An else if
		if elseif, ok := stmt.Else.(*ast.IfStmt); ok {
			checkIfStmt(elseif, parentFunc, f)
		}

	}
}

func checkCallExpression(t *ast.ExprStmt, parentFunc string, f string) {
	if s, ok := t.X.(*ast.CallExpr); ok {
		sf, ok := s.Fun.(*ast.SelectorExpr)
		if !ok {
			addFuncToList(parentFunc)
			return
		}
		if f == sf.Sel.Name && len(s.Args) > 0 {
			addFuncToList(parentFunc)
			for _, arg := range s.Args {
				// Find references to strings
				if i, ok := arg.(*ast.Ident); ok {
					if i.Obj != nil {
						if as, ok := i.Obj.Decl.(*ast.AssignStmt); ok {
							if rhs, ok := as.Rhs[0].(*ast.BasicLit); ok {
								fmt.Printf("	%s\n", rhs.Value)
								translations[rhs.Value] = ""
								break
							}
						}
					}
				}
				// Find string arguments
				if argString, ok := arg.(*ast.BasicLit); ok {
					stringToTranslate := argString.Value
					// Don't translate integers
					if _, err := strconv.Atoi(stringToTranslate); err != nil {
						//Don't translate URLs
						if u, err := url.Parse(stringToTranslate[1 : len(stringToTranslate)-1]); err != nil || u.Scheme == "" || u.Host == "" {
							// Don't translate empty strings
							if len(stringToTranslate) > 2 {
								translations[stringToTranslate] = ""
								fmt.Printf("	%s\n", stringToTranslate)
								break
							}
						}
					}
				}
			}
		}
	}
}
