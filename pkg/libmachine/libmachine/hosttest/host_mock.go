package hosttest

import (
	"fmt"

	"k8s.io/minikube/pkg/libmachine/drivers/mockdriver"
)

// MockHost used for testing. When commands are run, the output from CommandOutput
// is used, if present. Then the output from Error is used, if present. Finally,
// "", nil is returned.
type MockHost struct {
	CommandOutput map[string]string
	Error         string
	Commands      map[string]int
	Driver        *mockdriver.MockDriver
}

// NewMockHost creates a new MockHost
func NewMockHost() *MockHost {
	return &MockHost{
		CommandOutput: make(map[string]string),
		Commands:      make(map[string]int),
		Driver:        &mockdriver.MockDriver{},
	}
}

// RunSSHCommand runs a SSH command, returning output
func (m MockHost) RunSSHCommand(cmd string) (string, error) {
	m.Commands[cmd] = 1
	output, ok := m.CommandOutput[cmd]
	if ok {
		return output, nil
	}
	if m.Error != "" {
		return "", fmt.Errorf(m.Error)
	}
	return "", nil
}
