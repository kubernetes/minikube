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
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func main() {
	validateSchema("deploy/minikube/schema.json", "deploy/minikube/releases.json")
	validateSchema("deploy/minikube/schema.json", "deploy/minikube/releases-beta.json")
	validateSchema("deploy/minikube/schema-v2.json", "deploy/minikube/releases-v2.json")
	validateSchema("deploy/minikube/schema-v2.json", "deploy/minikube/releases-beta-v2.json")
	os.Exit(0)
}

func validateSchema(schemaPathString, docPathString string) {
	sch, err := jsonschema.Compile(schemaPathString)
	if err != nil {
		log.Fatal(err)
	}

	data, err := os.ReadFile(docPathString)
	if err != nil {
		log.Fatal(err)
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		log.Fatal(err)
	}

	if err = sch.Validate(v); err != nil {
		fmt.Printf("The document %s is invalid, see errors:\n%#v", docPathString, err)
		return
	}

	fmt.Printf("The document %s is valid\n", docPathString)
}
