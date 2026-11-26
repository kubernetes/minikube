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

package exit

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/reason"
)

// TestExitErrorSubprocess verifies end to end that Error keeps the caller's
// reason, message and exit code even when the error text matches a known
// issue (#24122). Error calls os.Exit, so the actual calls run in a child
// process re-executing this test binary.
func TestExitErrorSubprocess(t *testing.T) {
	if c := os.Getenv("MINIKUBE_EXIT_ERROR_CASE"); c != "" {
		// Child process: Error must exit, never return.
		// The error text matches the INET_PROXY_CONFUSION known issue.
		err := errors.New("http: server gave HTTP response to HTTPS client")
		switch c {
		case "generic-code":
			// The issue's own example: a generic-code kind must survive.
			Error(reason.InternalJSONMarshal, "json encoding failure", err)
		case "generic-kind":
			// The new generic fallback kind must survive as well.
			Error(reason.InternalError, "version failure", err)
		case "specific-code":
			// A specific exit code must survive too.
			Error(reason.RsrcInsufficientCores, "not enough cores", err)
		}
		t.Fatalf("Error did not exit for case %q", c)
	}

	tests := []struct {
		name      string
		envCase   string
		wantCode  int
		wantIn    []string
		wantNotIn []string
	}{
		{
			name:      "generic-code kind keeps caller reason",
			envCase:   "generic-code",
			wantCode:  reason.ExProgramError,
			wantIn:    []string{"MK_JSON_MARSHAL", "json encoding failure"},
			wantNotIn: []string{"INET_PROXY_CONFUSION"},
		},
		{
			name:      "generic kind keeps caller reason",
			envCase:   "generic-kind",
			wantCode:  reason.ExFailure,
			wantIn:    []string{"MK_INTERNAL_ERROR", "version failure"},
			wantNotIn: []string{"INET_PROXY_CONFUSION"},
		},
		{
			name:      "specific-code kind keeps caller reason",
			envCase:   "specific-code",
			wantCode:  29,
			wantIn:    []string{"RSRC_INSUFFICIENT_CORES", "not enough cores"},
			wantNotIn: []string{"INET_PROXY_CONFUSION"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run", "^TestExitErrorSubprocess$", "-test.count=1")
			cmd.Env = append(os.Environ(), "MINIKUBE_EXIT_ERROR_CASE="+tc.envCase)
			out, err := cmd.CombinedOutput()
			exitErr, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("expected process to exit with a code, got err=%v output:\n%s", err, out)
			}
			if got := exitErr.ExitCode(); got != tc.wantCode {
				t.Errorf("exit code = %d, want %d, output:\n%s", got, tc.wantCode, out)
			}
			for _, s := range tc.wantIn {
				if !strings.Contains(string(out), s) {
					t.Errorf("output missing %q:\n%s", s, out)
				}
			}
			for _, s := range tc.wantNotIn {
				if strings.Contains(string(out), s) {
					t.Errorf("output unexpectedly contains %q:\n%s", s, out)
				}
			}
		})
	}
}
