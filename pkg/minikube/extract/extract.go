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

package extract

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
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/exit"
)

var blacklist []string = []string{"%s: %v"}

type extractor struct {
	funcs        map[string]struct{}
	fs           *stack.Stack
	translations map[string]interface{}
	currentFunc  string
	parentFunc   string
	filename     string
}

func ExtractTranslatableStrings(paths []string, functions []string, output string) {
	extractor := newExtractor(functions)

	console.OutStyle(console.Waiting, "Compiling translation strings...")
	for extractor.fs.Len() > 0 {
		f := extractor.fs.Pop().(string)
		extractor.currentFunc = f
		glog.Infof("Checking function: %s\n", f)
		for _, root := range paths {
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if shouldCheckFile(path) {
					extractor.filename = path
					return inspectFile(extractor)
				}
				return nil
			})

			if err != nil {
				exit.WithError("Extracting strings", err)
			}
		}
	}

	err := writeStringsToFiles(extractor, output)

	if err != nil {
		exit.WithError("Writing translation files", err)
	}

	console.OutStyle(console.Ready, "Done!")
}

func newExtractor(functionsToCheck []string) *extractor {
	funcs := make(map[string]struct{})
	fs := stack.New()

	for _, f := range functionsToCheck {
		funcs[f] = struct{}{}
		fs.Push(f)
	}

	return &extractor{
		funcs:        funcs,
		fs:           fs,
		translations: make(map[string]interface{}),
	}
}

func writeStringsToFiles(e *extractor, output string) error {
	err := filepath.Walk(output, func(path string, info os.FileInfo, err error) error {
		if info.Mode().IsDir() {
			return nil
		}
		console.OutStyle(console.Check, "Writing to %s", filepath.Base(path))
		var currentTranslations map[string]interface{}
		f, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "reading translation file")
		}
		err = json.Unmarshal(f, &currentTranslations)
		if err != nil {
			return errors.Wrap(err, "unmarshalling current translations")
		}

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

func addParentFuncToList(e *extractor) {
	if _, ok := e.funcs[e.parentFunc]; !ok {
		e.funcs[e.parentFunc] = struct{}{}
		e.fs.Push(e.parentFunc)
	}
}

func shouldCheckFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

func inspectFile(e *extractor) error {
	fset := token.NewFileSet()
	r, err := ioutil.ReadFile(e.filename)
	if err != nil {
		return err
	}
	glog.Infof("Parsing %s\n", e.filename)
	fmt.Printf("Parsing %s\n", e.filename)
	file, err := parser.ParseFile(fset, "", r, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(file, func(x ast.Node) bool {
		fd, ok := x.(*ast.FuncDecl)

		// Only check functions for now.
		if !ok {
			// Deal with Solutions text here

			// Deal with Cobra stuff here
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

		// Check line inside the function
		for _, stmt := range fd.Body.List {
			checkStmt(stmt, e)
		}
		return true
	})

	return nil
}

func checkStmt(stmt ast.Stmt, e *extractor) {
	// If this line is an expression, see if it's a function call
	if expr, ok := stmt.(*ast.ExprStmt); ok {
		checkCallExpression(expr, e)
	}

	// If this line is the beginning of an if statement, then check of the body of the block
	if ifstmt, ok := stmt.(*ast.IfStmt); ok {
		checkIfStmt(ifstmt, e)
	}

	// Same for loops
	if forloop, ok := stmt.(*ast.ForStmt); ok {
		for _, s := range forloop.Body.List {
			checkStmt(s, e)
		}
	}
}

func checkIfStmt(stmt *ast.IfStmt, e *extractor) {
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

func checkCallExpression(expr *ast.ExprStmt, e *extractor) {
	s, ok := expr.X.(*ast.CallExpr)

	// This line isn't a function call
	if !ok {
		return
	}

	sf, ok := s.Fun.(*ast.SelectorExpr)
	if !ok {
		addParentFuncToList(e)
		return
	}

	// Wrong function or called with no arguments.
	if e.currentFunc != sf.Sel.Name || len(s.Args) == 0 {
		return
	}

	for _, arg := range s.Args {
		// This argument is an identifier.
		if i, ok := arg.(*ast.Ident); ok {
			if checkIdentForStringValue(i, e) {
				break
			}
		}

		// This argument is a string.
		if argString, ok := arg.(*ast.BasicLit); ok {
			if addStringToList(argString.Value, e) {
				break
			}
		}
	}

}

func checkIdentForStringValue(i *ast.Ident, e *extractor) bool {
	// This identifier is nil
	if i.Obj == nil {
		return false
	}

	as, ok := i.Obj.Decl.(*ast.AssignStmt)

	// This identifier wasn't assigned anything
	if !ok {
		return false
	}

	rhs, ok := as.Rhs[0].(*ast.BasicLit)

	// This identifier was not assigned a string/basic value
	if !ok {
		return false
	}

	if addStringToList(rhs.Value, e) {
		return true
	}

	return false
}

func addStringToList(s string, e *extractor) bool {
	// Empty strings don't need translating
	if len(s) <= 2 {
		return false
	}

	// Parse out quote marks
	stringToTranslate := s[1 : len(s)-1]

	// Don't translate integers
	if _, err := strconv.Atoi(stringToTranslate); err == nil {
		return false
	}

	// Don't translate URLs
	if u, err := url.Parse(stringToTranslate); err == nil && u.Scheme != "" && u.Host != "" {
		return false
	}

	// Don't translate commands
	if strings.HasPrefix(stringToTranslate, "sudo ") {
		return false
	}

	// Don't translate blacklisted strings
	for _, b := range blacklist {
		if b == stringToTranslate {
			return false
		}
	}

	// Hooray, we can translate the string!
	e.translations[stringToTranslate] = ""
	//fmt.Printf("	%s\n", stringToTranslate)
	return true
}
