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

package config

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/out"
)

func TestAddonsList(t *testing.T) {
	tests := []struct {
		name      string
		printDocs bool
		want      int
	}{
		{"DisabledDocs", false, 9},
		{"EnabledDocs", true, 12},
	}

	for _, tt := range tests {
		t.Run("NonExistingClusterTable"+tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			old := os.Stdout
			defer func() { os.Stdout = old }()
			os.Stdout = w
			printAddonsList(nil, tt.printDocs)
			if err := w.Close(); err != nil {
				t.Fatalf("failed to close pipe: %v", err)
			}
			buf := bufio.NewScanner(r)
			pipeCount := 0
			got := ""
			// Pull the first 3 lines from stdout
			for i := 0; i < 3; i++ {
				if !buf.Scan() {
					t.Fatalf("failed to read stdout")
				}
				pipeCount += strings.Count(buf.Text(), "|")
				got += buf.Text()
			}
			if err := buf.Err(); err != nil {
				t.Errorf("failed to read stdout: %v", err)
			}
			// The lines we pull should look something like
			// |------------|------------|(------|)
			// | ADDON NAME | MAINTAINER |( DOCS |)
			// |------------|------------|(------|)
			// which has 9 or 12 pipes
			expected := tt.want
			if pipeCount != expected {
				t.Errorf("Expected header to have %d pipes; got = %d: %q", expected, pipeCount, got)
			}
		})
	}

	t.Run("NonExistingClusterJSON", func(t *testing.T) {
		type addons struct {
			Ambassador *interface{} `json:"ambassador"`
		}

		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}
		old := os.Stdout
		defer func() {
			os.Stdout = old
			out.SetOutFile(os.Stdout)
		}()
		os.Stdout = w
		out.SetOutFile(os.Stdout)
		printAddonsJSON(nil)
		if err := w.Close(); err != nil {
			t.Fatalf("failed to close pipe: %v", err)
		}
		got := addons{}
		dec := json.NewDecoder(r)
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if got.Ambassador == nil {
			t.Errorf("expected `ambassador` field to not be nil, but was")
		}
	})
}
