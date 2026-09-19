/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/reason"
)

func TestViewCommand(t *testing.T) {
	if os.Getenv("MINIKUBE_VIEW_TEST_CHILD") == "1" {
		viewFormat = os.Getenv("MINIKUBE_VIEW_TEST_FORMAT")
		if err := View(); err != nil {
			t.Fatal(err)
		}
		os.Exit(0)
	}

	tests := []struct {
		name       string
		format     string
		config     map[string]interface{}
		wantOutput string
		wantExit   bool
	}{
		{name: "invalid format and empty config", format: "{{", wantExit: true},
		{name: "valid format and empty config", format: "{{.ConfigKey}}", wantOutput: ""},
		{name: "valid format and populated config", format: "- {{.ConfigKey}}: {{.ConfigValue}}\n", config: map[string]interface{}{"driver": "docker"}, wantOutput: "- driver: docker\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			if len(tt.config) != 0 {
				configDir := filepath.Join(home, ".minikube", "config")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatal(err)
				}
				data, err := json.Marshal(tt.config)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(configDir, "config.json"), data, 0600); err != nil {
					t.Fatal(err)
				}
			}

			cmd := exec.Command(os.Args[0], "-test.run=^TestViewCommand$")
			cmd.Env = append(os.Environ(),
				"MINIKUBE_VIEW_TEST_CHILD=1",
				"MINIKUBE_VIEW_TEST_FORMAT="+tt.format,
				localpath.MinikubeHome+"="+home)
			output, err := cmd.CombinedOutput()
			if tt.wantExit {
				exitErr, ok := err.(*exec.ExitError)
				if !ok || exitErr.ExitCode() != reason.InternalViewTmpl.ExitCode || !strings.Contains(string(output), reason.InternalViewTmpl.ID) {
					t.Fatalf("View error = %v, output = %q; want exit %d containing %s", err, output, reason.InternalViewTmpl.ExitCode, reason.InternalViewTmpl.ID)
				}
				return
			}
			if err != nil {
				t.Fatalf("View failed: %v; output: %s", err, output)
			}
			if string(output) != tt.wantOutput {
				t.Fatalf("View output = %q, want %q", output, tt.wantOutput)
			}
		})
	}
}
