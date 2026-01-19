/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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
	"io"
	"os"
	"testing"

	"github.com/google/go-github/v81/github"
)

func TestClassify(t *testing.T) {
	cfg := Config{
		SkipPrefixes:  []string{"ci:", "test:", "build(deps):"},
		PrefixGroups:  map[string]string{"addon:": "addons"},
		ContainGroups: map[string]string{"bug": "bug fix", "fix": "bug fix"},
		AllowedLabels: []string{"kind/feature", "kind/bug"},
	}
	allowed := map[string]struct{}{"kind/feature": {}, "kind/bug": {}}

	tests := []struct {
		name      string
		title     string
		labels    []string
		wantGroup string
		wantSkip  bool
	}{
		{"skip prefix", "ci: update", nil, "", true},
		{"prefix group", "addon: new addon", nil, "addons", false},
		{"contain group", "fix critical bug", nil, "bug fix", false},
		{"contain over label", "bug fix", []string{"kind/feature"}, "bug fix", false},
		{"label group", "some change", []string{"kind/bug"}, "kind/bug", false},
		{"fallback", "random", nil, "other", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &github.PullRequest{Title: github.String(tt.title)}
			for _, l := range tt.labels {
				pr.Labels = append(pr.Labels, &github.Label{Name: github.String(l)})
			}
			gotGroup, gotSkip := classify(pr, cfg, allowed)
			if gotGroup != tt.wantGroup || gotSkip != tt.wantSkip {
				t.Fatalf("classify(%q) = (%q, %v); want (%q, %v)", tt.title, gotGroup, gotSkip, tt.wantGroup, tt.wantSkip)
			}
		})
	}
}

func TestGroupPullRequests(t *testing.T) {
	cfg := Config{
		SkipPrefixes:  []string{"ci:"},
		PrefixGroups:  map[string]string{"addon:": "addons"},
		ContainGroups: map[string]string{"fix": "bug fix"},
		AllowedLabels: []string{"kind/feature"},
	}

	prs := map[int]*github.PullRequest{
		1: {Title: github.String("addon: cni"), Number: github.Int(1)},
		2: {Title: github.String("ci: skip"), Number: github.Int(2)},
		3: {Title: github.String("feature"), Number: github.Int(3), Labels: []*github.Label{{Name: github.String("kind/feature")}}},
		4: {Title: github.String("misc"), Number: github.Int(4)},
		5: {Title: github.String("fix crash"), Number: github.Int(5)},
	}

	groups := groupPullRequests(prs, cfg)
	if len(groups) != 4 {
		t.Fatalf("expected 4 groups, got %d", len(groups))
	}
	if g := groups["addons"]; len(g) != 1 || g[0].GetNumber() != 1 {
		t.Fatalf("addons group unexpected: %+v", g)
	}
	if g := groups["kind/feature"]; len(g) != 1 || g[0].GetNumber() != 3 {
		t.Fatalf("kind/feature group unexpected: %+v", g)
	}
	if g := groups["bug fix"]; len(g) != 1 || g[0].GetNumber() != 5 {
		t.Fatalf("bug fix group unexpected: %+v", g)
	}
	if g := groups["other"]; len(g) != 1 || g[0].GetNumber() != 4 {
		t.Fatalf("other group unexpected: %+v", g)
	}
	if _, ok := groups["ci:"]; ok {
		t.Fatalf("unexpected group for skipped PR")
	}
}

func TestPrintGroups(t *testing.T) {
	groups := map[string][]*github.PullRequest{
		"z": {
			{Title: github.String("b"), Number: github.Int(2)},
			{Title: github.String("a"), Number: github.Int(1)},
		},
		"a": {
			{Title: github.String("same"), Number: github.Int(2)},
			{Title: github.String("same"), Number: github.Int(1)},
		},
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	printGroups(groups)
	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	got := string(out)
	want := "## a\n* same (#1)\n* same (#2)\n\n## z\n* a (#1)\n* b (#2)\n\n"
	if got != want {
		t.Errorf("printGroups output = %q, want %q", got, want)
	}
}
