package cruntime

import (
	"fmt"
)

// listCRIContainers returns a list of containers using crictl
func listCRIContainers(_ CommandRunner, _ string) ([]string, error) {
	// Should use crictl ps -a, but needs some massaging and testing.
	return []string{}, fmt.Errorf("unimplemented")
}

// criCRIContainers kills a list of containers using crictl
func killCRIContainers(CommandRunner, []string) error {
	return fmt.Errorf("unimplemented")
}

// StopCRIContainers stops containers using crictl
func stopCRIContainers(CommandRunner, []string) error {
	return fmt.Errorf("unimplemented")
}
