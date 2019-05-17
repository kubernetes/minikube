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
)

var funcs []string

func main() {
	paths := []string{"../pkg/minikube/", "../cmd/minikube"}
	funcs = []string{"OutStyle", "ErrStyle"}
	for _, f := range funcs {
		for _, root := range paths {
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if strings.HasSuffix(path, ".go") {
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
	//fmt.Printf("Parsing %s\n", filename)
	file, err := parser.ParseFile(fset, "", r, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(file, func(x ast.Node) bool {
		fd, ok := x.(*ast.FuncDecl)
		if !ok {
			return true
		}

		fmt.Println(fd.Name)
		for _, stmt := range fd.Body.List {
			fmt.Printf("   %s\n", stmt)
			/*if !ok {
				return true
			}

			if strings.Contains(fmt.Sprintf("%s", s.Fun), f) {
				argString := fmt.Sprintf("%s", s.Args[1])
				if strings.Contains(argString, "\"") {
					fmt.Printf("%s\n", argString[strings.Index(argString, "\""):strings.LastIndex(argString, "\"")+1])
				} else {
					funcs = append(funcs)
				}
			}*/
		}
		return true
	})

	return nil
}
