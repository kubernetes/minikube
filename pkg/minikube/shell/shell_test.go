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

package shell

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateUsageHint(t *testing.T) {
	var testCases = []struct {
		shellType, hintContains string
	}{
		{"", "eval"},
		{"powershell", "Invoke-Expression"},
		{"bash", "eval"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.shellType, func(t *testing.T) {
			hint := generateUsageHint(tc.shellType, "foo", "bar")
			if !strings.Contains(hint, tc.hintContains) {
				t.Errorf("Hint doesn't contain expected string. Expected to find '%v' in '%v'", tc.hintContains, hint)
			}
		})
	}
}

func TestCfgSet(t *testing.T) {
	var testCases = []struct {
		plz, cmd string
		ec       EnvConfig
		expected string
	}{
		{"", "eval", EnvConfig{""}, `"`},
		{"", "eval", EnvConfig{"bash"}, `"`},
		{"", "eval", EnvConfig{"powershell"}, `"`},
		{"", "eval", EnvConfig{"cmd"}, ``},
		{"", "eval", EnvConfig{"emacs"}, `")`},
		{"", "eval", EnvConfig{"none"}, ``},
		{"", "eval", EnvConfig{"fish"}, `";`},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ec.Shell, func(t *testing.T) {
			conf := CfgSet(tc.ec, tc.plz, tc.cmd)
			expected := strings.TrimSpace(tc.expected)
			got := strings.TrimSpace(conf.Suffix)
			if expected != got {
				t.Errorf("Expected suffix '%v' but got '%v'", expected, got)
			}
		})
	}
}

func TestUnsetScript(t *testing.T) {
	var testCases = []struct {
		vars     []string
		ec       EnvConfig
		expected string
	}{
		{[]string{"baz"}, EnvConfig{""}, `unset baz`},
		{[]string{"baz"}, EnvConfig{"bash"}, `unset baz`},
		{[]string{"baz"}, EnvConfig{"powershell"}, `Remove-Item Env:\\baz`},
		{[]string{"baz"}, EnvConfig{"cmd"}, `SET baz=`},
		{[]string{"baz"}, EnvConfig{"fish"}, `set -e baz;`},
		{[]string{"baz"}, EnvConfig{"emacs"}, `(setenv "baz" nil)`},
		{[]string{"baz"}, EnvConfig{"none"}, `baz`},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ec.Shell, func(t *testing.T) {
			var b bytes.Buffer

			if err := UnsetScript(tc.ec, &b, tc.vars); err != nil {
				t.Fatalf("Unexpected error when unseting script happen: %v", err)
			} else {
				writtenMessage := strings.TrimSpace(b.String())
				expected := strings.TrimSpace(tc.expected)
				if writtenMessage != expected {
					t.Fatalf("Expected '%v' but got '%v' ", tc.expected, writtenMessage)
				}
			}
		})
	}
}

func TestDetect(t *testing.T) {
	if s, err := Detect(); err != nil {
		t.Fatalf("unexpected error: '%v' during shell detection", err)
	} else if s == "" {
		t.Fatalf("Detected shell expected to be non empty string")
	}
}

func TestSetScript(t *testing.T) {
	ec := EnvConfig{"bash"}
	var w bytes.Buffer
	if err := SetScript(ec, &w, "foo", nil); err != nil {
		t.Fatalf("Unexpected error: '%v' during Setting script", err)
	}
	if w.String() != "foo" {
		t.Fatalf("Expected foo writed by SetScript, but got '%v'", w.String())
	}
}
