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

package node

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/driver"
)

func TestShouldWarnVirtualBox(t *testing.T) {
	tests := []struct {
		name            string
		driverName      string
		driverSpecified bool
		wantWarning     bool
		want            bool
	}{
		{
			name:            "✅ user-specified virtualbox with warning enabled → warn",
			driverName:      driver.VirtualBox,
			driverSpecified: true,
			wantWarning:     true,
			want:            true,
		},
		{
			name:            "❌ auto-selected virtualbox → suppress (#15456)",
			driverName:      driver.VirtualBox,
			driverSpecified: false,
			wantWarning:     true,
			want:            false,
		},
		{
			name:            "❌ user-specified virtualbox with warning opted out → suppress",
			driverName:      driver.VirtualBox,
			driverSpecified: true,
			wantWarning:     false,
			want:            false,
		},
		{
			name:            "❌ auto-selected virtualbox with warning opted out → suppress",
			driverName:      driver.VirtualBox,
			driverSpecified: false,
			wantWarning:     false,
			want:            false,
		},
		{
			name:            "❌ non-virtualbox driver, user-specified → no warning",
			driverName:      driver.Docker,
			driverSpecified: true,
			wantWarning:     true,
			want:            false,
		},
		{
			name:            "❌ non-virtualbox driver, auto-selected → no warning",
			driverName:      driver.Docker,
			driverSpecified: false,
			wantWarning:     true,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldWarnVirtualBox(tt.driverName, tt.driverSpecified, tt.wantWarning)
			if got != tt.want {
				t.Errorf("shouldWarnVirtualBox(%q, specified=%v, want=%v) = %v, want %v",
					tt.driverName, tt.driverSpecified, tt.wantWarning, got, tt.want)
			}
		})
	}
}
