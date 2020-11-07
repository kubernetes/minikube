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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
	"k8s.io/minikube/pkg/generate"
)

func TestGenerateDocs(t *testing.T) {
	pflag.BoolP("help", "h", false, "") // avoid 'Docs are not updated. Please run `make generate-docs` to update commands documentation' error
	dir := "../../../site/content/en/docs/commands/"

	for _, sc := range RootCmd.Commands() {
		t.Run(sc.Name(), func(t *testing.T) {
			if sc.Hidden {
				t.Skip()
			}
			fp := filepath.Join(dir, fmt.Sprintf("%s.md", sc.Name()))
			expectedContents, err := ioutil.ReadFile(fp)
			if err != nil {
				t.Fatalf("Docs are not updated. Please run `make generate-docs` to update commands documentation: %v", err)
			}
			actualContents, err := generate.DocForCommand(sc)
			if err != nil {
				t.Fatalf("error getting contents: %v", err)
			}
			if diff := cmp.Diff(actualContents, string(expectedContents)); diff != "" {
				t.Fatalf("Docs are not updated. Please run `make generate-docs` to update commands documentation: %s", diff)
			}
		})
	}
}
