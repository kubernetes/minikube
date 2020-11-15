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
	"k8s.io/minikube/pkg/util/lock"
)

// exclude is a list of strings to explicitly omit from translation files.
var exclude = []string{
	"{{.error}}",
	"{{.url}}",
	"  {{.url}}",
	"{{.msg}}: {{.err}}",
	"{{.key}}={{.value}}",
	"opt {{.docker_option}}",
	"kube-system",
	"env {{.docker_env}}",
	"\\n",
	"==\u003e {{.name}} \u003c==",
	"- {{.profile}}",
	"    - {{.profile}}",
}

// ErrMapFile is a constant to refer to the err_map file, which contains the Advice strings.
const ErrMapFile string = "pkg/minikube/reason/known_issues.go"

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

	// Check all key value pairs for possible help text
	if kvp, ok := stmt.(*ast.KeyValueExpr); ok {
		checkKeyValueExpression(kvp, e)
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
			if sfc, ok := sf.X.(*ast.CallExpr); ok {
				extractFlagHelpText(s, sfc, e)
				return
			}
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

// checkArguments checks the arguments of a function call for strings
func checkArguments(s *ast.CallExpr, e *state) {
	matched := false
	for _, arg := range s.Args {
		// This argument is an identifier.
		if i, ok := arg.(*ast.Ident); ok {
			if s := checkIdentForStringValue(i); s != "" {
				e.translations[s] = ""
				matched = true
				break
			}
		}

		// This argument is a string.
		if argString, ok := arg.(*ast.BasicLit); ok {
			if s := checkString(argString.Value); s != "" {
				e.translations[s] = ""
				matched = true
				break
			}
		}
	}

	// No string arguments were found, check everything the calls this function for strings
	if !matched {
		addParentFuncToList(e)
	}

}

// checkIdentForStringValue takes a identifier and sees if it's a variable assigned to a string
func checkIdentForStringValue(i *ast.Ident) string {
	// This identifier is nil
	if i.Obj == nil {
		return ""
	}

	var s string

	// This identifier was directly assigned a value
	if as, ok := i.Obj.Decl.(*ast.AssignStmt); ok {
		if rhs, ok := as.Rhs[0].(*ast.BasicLit); ok {
			s = rhs.Value
		}

	}

	// This Identifier is part of the const or var declaration
	if vs, ok := i.Obj.Decl.(*ast.ValueSpec); ok {
		for j, n := range vs.Names {
			if n.Name == i.Name {
				if len(vs.Values) < j+1 {
					// There's no way anything was assigned here, abort
					return ""
				}
				if v, ok := vs.Values[j].(*ast.BasicLit); ok {
					s = v.Value
					break
				}
			}
		}
	}

	return checkString(s)

}

// checkString checks if a string is meant to be translated
func checkString(s string) string {
	// Empty strings don't need translating
	if len(s) <= 2 {
		return ""
	}

	// Parse out quote marks
	stringToTranslate := s[1 : len(s)-1]

	// Don't translate integers
	if _, err := strconv.Atoi(stringToTranslate); err == nil {
		return ""
	}

	// Don't translate URLs
	if u, err := url.Parse(stringToTranslate); err == nil && u.Scheme != "" && u.Host != "" {
		return ""
	}

	// Don't translate commands
	if strings.HasPrefix(stringToTranslate, "sudo ") {
		return ""
	}

	// Don't translate excluded strings
	for _, e := range exclude {
		if e == stringToTranslate {
			return ""
		}
	}

	// Hooray, we can translate the string!
	return stringToTranslate
}

// checkKeyValueExpression checks all kvps for help text
func checkKeyValueExpression(kvp *ast.KeyValueExpr, e *state) {
	// The key must be an identifier
	i, ok := kvp.Key.(*ast.Ident)
	if !ok {
		return
	}

	// Specifically, it needs to be "Short" or "Long"
	if i.Name == "Short" || i.Name == "Long" {
		// The help text is directly a string, the most common case
		if help, ok := kvp.Value.(*ast.BasicLit); ok {
			s := checkString(help.Value)
			if s != "" {
				e.translations[s] = ""
			}
		}

		// The help text is assigned to a variable, only happens if it's very long
		if help, ok := kvp.Value.(*ast.Ident); ok {
			s := checkIdentForStringValue(help)
			if s != "" {
				e.translations[s] = ""
			}
		}

		// Ok now this is just a mess
		if help, ok := kvp.Value.(*ast.BinaryExpr); ok {
			s := checkBinaryExpression(help)
			if s != "" {
				e.translations[s] = ""
			}
		}
	}
}

// checkBinaryExpression checks binary expressions, stuff of the form x + y, for strings and concats them
func checkBinaryExpression(b *ast.BinaryExpr) string {
	// Check the left side
	var s string
	if l, ok := b.X.(*ast.BasicLit); ok {
		if x := checkString(l.Value); x != "" {
			s += x
		}
	}

	if i, ok := b.X.(*ast.Ident); ok {
		if x := checkIdentForStringValue(i); x != "" {
			s += x
		}
	}

	if b1, ok := b.X.(*ast.BinaryExpr); ok {
		if x := checkBinaryExpression(b1); x != "" {
			s += x
		}
	}

	//Check the right side
	if l, ok := b.Y.(*ast.BasicLit); ok {
		if x := checkString(l.Value); x != "" {
			s += x
		}
	}

	if i, ok := b.Y.(*ast.Ident); ok {
		if x := checkIdentForStringValue(i); x != "" {
			s += x
		}
	}

	if b1, ok := b.Y.(*ast.BinaryExpr); ok {
		if x := checkBinaryExpression(b1); x != "" {
			s += x
		}
	}

	return s
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
		fmt.Printf("Writing to %s", filepath.Base(path))
		currentTranslations := make(map[string]interface{})
		f, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "reading translation file")
		}
		// Unmarhsal nonempty files
		if len(f) > 0 {
			err = json.Unmarshal(f, &currentTranslations)
			if err != nil {
				return errors.Wrap(err, "unmarshalling current translations")
			}
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

		t := 0 // translated
		u := 0 // untranslated
		for k := range e.translations {
			if currentTranslations[k] != "" {
				t++
			} else {
				u++
			}
		}

		c, err := json.MarshalIndent(currentTranslations, "", "\t")
		if err != nil {
			return errors.Wrap(err, "marshalling translations")
		}
		err = lock.WriteFile(path, c, info.Mode())
		if err != nil {
			return errors.Wrap(err, "writing translation file")
		}

		fmt.Printf(" (%d translated, %d untranslated)\n", t, u)
		return nil
	})

	if err != nil {
		return err
	}

	c, err := json.MarshalIndent(e.translations, "", "\t")
	if err != nil {
		return errors.Wrap(err, "marshalling translations")
	}
	path := filepath.Join(output, "strings.txt")
	err = lock.WriteFile(path, c, 0644)
	if err != nil {
		return errors.Wrap(err, "writing translation file")
	}

	return nil
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
			s := checkString(advice.Value)
			if s != "" {
				e.translations[s] = ""
			}
		}
		return true
	})

	return nil
}

// extractFlagHelpText finds usage text for all command flags and adds them to the list to translate
func extractFlagHelpText(c *ast.CallExpr, sfc *ast.CallExpr, e *state) {
	// We're looking for calls of the form cmd.Flags().VarP()
	flags, ok := sfc.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	if flags.Sel.Name != "Flags" || len(c.Args) == 1 {
		return
	}

	// The usage text for flags is always the final argument in the Flags() call
	usage, ok := c.Args[len(c.Args)-1].(*ast.BasicLit)
	if !ok {
		// Something has gone wrong, abort
		return
	}
	s := checkString(usage.Value)
	if s != "" {
		e.translations[s] = ""
	}
}
