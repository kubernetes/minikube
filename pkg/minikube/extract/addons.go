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
	"path"
	"unicode"
	"unicode/utf8"

	"gopkg.in/yaml.v2"
	"k8s.io/minikube/pkg/version"
)

func AddonImages() error {
	fset := token.NewFileSet()
	r, err := os.ReadFile("pkg/minikube/assets/addons.go")
	if err != nil {
		return err
	}

	addonToImages := make(map[string]interface{})
	currentAddon := ""
	file, err := parser.ParseFile(fset, "", r, parser.ParseComments)
	if err != nil {
		return err
	}
	ast.Inspect(file, func(x ast.Node) bool {
		if kvp, ok := x.(*ast.KeyValueExpr); ok {
			if key, ok := kvp.Key.(*ast.BasicLit); ok {
				k := key.Value[1 : len(key.Value)-1]
				first, _ := utf8.DecodeRuneInString(k)
				// This is an addon name
				if unicode.IsLower(first) {
					currentAddon = k
				} else if unicode.IsUpper(first) {
					// This is a variable name pointing to an image

					// Special-case storage-provisioner since it's hard to parse
					if k == "StorageProvisioner" {
						addonToImages[currentAddon] = map[string]string{k: fmt.Sprintf("gcr.io/k8s-minikube/storage-provisioner:%s", version.GetStorageProvisionerVersion())}
						return true
					}
					if v, ok := kvp.Value.(*ast.BasicLit); ok {
						val := v.Value[1 : len(v.Value)-1]
						if _, ok := addonToImages[currentAddon]; !ok {
							// This is the first image we've found for this addon
							addonToImages[currentAddon] = map[string]string{k: val}
						} else {
							ca := addonToImages[currentAddon].(map[string]string)
							if i, ok := ca[k]; ok {
								// We have an explicit registry definition, prepend it to the image already in the map
								i = path.Join(val, i)
								ca[k] = i
							} else {
								// This is a new image to append to the list for this addon
								ca[k] = val
							}
							addonToImages[currentAddon] = ca
						}
					}
				}
			}
		}
		return true
	})

	y, err := yaml.Marshal(addonToImages)
	if err != nil {
		return err
	}
	return os.WriteFile("hack/addons/addons.yaml", y, 0777)
}
