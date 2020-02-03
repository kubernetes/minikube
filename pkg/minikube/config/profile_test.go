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

// TestListProfiles uses a different MINIKUBE_HOME with rest of tests since it relies on file list index
func TestListProfiles(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/profile/.minikube")
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
		if val[tt.index].Config.VMDriver != tt.vmDriver {
			t.Errorf("expected %s got %v", tt.vmDriver, val[tt.index].Config.VMDriver)
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

func TestProfileNameInReservedKeywords(t *testing.T) {
	var testCases = []struct {
		name     string
		expected bool
	}{
		{"start", true},
		{"stop", true},
		{"status", true},
		{"delete", true},
		{"config", true},
		{"open", true},
		{"profile", true},
		{"addons", true},
		{"cache", true},
		{"logs", true},
		{"myprofile", false},
		{"log", false},
	}
	for _, tt := range testCases {
		got := ProfileNameInReservedKeywords(tt.name)
		if got != tt.expected {
			t.Errorf("expected ProfileNameInReservedKeywords(%s)=%t but got %t ", tt.name, tt.expected, got)
		}
	}
}

func TestProfileExists(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	var testCases = []struct {
		name     string
		expected bool
	}{
		{"p1", true},
		{"p2", true},
		{"p3_empty", true},
		{"p4_invalid_file", true},
		{"p5_partial_config", true},
		{"p6_no_file", false},
	}
	for _, tt := range testCases {
		got := ProfileExists(tt.name, miniDir)
		if got != tt.expected {
			t.Errorf("expected ProfileExists(%q,%q)=%t but got %t ", tt.name, miniDir, tt.expected, got)
		}

	}

}

func TestCreateEmptyProfile(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	var testCases = []struct {
		name      string
		expectErr bool
	}{
		{"p13", false},
		{"p_13", false},
	}
	for _, tc := range testCases {
		n := tc.name // capturing  loop variable
		gotErr := CreateEmptyProfile(n, miniDir)
		if gotErr != nil && tc.expectErr == false {
			t.Errorf("expected CreateEmptyProfile not to error but got err=%v", gotErr)
		}

		defer func() { // tear down
			err := DeleteProfile(n, miniDir)
			if err != nil {
				t.Errorf("error test tear down %v", err)
			}
		}()

	}

}

func TestCreateProfile(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	var testCases = []struct {
		name      string
		cfg       *MachineConfig
		expectErr bool
	}{
		{"p_empty_config", &MachineConfig{}, false},
		{"p_partial_config", &MachineConfig{KubernetesConfig: KubernetesConfig{
			ShouldLoadCachedImages: false}}, false},
		{"p_partial_config2", &MachineConfig{
			KeepContext: false, KubernetesConfig: KubernetesConfig{
				ShouldLoadCachedImages: false}}, false},
	}
	for _, tc := range testCases {
		n := tc.name // capturing  loop variable
		gotErr := SaveProfile(n, tc.cfg, miniDir)
		if gotErr != nil && tc.expectErr == false {
			t.Errorf("expected CreateEmptyProfile not to error but got err=%v", gotErr)
		}

		defer func() { // tear down

			err := DeleteProfile(n, miniDir)
			if err != nil {
				t.Errorf("error test tear down %v", err)
			}
		}()
	}

}

func TestDeleteProfile(t *testing.T) {
	miniDir, err := filepath.Abs("./testdata/.minikube2")
	if err != nil {
		t.Errorf("error getting dir path for ./testdata/.minikube : %v", err)
	}

	err = CreateEmptyProfile("existing_prof", miniDir)
	if err != nil {
		t.Errorf("error setting up TestDeleteProfile %v", err)
	}

	var testCases = []struct {
		name      string
		expectErr bool
	}{
		{"existing_prof", false},
		{"non_existing_prof", false},
	}
	for _, tc := range testCases {
		gotErr := DeleteProfile(tc.name, miniDir)
		if gotErr != nil && tc.expectErr == false {
			t.Errorf("expected CreateEmptyProfile not to error but got err=%v", gotErr)
		}
	}

}
