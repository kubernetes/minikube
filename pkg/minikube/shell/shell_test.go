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
	"os"
	"strings"
	"testing"
)

func TestGenerateUsageHint(t *testing.T) {
	var testCases = []struct {
		ec       EnvConfig
		expected string
	}{
		{EnvConfig{""}, `# foo
# eval $(bar)`},
		{EnvConfig{"powershell"}, `# foo
# & bar | Invoke-Expression`},
		{EnvConfig{"bash"}, `# foo
# eval $(bar)`},
		{EnvConfig{"powershell"}, `# foo
# & bar | Invoke-Expression`},
		{EnvConfig{"emacs"}, `;; foo
;; (with-temp-buffer (shell-command "bar" (current-buffer)) (eval-buffer))`},
		{EnvConfig{"fish"}, `# foo
# bar | source`},
		{EnvConfig{"none"}, ``},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.ec.Shell, func(t *testing.T) {
			got := strings.TrimSpace(generateUsageHint(tc.ec, "foo", "bar"))
			expected := strings.TrimSpace(tc.expected)
			if got != expected {
				t.Errorf("Expected '%v' but got '%v'", expected, got)
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
		{[]string{"baz", "bar"}, EnvConfig{""}, `unset baz;
unset bar;`},
		{[]string{"baz", "bar"}, EnvConfig{"bash"}, `unset baz;
unset bar;`},
		{[]string{"baz", "bar"}, EnvConfig{"powershell"}, `Remove-Item Env:\\baz
Remove-Item Env:\\bar`},
		{[]string{"baz", "bar"}, EnvConfig{"cmd"}, `SET baz=
SET bar=`},
		{[]string{"baz", "bar"}, EnvConfig{"fish"}, `set -e baz;
set -e bar;`},
		{[]string{"baz", "bar"}, EnvConfig{"emacs"}, `(setenv "baz" nil)
(setenv "bar" nil)`},
		{[]string{"baz", "bar"}, EnvConfig{"none"}, "baz\nbar"},
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

func TestDetectSet(t *testing.T) {
	orgShellEnv := os.Getenv("SHELL")
	defer os.Setenv("SHELL", orgShellEnv)

	os.Setenv("SHELL", "/bin/bash")
	if s, err := Detect(); err != nil {
		t.Fatalf("unexpected error: '%v' during shell detection. Returned shell: %s", err, s)
	} else if s == "" {
		t.Fatalf("Detected shell expected to be non empty string")
	}
}

func TestDetectUnset(t *testing.T) {
	orgShellEnv := os.Getenv("SHELL")
	defer os.Setenv("SHELL", orgShellEnv)

	os.Unsetenv("SHELL")
	if s, err := Detect(); err != nil {
		t.Fatalf("unexpected error: '%v' during shell detection. Returned shell: %s", err, s)
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
	if ec.Shell == "" {
		t.Fatalf("Expected no empty shell")
	}
}
