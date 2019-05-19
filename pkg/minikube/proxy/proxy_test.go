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

package proxy

import (
	"fmt"
	"os"
	"testing"
)

func TestIsValidEnv(t *testing.T) {
	var testCases = []struct {
		env  string
		want bool
	}{
		{"", false},
		{"HTTPS-PROXY", false},
		{"NOPROXY", false},
	}
	for _, tc := range testCases {
		t.Run(tc.env, func(t *testing.T) {
			got := isValidEnv(tc.env)
			if got != tc.want {
				t.Errorf("isValidEnv(\"%v\") got %v; want %v", tc.env, got, tc.want)
			}

		})
	}

}
func TestIsInBlock(t *testing.T) {

	var testCases = []struct {
		ip        string
		block     string
		want      bool
		wanntAErr bool
	}{
		{"", "192.168.0.1/32", false, true},
		{"192.168.0.1", "192.168.0.1/32", true, false},
		{"192.168.0.2", "192.168.0.1/32", false, false},
		{"192.168.0.1", "192.168.0.1/18", true, false},
		{"abcd", "192.168.0.1/18", false, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s in %s", tc.ip, tc.block), func(t *testing.T) {
			got, err := isInBlock(tc.ip, tc.block)
			gotErr := false
			if err != nil {
				gotErr = true
			}
			if gotErr != tc.wanntAErr {
				t.Errorf("isInBlock(%v,%v) got error is %v ; want error is %v", tc.ip, tc.block, gotErr, tc.wanntAErr)
			}

			if got != tc.want {
				t.Errorf("isInBlock(%v,%v) got %v; want %v", tc.ip, tc.block, got, tc.want)
			}

		})
	}
}

func TestUpdateEnv(t *testing.T) {
	var testCases = []struct {
		ip      string
		env     string
		wantErr bool
	}{
		{"192.168.0.13", "NO_PROXY", false},
		{"", "NO_PROXY", true},
		{"", "", true},
		{"192.168.0.13", "", true},
		{"192.168.0.13", "NPROXY", true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s in %s", tc.ip, tc.env), func(t *testing.T) {
			origVal := os.Getenv(tc.env)
			gotErr := false
			err := updateEnv(tc.ip, tc.env)
			if err != nil {
				gotErr = true
			}
			if gotErr != tc.wantErr {
				t.Errorf("updateEnv(%v,%v) got error is %v ; want error is %v", tc.ip, tc.env, gotErr, tc.wantErr)
			}
			err = os.Setenv(tc.env, origVal)
			if err != nil && tc.env != "" {
				t.Errorf("Error reverting the env var (%s) to its original value (%s)", tc.env, origVal)
			}

		})
	}

}

func TestCheckEnv(t *testing.T) {
	var testCases = []struct {
		ip           string
		envName      string
		want         bool
		mockEnvValue string
	}{
		{"", "NO_PROXY", false, ""},
		{"192.168.0.13", "NO_PROXY", false, ""},
		{"192.168.0.13", "NO_PROXY", false, ","},
		{"192.168.0.13", "NO_PROXY", true, "192.168.0.13"},
		{"192.168.0.13", "NO_PROXY", true, ",192.168.0.13"},
		{"192.168.0.13", "NO_PROXY", true, "10.10.0.13,192.168.0.13"},
		{"192.168.0.13", "NO_PROXY", true, "192.168.0.13/22"},
		{"192.168.0.13", "NO_PROXY", true, "10.10.0.13,192.168.0.13"},
		{"192.168.0.13", "NO_PROXY", false, "10.10.0.13/22"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s in %s", tc.ip, tc.envName), func(t *testing.T) {
			originalEnv := os.Getenv(tc.envName)
			defer func() { // revert to pre-test env var
				err := os.Setenv(tc.envName, originalEnv)
				if err != nil {
					t.Fatalf("Error reverting env (%s) to its original value (%s) var after test ", tc.envName, originalEnv)
				}
			}()

			// defer os.Setenv(tc.envName, originalEnv)
			err := os.Setenv(tc.envName, tc.mockEnvValue) // setting up the test case
			if err != nil {
				t.Error("Error setting env var for taste case")
			}
			got := checkEnv(tc.ip, tc.envName)
			if got != tc.want {
				t.Errorf("CheckEnv(%v,%v) got  %v ; want is %v", tc.ip, tc.envName, got, tc.want)
			}

		})
	}

}
