/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package virtualbox

import (
	"testing"

	"errors"

	"github.com/stretchr/testify/assert"
)

type MockLogsReader struct {
	content []string
	err     error
}

func (r *MockLogsReader) Read(path string) ([]string, error) {
	return r.content, r.err
}

func TestIsVTXEnabledInTheVM(t *testing.T) {
	driver := NewDriver("default", "path")

	var tests = []struct {
		description string
		content     []string
		err         error
	}{
		{"Empty log", []string{}, nil},
		{"Raw mode", []string{"Falling back to raw-mode: VT-x is disabled in the BIOS for all CPU modes"}, nil},
		{"Raw mode", []string{"HM: HMR3Init: Falling back to raw-mode: VT-x is not available"}, nil},
	}

	for _, test := range tests {
		driver.logsReader = &MockLogsReader{
			content: test.content,
			err:     test.err,
		}

		disabled, err := driver.IsVTXDisabledInTheVM()

		assert.False(t, disabled, test.description)
		assert.Equal(t, test.err, err)
	}
}

func TestIsVTXDisabledInTheVM(t *testing.T) {
	driver := NewDriver("default", "path")

	var tests = []struct {
		description string
		content     []string
		err         error
	}{
		{"VT-x Disabled", []string{"VT-x is disabled"}, nil},
		{"No HW virtualization", []string{"the host CPU does NOT support HW virtualization"}, nil},
		{"Unable to start VM", []string{"VERR_VMX_UNABLE_TO_START_VM"}, nil},
		{"Power up failed", []string{"00:00:00.318604 Power up failed (vrc=VERR_VMX_NO_VMX, rc=NS_ERROR_FAILURE (0X80004005))"}, nil},
		{"Unable to read log", nil, errors.New("Unable to read log")},
	}

	for _, test := range tests {
		driver.logsReader = &MockLogsReader{
			content: test.content,
			err:     test.err,
		}

		disabled, err := driver.IsVTXDisabledInTheVM()

		assert.True(t, disabled, test.description)
		assert.Equal(t, test.err, err)
	}
}
