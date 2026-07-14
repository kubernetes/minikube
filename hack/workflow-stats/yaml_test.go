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
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func parseYAML(t *testing.T, input string) *yaml.Node {
	t.Helper()

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(input), &root); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	return &root
}

func documentNode(t *testing.T, root *yaml.Node) *yaml.Node {
	t.Helper()

	if len(root.Content) != 1 {
		t.Fatalf("got %d document nodes, want 1", len(root.Content))
	}
	return root.Content[0]
}

func firstWorkflowStep(t *testing.T, doc *yaml.Node, jobName string) *yaml.Node {
	t.Helper()

	job := Get(Get(doc, "jobs"), jobName)
	if job == nil {
		t.Fatalf("missing job %q", jobName)
	}
	steps := Elements(Get(job, "steps"))
	if len(steps) == 0 {
		t.Fatalf("job %q has no steps", jobName)
	}
	return steps[0]
}

func TestYAMLNodeHelpers(t *testing.T) {
	root := parseYAML(t, `
name: CI
jobs:
  build:
    steps:
      - name: Checkout
        uses: actions/checkout@v4
`)
	doc := documentNode(t, root)

	if got := GetString(doc, "name"); got != "CI" {
		t.Fatalf("GetString(name) = %q, want CI", got)
	}

	jobs := Entries(Get(doc, "jobs"))
	if len(jobs) != 1 {
		t.Fatalf("len(jobs) = %d, want 1", len(jobs))
	}
	if jobs[0].Key != "build" {
		t.Fatalf("job key = %q, want build", jobs[0].Key)
	}

	steps := Elements(Get(jobs[0].Value, "steps"))
	if len(steps) != 1 {
		t.Fatalf("len(steps) = %d, want 1", len(steps))
	}
	if got := GetString(steps[0], "uses"); got != "actions/checkout@v4" {
		t.Fatalf("step uses = %q, want actions/checkout@v4", got)
	}
}

func TestSetScalarAddsTimeoutMinutes(t *testing.T) {
	root := parseYAML(t, `
name: CI
jobs:
  build:
    steps:
      - name: Long running step
        run: sleep 300
`)
	step := firstWorkflowStep(t, documentNode(t, root), "build")

	SetScalar(step, "timeout-minutes", "45")

	if got := GetString(step, "timeout-minutes"); got != "45" {
		t.Fatalf("timeout-minutes = %q, want 45", got)
	}

	out, err := EncodeYAML(root)
	if err != nil {
		t.Fatalf("EncodeYAML: %v", err)
	}
	text := string(out)
	if !strings.Contains(text, "timeout-minutes: 45") {
		t.Fatalf("encoded YAML missing timeout-minutes: 45:\n%s", text)
	}
}

func TestSetScalarUpdatesTimeoutMinutes(t *testing.T) {
	root := parseYAML(t, `
name: CI
jobs:
  build:
    steps:
      - name: Long running step
        timeout-minutes: 20 # keep timeout comment
        run: sleep 300
`)
	step := firstWorkflowStep(t, documentNode(t, root), "build")

	SetScalar(step, "timeout-minutes", "45")

	out, err := EncodeYAML(root)
	if err != nil {
		t.Fatalf("EncodeYAML: %v", err)
	}
	text := string(out)
	for _, want := range []string{
		"timeout-minutes: 45",
		"# keep timeout comment",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("encoded YAML missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "timeout-minutes: 20") {
		t.Fatalf("encoded YAML still has old timeout:\n%s", text)
	}
}

func TestEncodeYAMLPreservesWorkflowComments(t *testing.T) {
	root := parseYAML(t, `
# workflow comment
name: CI
on: # trigger comment
  pull_request:
jobs:
  # jobs comment
  build:
    steps:
      # step comment
      - name: Checkout # name comment
        uses: actions/checkout@v4
`)
	step := firstWorkflowStep(t, documentNode(t, root), "build")

	SetScalar(step, "timeout-minutes", "2")

	out, err := EncodeYAML(root)
	if err != nil {
		t.Fatalf("EncodeYAML: %v", err)
	}
	text := string(out)

	for _, want := range []string{
		"# workflow comment",
		"on:",
		"# trigger comment",
		"# jobs comment",
		"# step comment",
		"# name comment",
		"timeout-minutes: 2",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("encoded YAML missing %q:\n%s", want, text)
		}
	}
}

func TestSetScalarAppendsNewKeyWithoutReorderingExistingKeys(t *testing.T) {
	root := parseYAML(t, `
name: CI
jobs:
  build:
    steps:
      - name: Checkout
        if: always()
        uses: actions/checkout@v4
`)
	step := firstWorkflowStep(t, documentNode(t, root), "build")

	SetScalar(step, "timeout-minutes", "2")

	keys := make([]string, 0, len(Entries(step)))
	for _, entry := range Entries(step) {
		keys = append(keys, entry.Key)
	}
	want := []string{"name", "if", "uses", "timeout-minutes"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys = %v, want %v", keys, want)
		}
	}
}
