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
	"github.com/pkg/errors"
)

// blacklist is a list of strings to explicitly omit from translation files.
var blacklist = []string{
	"%s: %v",
	"%s.%s=%s",
	"%s/%d",
	"%s=%s",
	"%v",
	"GID:      %s",
	"MSize:    %d",
	"UID:      %s",
	"env %s",
	"opt %s",
}

// ErrMapFile is a constant to refer to the err_map file, which contains the Advice strings.
const ErrMapFile string = "pkg/minikube/problem/err_map.go"

// state is a struct that represent the current state of the extraction process
type state struct {
	// The list of functions to check for
	funcs map[funcType]struct{}

	// A stack representation of funcs for easy iteration
	fs *stack.Stack

	// The list of translatable strings, in map form for easy json marhsalling
	translations map[string]interface{}

	// The function call we're currently checking for
	currentFunc funcType

	// The function we're currently parsing
	parentFunc funcType

	// The file we're currently checking
	filename string

	// The package we're currently in
	currentPackage string
}

type funcType struct {
	pack string // The package the function is in
	name string // The name of the function
}

// newExtractor initializes state for extraction
func newExtractor(functionsToCheck []string) (*state, error) {
	funcs := make(map[funcType]struct{})
	fs := stack.New()

	for _, t := range functionsToCheck {
		// Functions must be of the form "package.function"
		t2 := strings.Split(t, ".")
		if len(t2) < 2 {
			return nil, errors.Wrap(nil, fmt.Sprintf("Invalid function string %s. Needs package name as well.", t))
		}
		f := funcType{
			pack: t2[0],
			name: t2[1],
		}
		funcs[f] = struct{}{}
		fs.Push(f)
	}

	return &state{
		funcs:        funcs,
		fs:           fs,
		translations: make(map[string]interface{}),
	}, nil
}

// SetParentFunc Sets the current parent function, along with package information
func setParentFunc(e *state, f string) {
	e.parentFunc = funcType{
		pack: e.currentPackage,
		name: f,
	}
}

// TranslatableStrings finds all strings to that need to be translated in paths and prints them out to all json files in output
func TranslatableStrings(paths []string, functions []string, output string) error {
	e, err := newExtractor(functions)

	if err != nil {
		return errors.Wrap(err, "Initializing")
	}

	fmt.Println("Compiling translation strings...")
	for e.fs.Len() > 0 {
		f := e.fs.Pop().(funcType)
		e.currentFunc = f
		for _, root := range paths {
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if shouldCheckFile(path) {
					e.filename = path
					return inspectFile(e)
				}
				return nil
			})

			if err != nil {
				return errors.Wrap(err, "Extracting strings")
			}
		}
	}

	err = writeStringsToFiles(e, output)

	if err != nil {
		return errors.Wrap(err, "Writing translation files")
	}

	fmt.Println("Done!")
	return nil
}

func shouldCheckFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

// inspectFile goes through the given file line by line looking for translatable strings
func inspectFile(e *state) error {
	fset := token.NewFileSet()
	r, err := ioutil.ReadFile(e.filename)
	if err != nil {
		return err
	}
	file, err := parser.ParseFile(fset, "", r, parser.ParseComments)
	if err != nil {
		return err
	}

	if e.filename == ErrMapFile {
		return extractAdvice(file, e)
	}

	ast.Inspect(file, func(x ast.Node) bool {
		if fi, ok := x.(*ast.File); ok {
			e.currentPackage = fi.Name.String()
			return true
		}

		if fd, ok := x.(*ast.FuncDecl); ok {
			setParentFunc(e, fd.Name.String())
			return true
		}

		checkNode(x, e)
		return true
	})

	return nil
}

// checkNode checks each node to see if it's a function call
func checkNode(stmt ast.Node, e *state) {
	// This line is a function call, that's what we care about
	if expr, ok := stmt.(*ast.CallExpr); ok {
		checkCallExpression(expr, e)
	}
}

// checkCallExpression takes a function call, and checks its arguments for strings
func checkCallExpression(s *ast.CallExpr, e *state) {
	for _, arg := range s.Args {
		// This argument is a function literal, check its body.
		if fl, ok := arg.(*ast.FuncLit); ok {
			for _, stmt := range fl.Body.List {
				checkNode(stmt, e)
			}
		}
	}

	var functionName string
	var packageName string

	// SelectorExpr is a function call to a separate package
	sf, ok := s.Fun.(*ast.SelectorExpr)
	if ok {
		// Parse out the package of the call
		sfi, ok := sf.X.(*ast.Ident)
		if !ok {
			return
		}
		packageName = sfi.Name
		functionName = sf.Sel.Name
	}

	// Ident is an identifier, in this case it's a function call in the same package
	id, ok := s.Fun.(*ast.Ident)
	if ok {
		functionName = id.Name
		packageName = e.currentPackage
	}

	// This is not a function call.
	if len(functionName) == 0 {
		return
	}

	// This is not the correct function call, or it was called with no arguments.
	if e.currentFunc.name != functionName || e.currentFunc.pack != packageName || len(s.Args) == 0 {
		return
	}

	checkArguments(s, e)
}

func checkArguments(s *ast.CallExpr, e *state) {
	matched := false
	for _, arg := range s.Args {
		// This argument is an identifier.
		if i, ok := arg.(*ast.Ident); ok {
			if checkIdentForStringValue(i, e) {
				matched = true
				break
			}
		}

		// This argument is a string.
		if argString, ok := arg.(*ast.BasicLit); ok {
			if addStringToList(argString.Value, e) {
				matched = true
				break
			}
		}
	}

	if !matched {
		addParentFuncToList(e)
	}

}

// checkIdentForStringValye takes a identifier and sees if it's a variable assigned to a string
func checkIdentForStringValue(i *ast.Ident, e *state) bool {
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

// addStringToList takes a string, makes sure it's meant to be translated then adds it to the list if so
func addStringToList(s string, e *state) bool {
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
	return true
}

// writeStringsToFiles writes translations to all translation files in output
func writeStringsToFiles(e *state, output string) error {
	err := filepath.Walk(output, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "accessing path")
		}
		if info.Mode().IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
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

		// Make sure to not overwrite already translated strings
		for k := range e.translations {
			if _, ok := currentTranslations[k]; !ok {
				currentTranslations[k] = ""
			}
		}

		// Remove translations from the file that are empty and were not extracted
		for k, v := range currentTranslations {
			if _, ok := e.translations[k]; !ok && len(v.(string)) == 0 {
				delete(currentTranslations, k)
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

// addParentFuncToList adds the current parent function to the list of functions to inspect more closely.
func addParentFuncToList(e *state) {
	if _, ok := e.funcs[e.parentFunc]; !ok {
		e.funcs[e.parentFunc] = struct{}{}
		e.fs.Push(e.parentFunc)
	}
}

// extractAdvice specifically extracts Advice strings in err_map.go, since they don't conform to our normal translatable string format.
func extractAdvice(f ast.Node, e *state) error {
	ast.Inspect(f, func(x ast.Node) bool {
		// We want the "Advice: <advice string>" key-value pair
		// First make sure we're looking at a kvp
		kvp, ok := x.(*ast.KeyValueExpr)
		if !ok {
			return true
		}

		// Now make sure we're looking at an Advice kvp
		i, ok := kvp.Key.(*ast.Ident)
		if !ok {
			return true
		}

		if i.Name == "Advice" {
			// At this point we know the value in the kvp is guaranteed to be a string
			advice, _ := kvp.Value.(*ast.BasicLit)
			addStringToList(advice.Value, e)
		}
		return true
	})

	return nil
}
