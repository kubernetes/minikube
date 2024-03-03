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
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/out"
)

// TestDocs generates list of tests
func TestDocs(docPath string, pathToCheck string) error {
	buf := bytes.NewBuffer([]byte{})
	date := time.Now().Format("2006-01-02")
	title := out.Fmt(title, out.V{"Command": "List of Integration Test Cases", "Description": "Auto generated list of all minikube integration tests and what they do.", "Date": date})
	_, err := buf.Write([]byte(title))
	if err != nil {
		return err
	}

	err = filepath.Walk(pathToCheck, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		r, e := os.ReadFile(path)
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
				td := parseFuncDocs(file, fd)
				_, e := buf.WriteString(td.toMarkdown())
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

	err = os.WriteFile(docPath, buf.Bytes(), 0o644)
	return err
}

// TestDoc is the documentation for a test case
type TestDoc struct {
	// name is the name of the test case
	name string
	// isSubTest is true if the test case is a top-level test case, false if it's a validation method
	isSubTest bool
	// description is parsed from the function comment
	description string
	// steps are parsed from comments starting with `docs: `
	steps []string
	// specialCases are parsed from comments starting with `docs(special): `
	specialCases []string
	// specialCases are parsed from comments starting with `docs(skip): `
	skips []string
}

// toMarkdown converts the TestDoc into a string in Markdown format
func (d *TestDoc) toMarkdown() string {
	b := &strings.Builder{}
	if d.isSubTest {
		b.WriteString("#### " + d.name + "\n")
	} else {
		b.WriteString("## " + d.name + "\n")
	}

	b.WriteString(d.description + "\n")
	if len(d.steps) > 0 {
		b.WriteString("Steps:\n")
		for _, s := range d.steps {
			b.WriteString("- " + s + "\n")
		}
		b.WriteString("\n")
	}

	if len(d.specialCases) > 0 {
		b.WriteString("Special cases:\n")
		for _, s := range d.specialCases {
			b.WriteString("- " + s + "\n")
		}
		b.WriteString("\n")
	}

	if len(d.skips) > 0 {
		b.WriteString("Skips:\n")
		for _, s := range d.skips {
			b.WriteString("- " + s + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// docsRegex is the regex of the docs comment starting with either `docs: ` or  `docs(...): `
var docsRegex = regexp.MustCompile(`docs(?:\((.*?)\))?:\s*`)

// parseFuncDocs parses the comments from a function starting with `docs`
func parseFuncDocs(file *ast.File, fd *ast.FuncDecl) TestDoc {
	d := TestDoc{
		name:        fd.Name.Name,
		description: strings.TrimPrefix(fd.Doc.Text(), fd.Name.Name+" "),
		isSubTest:   strings.HasPrefix(fd.Name.Name, "valid"),
	}

	for _, c := range file.Comments {
		for _, ci := range c.List {
			if ci.Pos() < fd.Pos() || ci.End() > fd.End() {
				// only generate docs for comments that are within the function scope
				continue
			}
			text := strings.TrimPrefix(ci.Text, "// ")
			m := docsRegex.FindStringSubmatch(text)
			if len(m) < 2 {
				// comment doesn't start with `docs: ` or `docs(...): `
				continue
			}
			matched := m[0]
			docsType := m[1]

			text = strings.TrimPrefix(text, matched)
			switch docsType {
			case "special":
				d.specialCases = append(d.specialCases, text)
			case "skip":
				d.skips = append(d.skips, text)
			case "":
				d.steps = append(d.steps, text)
			default:
				log.Printf("docs type %s is not recognized", docsType)
			}
		}
	}
	return d
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
