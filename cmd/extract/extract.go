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

/* This file scans all of minikube's code and finds all strings that need to be able to be translated.
It uses the more generic extract.TranslatableStringd, and prints all the translations
into every json file it can find in the translations directory.

Usage: from the root minikube directory, go run cmd/extract/extract.go
*/

package main

import (
	"os"
	"strings"

	"k8s.io/minikube/pkg/minikube/extract"
)

func main() {
	paths := []string{"cmd", "pkg"}
	functions := []string{"translate.T"}
	outDir := "translations"
	cwd, err := os.Getwd()
	if err != nil {
		panic("Getting current working directory failed")
	}

	if strings.Contains(cwd, "cmd") {
		panic("run extract.go from the minikube root directory")
	}

	if _, err = os.Stat(extract.ErrMapFile); os.IsNotExist(err) {
		panic("err_map.go doesn't exist")
	}
	err = extract.TranslatableStrings(paths, functions, outDir)

	if err != nil {
		panic(err)
	}
}
