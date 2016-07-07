/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
)

func main() {
	validateSchema("deploy/minikube/schema.json", "deploy/minikube/releases.json")
	validateSchema("deploy/minikube/k8s_schema.json", "deploy/minikube/k8s_releases.json")
	os.Exit(0)
}

func validateSchema(schemaPathString string, docPathString string) {
	schemaPath, _ := filepath.Abs(schemaPathString)
	schemaUri := "file://" + schemaPath
	schemaLoader := gojsonschema.NewReferenceLoader(schemaUri)

	docPath, _ := filepath.Abs(docPathString)
	docUri := "file://" + docPath
	docLoader := gojsonschema.NewReferenceLoader(docUri)

	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		panic(err.Error())
	}

	if result.Valid() {
		fmt.Printf("The document %s is valid\n", docPathString)
	} else {
		fmt.Printf("The document %s is not valid. see errors :\n", docPathString)
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		os.Exit(1)
	}
}
