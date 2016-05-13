package provision

import (
	"regexp"
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision/provisiontest"
	"github.com/docker/machine/libmachine/swarm"
)

func TestRedHatGenerateYumRepoList(t *testing.T) {
	info := &OsRelease{
		ID: "rhel",
	}
	p := NewRedHatProvisioner("rhel", nil)
	p.SetOsReleaseInfo(info)

	buf, err := generateYumRepoList(p)
	if err != nil {
		t.Fatal(err)
	}

	m, err := regexp.MatchString(".*centos/7.*", buf.String())
	if err != nil {
		t.Fatal(err)
	}

	if !m {
		t.Fatalf("expected match for centos/7")
	}
}

func TestRedHatDefaultStorageDriver(t *testing.T) {
	p := NewRedHatProvisioner("", &fakedriver.Driver{})
	p.SSHCommander = provisiontest.NewFakeSSHCommander(provisiontest.FakeSSHCommanderOptions{})
	p.Provision(swarm.Options{}, auth.Options{}, engine.Options{})
	if p.EngineOptions.StorageDriver != "devicemapper" {
		t.Fatal("Default storage driver should be devicemapper")
	}
}
