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

package kubeadm

import (
	"testing"

	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/minikube/command"
)

func TestGetAPIServerStatus(t *testing.T) {
	tests := []struct {
		name          string
		cmdToOutput   map[string]string
		expectedState string
	}{
		{
			name:          "stopped when no pid found",
			cmdToOutput:   map[string]string{},
			expectedState: state.Stopped.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcr := command.NewFakeCommandRunner()
			fcr.SetCommandToOutput(tc.cmdToOutput)

			b := &Bootstrapper{c: fcr, contextName: "minikube"}
			st, err := b.GetAPIServerStatus("127.0.0.1", 8443)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if st != tc.expectedState {
				t.Errorf("expected status %q, got %q", tc.expectedState, st)
			}
		})
	}
}
