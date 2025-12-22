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
	"encoding/json"
	"io"
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
		{"DisabledDocs", false, 3},
		{"EnabledDocs", true, 4},
	}

	for _, tt := range tests {
		t.Run("NonExistingClusterTable"+tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			defer r.Close()
			old := os.Stdout
			defer func() {
				os.Stdout = old
				out.SetOutFile(old)
			}()
			os.Stdout = w
			out.SetOutFile(w)

			done := make(chan string, 1)
			go func() {
				b, _ := io.ReadAll(r)
				done <- string(b)
			}()

			printAddonsList(nil, tt.printDocs)

			if err := w.Close(); err != nil {
				t.Fatalf("failed to close pipe: %v", err)
			}

			s := <-done
			lines := strings.Split(s, "\n")
			if len(lines) < 3 {
				t.Fatalf("failed to read stdout: got %d lines: %q", len(lines), s)
			}

			pipeCount := 0
			got := ""
			for i := 0; i < 3; i++ {
				pipeCount += strings.Count(lines[i], "│")
				got += lines[i]
			}
			// ┌─────────────────────────────┬────────────────────────────────────────┐
			// │         ADDON NAME          │               MAINTAINER               │
			// ├─────────────────────────────┼────────────────────────────────────────┤
			// ┌─────────────────────────────┬────────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────┐
			// │         ADDON NAME          │               MAINTAINER               │                                     DOCS                                      │
			// ├─────────────────────────────┼────────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────┤

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
