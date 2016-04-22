package cluster

import (
	"testing"

	"github.com/docker/machine/libmachine/state"
	"github.com/kubernetes/minikube/cli/machine"
)

func TestCreateHost(t *testing.T) {
	api := &machine.MockAPI{}

	exists, _ := api.Exists(machineName)
	if exists {
		t.Fatal("Machine already exists.")
	}
	_, err := createHost(api)
	if err != nil {
		t.Fatalf("Error creating host: %v", err)
	}
	exists, _ = api.Exists(machineName)
	if !exists {
		t.Fatal("Machine does not exist, but should.")
	}

	h, err := api.Load(machineName)
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
