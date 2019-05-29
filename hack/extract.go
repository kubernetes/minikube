package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-collections/collections/stack"
)

var funcs map[string]struct{}
var funcStack *stack.Stack

func main() {
	paths := []string{"../pkg/minikube/", "../cmd/minikube/"}
	//paths := []string{".."}
	funcs = map[string]struct{}{"OutStyle": {}, "ErrStyle": {}}
	funcStack = stack.New()
	funcStack.Push("OutStyle")
	funcStack.Push("ErrStyle")

	for funcStack.Len() > 0 {
		f := funcStack.Pop().(string)
		fmt.Printf("-----%s------\n", f)
		for _, root := range paths {
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
					return inspectFile(path, f)
				}
				return nil
			})

			if err != nil {
				panic(err)
			}
		}
	}
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
			return true
		}

		//fmt.Println(fd.Name)
		for _, stmt := range fd.Body.List {
			//fmt.Printf("   %s\n", stmt)

			//fmt.Printf("	%s: %s\n", reflect.TypeOf(stmt), stmt)
			checkStmt(stmt, fd.Name.String(), f)
		}
		return true
	})

	return nil
}

func checkStmt(stmt ast.Stmt, fd string, f string) {
	// If this line is an expression, see if it's a function call
	if t, ok := stmt.(*ast.ExprStmt); ok {
		checkCallExpression(t, fd, f)
	}

	// If this line is the beginning of an if statment, then check of the body of the block
	if b, ok := stmt.(*ast.IfStmt); ok {
		for _, s := range b.Body.List {
			checkStmt(s, fd, f)
		}
	}

	// Same for loops
	if b, ok := stmt.(*ast.ForStmt); ok {
		for _, s := range b.Body.List {
			checkStmt(s, fd, f)
		}
	}
}
func checkCallExpression(t *ast.ExprStmt, fd string, f string) {
	if s, ok := t.X.(*ast.CallExpr); ok {
		if strings.Contains(fmt.Sprintf("%s", s.Fun), f) && !strings.Contains(fmt.Sprintf("%s", s.Fun), "glog") && len(s.Args) > 0 {
			//fmt.Printf("	%s\n", fd.Name)
			//fmt.Printf("	%s\n", s.Args)
			// If this print call has a string with quote marks, then print that out as a phrase to translate.
			hit := false
			for i, arg := range s.Args {
				// Dealing with inconsistent argument ordering
				if (strings.Contains(f, "Style") || strings.Contains(f, "Fatal")) && i == 0 {
					continue
				}
				argString := fmt.Sprintf("%s", arg)
				//fmt.Println(argString)
				if strings.Contains(argString, "\"") {
					fmt.Printf("	%s\n", argString[strings.Index(argString, "\""):strings.LastIndex(argString, "\"")+1])
					hit = true
					break
				}
			}

			// Otherwise, this call used variables passed in from a previous call, add that function to the list of ones to check.
			if !hit {
				if _, ok := funcs[fd]; !ok {
					//fmt.Printf("	ADDING %s TO LIST\n", fd.Name)
					funcs[fd] = struct{}{}
					funcStack.Push(fd)
				}
			}
		}
	}
}
