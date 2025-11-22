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
