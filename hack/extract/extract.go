/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	"github.com/pkg/errors"
)

var Blacklist []string = []string{"%s: %v"}

type Extractor struct {
	funcs        map[string]struct{}
	fs           *stack.Stack
	translations map[string]interface{}
	currentFunc  string
	parentFunc   string
	filename     string
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if !strings.HasSuffix(cwd, "minikube") {
		fmt.Println("Please run extractor from top level minikube directory.")
		fmt.Println("Usage: go run hack/extract/extract.go")
		os.Exit(1)
	}

	paths := []string{"cmd", "pkg"}
	//paths := []string{"../cmd/minikube/cmd/delete.go"}
	//paths := []string{"../pkg/minikube/cluster/cluster.go"}
	extractor := newExtractor("Translate")

	fmt.Println("Compiling translation strings...")
	for extractor.fs.Len() > 0 {
		f := extractor.fs.Pop().(string)
		extractor.currentFunc = f
		//fmt.Printf("-----%s------\n", f)
		for _, root := range paths {
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if shouldCheckFile(path) {
					extractor.filename = path
					return inspectFile(extractor)
				}
				return nil
			})

			if err != nil {
				panic(err)
			}
		}
	}

	err = writeStringsToFiles(extractor)

	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")
}

func newExtractor(functionsToCheck ...string) *Extractor {
	funcs := make(map[string]struct{})
	fs := stack.New()

	for _, f := range functionsToCheck {
		funcs[f] = struct{}{}
		fs.Push(f)
	}

	return &Extractor{
		funcs:        funcs,
		fs:           fs,
		translations: make(map[string]interface{}),
	}
}

func writeStringsToFiles(e *Extractor) error {
	translationsFiles := "pkg/minikube/translate/translations"
	err := filepath.Walk(translationsFiles, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsDir() {
			return nil
		}
		fmt.Printf("Writing to %s\n", filepath.Base(path))
		var currentTranslations map[string]interface{}
		f, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "reading translation file")
		}
		err = json.Unmarshal(f, &currentTranslations)
		if err != nil {
			return errors.Wrap(err, "unmarshalling current translations")
		}

		//fmt.Println(currentTranslations)

		for k := range e.translations {
			//fmt.Println(k)
			if _, ok := currentTranslations[k]; !ok {
				currentTranslations[k] = ""
			}
		}

		c, err := json.MarshalIndent(currentTranslations, "", "\t")
		if err != nil {
			return errors.Wrap(err, "marshalling translations")
		}
		err = ioutil.WriteFile(path, c, info.Mode())
		if err != nil {
			return errors.Wrap(err, "writing translation file")
		}
		return nil
	})

	return err
}

func addParentFuncToList(e *Extractor) {
	if _, ok := e.funcs[e.parentFunc]; !ok {
		e.funcs[e.parentFunc] = struct{}{}
		e.fs.Push(e.parentFunc)
	}
}

func shouldCheckFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

func inspectFile(e *Extractor) error {
	fset := token.NewFileSet()
	r, err := ioutil.ReadFile(e.filename)
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

		e.parentFunc = fd.Name.String()

		for _, stmt := range fd.Body.List {
			checkStmt(stmt, e)
		}
		return true
	})

	return nil
}

func checkStmt(stmt ast.Stmt, e *Extractor) {
	// If this line is an expression, see if it's a function call
	if t, ok := stmt.(*ast.ExprStmt); ok {
		checkCallExpression(t, e)
	}

	// If this line is the beginning of an if statement, then check of the body of the block
	if b, ok := stmt.(*ast.IfStmt); ok {
		checkIfStmt(b, e)
	}

	// Same for loops
	if b, ok := stmt.(*ast.ForStmt); ok {
		for _, s := range b.Body.List {
			checkStmt(s, e)
		}
	}
}

func checkIfStmt(stmt *ast.IfStmt, e *Extractor) {
	for _, s := range stmt.Body.List {
		checkStmt(s, e)
	}
	if stmt.Else != nil {
		// A straight else
		if block, ok := stmt.Else.(*ast.BlockStmt); ok {
			for _, s := range block.List {
				checkStmt(s, e)
			}
		}

		// An else if
		if elseif, ok := stmt.Else.(*ast.IfStmt); ok {
			checkIfStmt(elseif, e)
		}

	}
}

func checkCallExpression(t *ast.ExprStmt, e *Extractor) {
	if s, ok := t.X.(*ast.CallExpr); ok {
		sf, ok := s.Fun.(*ast.SelectorExpr)
		if !ok {
			addParentFuncToList(e)
			return
		}
		if e.currentFunc == sf.Sel.Name && len(s.Args) > 0 {
			addParentFuncToList(e)
			for _, arg := range s.Args {
				// Find references to strings
				if i, ok := arg.(*ast.Ident); ok {
					if i.Obj != nil {
						if as, ok := i.Obj.Decl.(*ast.AssignStmt); ok {
							if rhs, ok := as.Rhs[0].(*ast.BasicLit); ok {
								if addStringToList(rhs.Value, e) {
									break
								}
							}
						}
					}
				}
				// Find string arguments
				if argString, ok := arg.(*ast.BasicLit); ok {
					if addStringToList(argString.Value, e) {
						break
					}
				}
			}
		}
	}
}

func addStringToList(s string, e *Extractor) bool {
	// Don't translate empty strings
	if len(s) > 2 {
		// Parse out quote marks
		stringToTranslate := s[1 : len(s)-1]
		// Don't translate integers
		if _, err := strconv.Atoi(stringToTranslate); err != nil {
			//Don't translate URLs
			if u, err := url.Parse(stringToTranslate); err != nil || u.Scheme == "" || u.Host == "" {
				// Don't translate commands
				if !strings.HasPrefix(stringToTranslate, "sudo ") {
					// Don't translate blacklisted strings
					for _, b := range Blacklist {
						if b == stringToTranslate {
							return false
						}
					}
					e.translations[stringToTranslate] = ""
					//fmt.Printf("	%s\n", stringToTranslate)
					return true
				}
			}
		}
	}
	return false
}
