/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"testing"

	"github.com/pkg/errors"
)

func TestValidateOS(t *testing.T) {
	tests := []struct {
		osType   string
		errorMsg string
	}{
		{"linux", ""},
		{"windows", ""},
		{"foo", "Invalid OS: foo. Valid OS are: linux, windows"},
	}
	for _, test := range tests {
		t.Run(test.osType, func(t *testing.T) {
			got := validateOS(test.osType)
			gotError := ""
			if got != nil {
				gotError = got.Error()
			}
			if gotError != test.errorMsg {
				t.Errorf("validateOS(osType=%v): got %v, expected %v", test.osType, got, test.errorMsg)
			}
		})
	}
}

// TestParseOSFlag is the main test function for parseOSFlag
func TestParseOSFlag(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedOS  string
		expectedVer string
		expectedErr error
	}{
		{
			name:        "Valid input with all fields",
			input:       "os=windows,version=2019",
			expectedOS:  "windows",
			expectedVer: "2019",
			expectedErr: nil,
		},
		{
			name:        "Valid input with default version for windows",
			input:       "os=windows",
			expectedOS:  "windows",
			expectedVer: "2022",
			expectedErr: nil,
		},
		{
			name:        "Valid input with linux and no version",
			input:       "os=linux",
			expectedOS:  "linux",
			expectedVer: "",
			expectedErr: nil,
		},
		{
			name:        "Invalid input with missing version",
			input:       "os=linux,version=",
			expectedOS:  "linux",
			expectedVer: "",
			expectedErr: nil,
		},
		{
			name:        "Invalid input with extra comma",
			input:       "os=linux,version=,",
			expectedOS:  "",
			expectedVer: "",
			expectedErr: errors.New("Invalid format for --os flag: os=linux,version=,"),
		},
		{
			name:        "Invalid input with no key-value pair",
			input:       "linux,version=2022",
			expectedOS:  "",
			expectedVer: "",
			expectedErr: errors.New("Invalid format for --os flag: linux,version=2022"),
		},
		{
			name:        "Valid input with extra spaces",
			input:       "os=linux , version=latest",
			expectedOS:  "linux",
			expectedVer: "",
			expectedErr: nil,
		},
		{
			name:        "Valid input with capital letters in keys",
			input:       "OS=linux,Version=2022",
			expectedOS:  "linux",
			expectedVer: "",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOS, gotVer, err := parseOSFlag(tt.input)

			if tt.expectedErr != nil && (err == nil || err.Error() != tt.expectedErr.Error()) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			} else if tt.expectedErr == nil && err != nil {
				t.Errorf("Expected no error, but got %v", err)
			}

			if gotOS != tt.expectedOS {
				t.Errorf("Expected OS %s, got %s", tt.expectedOS, gotOS)
			}

			if gotVer != tt.expectedVer {
				t.Errorf("Expected version %s, got %s", tt.expectedVer, gotVer)
			}
		})
	}
}
