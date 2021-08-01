/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package generate

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/out"
)

func TestDocs(docPath string, pathToCheck string) error {
	counter := 0
	buf := bytes.NewBuffer([]byte{})
	date := time.Now().Format("2006-01-02")
	title := out.Fmt(title, out.V{"Command": "Integration Tests", "Description": "All minikube integration tests", "Date": date})
	_, err := buf.Write([]byte(title))
	if err != nil {
		return err
	}

	err = filepath.Walk(pathToCheck, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		r, e := ioutil.ReadFile(path)
		if e != nil {
			return errors.Wrap(e, fmt.Sprintf("error reading file %s", path))
		}
		file, e := parser.ParseFile(fset, "", r, parser.ParseComments)
		if e != nil {
			return errors.Wrap(e, fmt.Sprintf("error parsing file %s", path))
		}

		ast.Inspect(file, func(x ast.Node) bool {
			if fd, ok := x.(*ast.FuncDecl); ok {
				fnName := fd.Name.Name
				if !shouldParse(fnName) {
					return true
				}

				if strings.HasPrefix(fnName, "valid") {
					e := writeSubTest(fnName, buf)
					if e != nil {
						return false
					}
				} else {
					e := writeTest(fnName, buf)
					if e != nil {
						return false
					}
				}

				counter++
				comments := fd.Doc
				if comments == nil {
					e := writeComment(fnName, "// NEEDS DOC\n", buf)
					return e == nil
				}
				for _, comment := range comments.List {
					if strings.Contains(comment.Text, "TODO") {
						continue
					}
					e := writeComment(fnName, comment.Text, buf)
					if e != nil {
						return false
					}
				}
				_, e := buf.WriteString("\n")
				if e != nil {
					return false
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(docPath, buf.Bytes(), 0o644)
	return err
}

func shouldParse(name string) bool {
	if strings.HasPrefix(name, "Test") && !strings.HasPrefix(name, "TestMain") {
		return true
	}

	if strings.HasPrefix(name, "valid") {
		return true
	}

	return false
}

func writeTest(testName string, w *bytes.Buffer) error {
	_, err := w.WriteString("## " + testName + "\n")
	return err
}

func writeSubTest(testName string, w *bytes.Buffer) error {
	_, err := w.WriteString("#### " + testName + "\n")
	return err
}

func writeComment(testName string, comment string, w *bytes.Buffer) error {
	// Remove the leading // from the testdoc comments
	comment = comment[3:]
	comment = strings.TrimPrefix(comment, testName+" ")
	_, err := w.WriteString(comment + "\n")
	return err
}
