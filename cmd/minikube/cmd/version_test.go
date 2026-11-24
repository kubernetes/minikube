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

package cmd

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

var updateVersionTestdata = flag.Bool("update-version-testdata", false, "update version command testdata files")

// formatVersion takes an output format and a short flag. These constants make
// the call sites read naturally instead of using bare "" and booleans.
const (
	defaultFormat = ""
	jsonFormat    = "json"
	yamlFormat    = "yaml"
	longForm      = false
	shortForm     = true
)

func TestFormatVersion(t *testing.T) {
	kicbaseImages := []string{
		"quay.io/minikube/kicbase:v1.39.0@sha256:4a1c825b61479e6c898851ea66f13c620aaeab6002746e95067fc2c4b38a0b24",
		"docker.io/minikube/kicbase:v1.39.0@sha256:4a1c825b61479e6c898851ea66f13c620aaeab6002746e95067fc2c4b38a0b24",
	}
	withComponents := versionInfo{
		Version: "v1.39.0",
		Commit:  "937f6eb62e261321adbf97c87618d74817e54c54",
		Kicbase: kicbaseImages,
		Components: map[string]string{
			"buildctl":    "buildctl github.com/moby/buildkit v0.34.2 8c579433824224dc64a158cf8734495b6e2b186a",
			"containerd":  "containerd github.com/containerd/containerd/v2 v2.4.2 b7a25261e029d68978bbf290df0a62fc819b3c51",
			"cri-dockerd": "cri-dockerd 0.4.10 (2527b04)",
			"crictl":      "crictl version v1.35.0",
			"crio":        "crio version 1.35.0",
			"crun":        "crun version 1.18.2",
			"ctr":         "ctr github.com/containerd/containerd/v2 v2.4.2",
			"docker":      "Docker version 28.5.2, build ecc6942",
			"dockerd":     "Docker version 28.5.2, build 89c5e8f",
			"podman":      "podman version 3.4.7",
			"runc":        "runc version 1.5.3",
		},
	}
	withoutComponents := versionInfo{
		Version: withComponents.Version,
		Commit:  withComponents.Commit,
		Kicbase: withComponents.Kicbase,
	}

	t.Run("default", func(t *testing.T) {
		got, err := formatVersion(withoutComponents, defaultFormat, longForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "default.txt", got)
	})

	t.Run("default-empty-commit", func(t *testing.T) {
		got, err := formatVersion(versionInfo{Version: withComponents.Version, Kicbase: withComponents.Kicbase}, defaultFormat, longForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "default-empty-commit.txt", got)
	})

	t.Run("default-components", func(t *testing.T) {
		got, err := formatVersion(withComponents, defaultFormat, longForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "default-components.txt", got)
	})

	// --short prints only the raw version.
	t.Run("short", func(t *testing.T) {
		got, err := formatVersion(withoutComponents, defaultFormat, shortForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "short.txt", got)
	})

	// --short suppresses component collection, and the default text formatter
	// ignores component data even if it is present.
	t.Run("short-components", func(t *testing.T) {
		got, err := formatVersion(withComponents, defaultFormat, shortForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "short-components.txt", got)
	})

	// --short only short-circuits the default text output. JSON is still
	// rendered, but the command skipped component collection because --short
	// was set, so no component data is present.
	t.Run("short-json", func(t *testing.T) {
		got, err := formatVersion(withoutComponents, jsonFormat, shortForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "short-json.json", got)
	})

	// --short only short-circuits the default text output. YAML is still
	// rendered, but the command skipped component collection because --short
	// was set, so no component data is present.
	t.Run("short-yaml", func(t *testing.T) {
		got, err := formatVersion(withoutComponents, yamlFormat, shortForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "short-yaml.yaml", got)
	})

	t.Run("json", func(t *testing.T) {
		got, err := formatVersion(withoutComponents, jsonFormat, longForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "json.json", got)
	})

	t.Run("yaml", func(t *testing.T) {
		got, err := formatVersion(withoutComponents, yamlFormat, longForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "yaml.yaml", got)
	})

	t.Run("components-json", func(t *testing.T) {
		got, err := formatVersion(withComponents, jsonFormat, longForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "components-json.json", got)
	})

	t.Run("components-yaml", func(t *testing.T) {
		got, err := formatVersion(withComponents, yamlFormat, longForm)
		if err != nil {
			t.Fatalf("formatVersion: %v", err)
		}
		checkTestdata(t, "components-yaml.yaml", got)
	})
}

func checkTestdata(t *testing.T, name, got string) {
	t.Helper()

	path := filepath.Join("testdata", "version", name)
	if *updateVersionTestdata {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("failed to update testdata file: %v", err)
		}
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read testdata file %s: %v", path, err)
	}
	if diff := cmp.Diff(string(want), got); diff != "" {
		t.Errorf("output mismatch for %s (-want +got):\n%s", name, diff)
	}
	checkParsable(t, name, string(want))
}

func checkParsable(t *testing.T, name, output string) {
	t.Helper()

	switch filepath.Ext(name) {
	case ".json":
		var decoded interface{}
		if err := json.Unmarshal([]byte(output), &decoded); err != nil {
			t.Errorf("%s is not valid JSON: %v", name, err)
		}
	case ".yaml":
		var decoded interface{}
		if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
			t.Errorf("%s is not valid YAML: %v", name, err)
		}
	}
}
