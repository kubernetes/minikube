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

// Package sysinit provides an abstraction over init systems like systemctl
package sysinit

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/command"
)

func TestEnable(t *testing.T) {
	tests := []struct {
		service   string
		shouldErr bool
	}{
		{
			service: "docker",
		}, {
			service:   "kubelet",
			shouldErr: true,
		},
	}
	cr := command.NewFakeCommandRunner()
	cr.SetCommandToOutput(map[string]string{
		"sudo systemctl enable docker": "",
	})
	sd := &Systemd{
		r: cr,
	}
	for _, test := range tests {
		t.Run(test.service, func(t *testing.T) {
			err := sd.Enable(test.service)
			if err == nil && test.shouldErr {
				t.Fatalf("expected %s service to error, but it did not", test.service)
			}
			if err != nil && !test.shouldErr {
				t.Fatalf("expected %s service to pass, but it did not", test.service)
			}
		})
	}
}
