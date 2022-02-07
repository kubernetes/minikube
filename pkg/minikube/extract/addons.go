/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

func AddonImages() error {
	fset := token.NewFileSet()
	r, err := os.ReadFile("pkg/minikube/assets/addons.go")
	if err != nil {
		return err
	}

	file, err := parser.ParseFile(fset, "", r, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(file, func(x ast.Node) bool {
		if kvp, ok := x.(*ast.KeyValueExpr); ok {
			fmt.Println(kvp.Key)
		}
		return true
	})
	return nil
}
