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

package main

import (
	"errors"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLHelpers(t *testing.T) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(`
name: CI
jobs:
  build:
    steps:
      - name: Test
`), &root); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	doc := root.Content[0]

	name, err := GetString(doc, "name")
	if err != nil {
		t.Fatalf("GetString name: %v", err)
	}
	if name != "CI" {
		t.Fatalf("name = %q, want CI", name)
	}

	jobsNode, err := Get(doc, "jobs")
	if err != nil {
		t.Fatalf("Get jobs: %v", err)
	}
	jobs, err := Entries(jobsNode)
	if err != nil {
		t.Fatalf("Entries jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].Key != "build" {
		t.Fatalf("jobs = %#v, want one build job", jobs)
	}

	stepsNode, err := Get(jobs[0].Value, "steps")
	if err != nil {
		t.Fatalf("Get steps: %v", err)
	}
	steps, err := Elements(stepsNode)
	if err != nil {
		t.Fatalf("Elements steps: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("len(steps) = %d, want 1", len(steps))
	}

	if err := SetScalar(steps[0], "timeout-minutes", "5"); err != nil {
		t.Fatalf("SetScalar timeout-minutes: %v", err)
	}
	timeout, err := GetString(steps[0], "timeout-minutes")
	if err != nil {
		t.Fatalf("GetString timeout-minutes: %v", err)
	}
	if timeout != "5" {
		t.Fatalf("timeout-minutes = %q, want 5", timeout)
	}
}

func TestSetScalarAppendKeepsTrailingComment(t *testing.T) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(`name: Test
run: make test
# trailing comment
`), &root); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if err := SetScalar(root.Content[0], "timeout-minutes", "15"); err != nil {
		t.Fatalf("SetScalar: %v", err)
	}

	out, err := EncodeYAML(&root)
	if err != nil {
		t.Fatalf("EncodeYAML: %v", err)
	}

	want := `name: Test
run: make test
timeout-minutes: 15
# trailing comment
`
	if string(out) != want {
		t.Fatalf("EncodeYAML =\n%s\nwant\n%s", out, want)
	}
}

func TestYAMLHelperErrors(t *testing.T) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(`
name:
  nested: value
jobs: []
`), &root); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	doc := root.Content[0]

	if _, err := Get(doc, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get missing error = %v, want ErrNotFound", err)
	}
	if _, err := GetString(doc, "name"); err == nil || !strings.Contains(err.Error(), "got mapping, want scalar") {
		t.Fatalf("GetString name error = %v, want wrong kind", err)
	}
	jobsNode, err := Get(doc, "jobs")
	if err != nil {
		t.Fatalf("Get jobs: %v", err)
	}
	if _, err := Entries(jobsNode); err == nil || !strings.Contains(err.Error(), "got sequence, want mapping") {
		t.Fatalf("Entries jobs error = %v, want wrong kind", err)
	}
}
