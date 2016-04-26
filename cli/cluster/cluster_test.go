package cluster

import (
	"fmt"
	"testing"

	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/kubernetes/minikube/cli/constants"
	"github.com/kubernetes/minikube/cli/tests"
)

func TestCreateHost(t *testing.T) {
	api := &tests.MockAPI{}

	exists, _ := api.Exists(constants.MachineName)
	if exists {
		t.Fatal("Machine already exists.")
	}
	_, err := createHost(api)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	exists, _ = api.Exists(constants.MachineName)
	if !exists {
		t.Fatal("Machine does not exist, but should.")
	}

	h, err := api.Load(constants.MachineName)
	if err != nil {
		t.Fatalf("Error loading machine: %v", err)
	}

	if s, _ := h.Driver.GetState(); s != state.Running {
		t.Fatalf("Machine is not running. State is: %s", s)
	}

	if h.DriverName != "virtualbox" {
		t.Fatalf("Wrong driver name: %v. Should be virtualbox.", h.DriverName)
	}
}

type mockHost struct {
	Commands []string
}

func (m mockHost) RunSSHCommand(cmd string) (string, error) {
	m.Commands = append(m.Commands, cmd)
	return "", nil
}

func TestStartCluster(t *testing.T) {
	h := mockHost{}
	err := StartCluster(h)
	if err != nil {
		t.Fatalf("Error starting cluster: %s", err)
	}
}

type mockHostError struct{}

func (m mockHostError) RunSSHCommand(cmd string) (string, error) {
	return "", fmt.Errorf("Error calling command: %s", cmd)
}

func TestStartClusterError(t *testing.T) {
	h := mockHostError{}
	err := StartCluster(h)
	if err == nil {
		t.Fatal("Error not thrown starting cluster.")
	}
}

func TestStartHostExists(t *testing.T) {
	api := &tests.MockAPI{}
	// Create an initial host.
	_, err := createHost(api)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}

	// Make sure the next call to Create will fail, to assert it doesn't get called again.
	api.CreateError = true
	if err := api.Create(&host.Host{}); err == nil {
		t.Fatal("api.Create did not fail, but should have.")
	}

	// This should pass without calling Create because the host exists already.
	h, err := StartHost(api)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != constants.MachineName {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}
}

func TestStartHost(t *testing.T) {
	api := &tests.MockAPI{}

	h, err := StartHost(api)
	if err != nil {
		t.Fatal("Error starting host.")
	}
	if h.Name != constants.MachineName {
		t.Fatalf("Machine created with incorrect name: %s", h.Name)
	}
	if exists, _ := api.Exists(h.Name); !exists {
		t.Fatal("Machine not saved.")
	}
}
