package provision

import (
	"testing"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/provisiontest"
	"github.com/docker/machine/libmachine/swarm"
)

func TestRedHatDefaultStorageDriver(t *testing.T) {
	p := NewRedHatProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	_ = p.Provision(swarm.Options{}, auth.Options{}, engine.Options{})
	if p.EngineOptions.StorageDriver != "overlay2" {
		t.Fatal("Default storage driver should be overlay2")
	}
}
