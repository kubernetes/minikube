/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/generate"
)

func TestGenerateTestDocs(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("creating temp dir failed: %v", err)
	}
	defer os.RemoveAll(tempdir)
	docPath := filepath.Join(tempdir, "tests.md")

	err = generate.TestDocs(docPath, "../../../test/integration")
	if err != nil {
		t.Fatalf("error generating test docs: %v", err)
	}
	actualContents, err := ioutil.ReadFile(docPath)
	if err != nil {
		t.Fatalf("error reading generated file: %v", err)
	}

	rest := string(actualContents)
	for rest != "" {
		rest = checkForNeedsDoc(t, rest)
	}
}

func checkForNeedsDoc(t *testing.T, content string) string {
	needs := "\nNEEDS DOC\n"
	index := strings.Index(content, needs)
	if index < 0 {
		return ""
	}

	topHalf := content[:index]
	testName := topHalf[strings.LastIndex(topHalf, "\n"):]
	t.Errorf("%s is missing a doc string.", testName)
	return content[index+len(needs):]
}
