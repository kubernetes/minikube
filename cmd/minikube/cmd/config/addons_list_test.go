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

	"k8s.io/minikube/pkg/minikube/assets"
	pkgConfig "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/out"
)

func captureStdout(t *testing.T, f func()) string {
	t.Helper()

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

	f()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe: %v", err)
	}

	return <-done
}

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
			s := captureStdout(t, func() {
				printAddonsList(nil, tt.printDocs)
			})
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
			// │         ADDON NAME          │              DESCRIPTION               │
			// ├─────────────────────────────┼────────────────────────────────────────┤
			// ┌─────────────────────────────┬────────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────┐
			// │         ADDON NAME          │              DESCRIPTION               │                                     DOCS                                      │
			// ├─────────────────────────────┼────────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────┤

			expected := tt.want
			if pipeCount != expected {
				t.Errorf("Expected header to have %d pipes; got = %d: %q", expected, pipeCount, got)
			}
			if !strings.Contains(s, "DESCRIPTION") {
				t.Errorf("expected output to contain DESCRIPTION header: %q", s)
			}
			if strings.Contains(s, "MAINTAINER") {
				t.Errorf("expected output to not contain MAINTAINER header: %q", s)
			}
			if !strings.Contains(s, "Web interface for managing Kubernetes resources") {
				t.Errorf("expected output to contain dashboard description: %q", s)
			}
			if !tt.printDocs && strings.Count(s, "https://minikube.sigs.k8s.io/docs/handbook/dashboard/") != 1 {
				t.Errorf("expected dashboard docs URL to appear exactly once: %q", s)
			}
			if !strings.Contains(s, "Runs a local container image registry") {
				t.Errorf("expected output to contain registry description without docs URL: %q", s)
			}
			if strings.Contains(s, "3rd party (Ambassador)") {
				t.Errorf("expected output to not contain maintainer metadata in description column: %q", s)
			}
		})
	}

	t.Run("ExistingClusterTableDescription", func(t *testing.T) {
		cc := &pkgConfig.ClusterConfig{
			Name:   "minikube",
			Addons: map[string]bool{"dashboard": false},
		}
		s := captureStdout(t, func() {
			printAddonsList(cc, false)
		})
		if !strings.Contains(s, "DESCRIPTION") {
			t.Errorf("expected output to contain DESCRIPTION header: %q", s)
		}
		if strings.Contains(s, "MAINTAINER") {
			t.Errorf("expected output to not contain MAINTAINER header: %q", s)
		}
		if !strings.Contains(s, "Web interface for managing Kubernetes resources") {
			t.Errorf("expected output to contain dashboard description: %q", s)
		}
		if strings.Count(s, "https://minikube.sigs.k8s.io/docs/handbook/dashboard/") != 1 {
			t.Errorf("expected dashboard docs URL to appear exactly once: %q", s)
		}
		if strings.Contains(s, "3rd party (Ambassador)") {
			t.Errorf("expected output to not contain maintainer metadata in description column: %q", s)
		}
	})

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

func TestAddonDescription(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "ambassador",
			want: "API gateway and ingress controller\nhttps://minikube.sigs.k8s.io/docs/handbook/addons/ambassador/",
		},
		{
			name: "auto-pause",
			want: "Pauses idle Kubernetes workloads automatically",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addon := assets.Addons[tt.name]
			originalDescription := addon.Description
			originalDocs := addon.Docs
			got := addonDescription(addon)
			if got != tt.want {
				t.Errorf("expected description %q, got %q", tt.want, got)
			}
			if strings.TrimSpace(got) != got {
				t.Errorf("expected description to not have surrounding spaces: %q", got)
			}
			if strings.Contains(got, "  ") {
				t.Errorf("expected description to not contain double spaces: %q", got)
			}
			if addon.Docs != "" && strings.Count(got, addon.Docs) != 1 {
				t.Errorf("expected docs URL to be appended exactly once: %q", got)
			}
			if addon.Description != originalDescription {
				t.Errorf("expected stored description to remain %q, got %q", originalDescription, addon.Description)
			}
			if addon.Docs != originalDocs {
				t.Errorf("expected stored docs URL to remain %q, got %q", originalDocs, addon.Docs)
			}
		})
	}
}

func TestFormatAddonDescription(t *testing.T) {
	tests := []struct {
		name             string
		description      string
		documentationURL string
		want             string
	}{
		{
			name:             "with URL",
			description:      "API gateway and ingress controller",
			documentationURL: "https://minikube.sigs.k8s.io/docs/handbook/addons/ambassador/",
			want:             "API gateway and ingress controller\nhttps://minikube.sigs.k8s.io/docs/handbook/addons/ambassador/",
		},
		{
			name:             "without URL",
			description:      "Pauses idle Kubernetes workloads automatically",
			documentationURL: "",
			want:             "Pauses idle Kubernetes workloads automatically",
		},
		{
			name:             "trims inputs",
			description:      " API gateway and ingress controller ",
			documentationURL: " https://minikube.sigs.k8s.io/docs/handbook/addons/ambassador/ ",
			want:             "API gateway and ingress controller\nhttps://minikube.sigs.k8s.io/docs/handbook/addons/ambassador/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAddonDescription(tt.description, tt.documentationURL)
			if got != tt.want {
				t.Errorf("expected description %q, got %q", tt.want, got)
			}
			if strings.TrimSpace(got) != got {
				t.Errorf("expected description to not have surrounding spaces: %q", got)
			}
			if strings.HasSuffix(got, "\n") {
				t.Errorf("expected description to not have trailing newline: %q", got)
			}
			if tt.documentationURL == "" && strings.Contains(got, "\n") {
				t.Errorf("expected empty docs URL to not add an extra line: %q", got)
			}
			if tt.documentationURL != "" && strings.Count(got, strings.TrimSpace(tt.documentationURL)) != 1 {
				t.Errorf("expected docs URL to appear exactly once: %q", got)
			}
		})
	}
}
