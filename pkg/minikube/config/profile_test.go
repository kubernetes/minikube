/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package config

import (
	"path/filepath"
	"testing"
)

func TestListProfiles(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}
	// test cases for valid profiles
	var testCasesValidProfs = []struct {
		index      int
		expectName string
		vmDriver   string
	}{
		{0, "p1", "hyperkit"},
		{1, "p2", "virtualbox"},
	}

	// test cases for invalid profiles
	var testCasesInValidProfs = []struct {
		index      int
		expectName string
		vmDriver   string
	}{
		{0, "p3_empty", ""},
		{1, "p4_invalid_file", ""},
		{2, "p5_partial_config", ""},
	}

	val, inv, err := ListProfiles(miniDir)

	for _, tt := range testCasesValidProfs {
		if val[tt.index].Name != tt.expectName {
			t.Errorf("expected %s got %v", tt.expectName, val[tt.index].Name)
		}
		if val[tt.index].Config.MachineConfig.VMDriver != tt.vmDriver {
			t.Errorf("expected %s got %v", tt.vmDriver, val[tt.index].Config.MachineConfig.VMDriver)
		}

	}

	// making sure it returns the invalid profiles
	for _, tt := range testCasesInValidProfs {
		if inv[tt.index].Name != tt.expectName {
			t.Errorf("expected %s got %v", tt.expectName, inv[tt.index].Name)
		}
	}

	if err != nil {
		t.Errorf("error listing profiles %v", err)
	}
}
