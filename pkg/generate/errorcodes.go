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
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/out"
)

func ErrorCodes(docPath string, pathsToCheck []string) error {
	buf := bytes.NewBuffer([]byte{})
	date := time.Now().Format("2006-01-02")
	title := out.Fmt(title, out.V{"Command": "Error Codes", "Description": "minikube error codes and strings", "Date": date})
	_, err := buf.Write([]byte(title))
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	for _, pathToCheck := range pathsToCheck {
		r, err := ioutil.ReadFile(pathToCheck)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error reading file %s", pathToCheck))
		}
		file, err := parser.ParseFile(fset, "", r, parser.ParseComments)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("error parsing file %s", pathToCheck))
		}

		if strings.Contains(pathToCheck, "exitcodes.go") {
			buf.WriteString("## Error Codes\n\n")
			currentGroup := ""
			currentError := ""
			ast.Inspect(file, func(x ast.Node) bool {
				if c, ok := x.(*ast.Comment); ok {
					// Start a new group of errors
					comment := c.Text
					if !strings.HasPrefix(comment, "// Error codes specific") {
						return true
					}
					currentGroup = strings.Replace(comment, "//", "###", 1)
					buf.WriteString("\n" + currentGroup + "\n")
				}
				if id, ok := x.(*ast.Ident); ok {
					// This is the name of the error, e.g. ExGuestError
					currentError = id.Name
				}
				if s, ok := x.(*ast.BasicLit); ok {
					// Filter out random strings that aren't error codes
					if currentError == "" {
						return true
					}

					// No specific group means generic errors
					if currentGroup == "" {
						currentGroup = "### Generic Errors"
						buf.WriteString("\n" + currentGroup + "\n")
					}

					// This is the numeric code of the error, e.g. 80 for ExGuest Error
					code := s.Value
					buf.WriteString(fmt.Sprintf("%s: %s  \n", code, currentError))
				}
				return true
			})
			buf.WriteString("\n\n")
		}

		if strings.Contains(pathToCheck, "reason.go") {
			buf.WriteString("## Error Strings\n\n")
			currentNode := ""
			currentID := ""
			currentComment := ""
			ast.Inspect(file, func(x ast.Node) bool {
				if id, ok := x.(*ast.Ident); ok {
					currentNode = id.Name
					if strings.HasPrefix(currentNode, "Ex") && currentNode != "ExitCode" {
						// We have all the info we're going to get on this error, print it out
						buf.WriteString(fmt.Sprintf("%s (Exit code %v)  \n", currentID, currentNode))
						if currentComment != "" {
							buf.WriteString(currentComment + "  \n")
						}
						buf.WriteString("\n")
						currentComment = ""
						currentID = ""
						currentNode = ""
					}
				}
				if s, ok := x.(*ast.BasicLit); ok {
					if currentNode == "ID" {
						currentID = s.Value
					}
				}
				if c, ok := x.(*ast.Comment); ok {
					currentComment = c.Text[3:]
				}
				return true
			})
		}
	}

	return ioutil.WriteFile(docPath, buf.Bytes(), 0o644)
}
