package commands

import (
	"testing"

	"k8s.io/minikube/pkg/libmachine/commands/commandstest"
	"k8s.io/minikube/pkg/libmachine/drivers/fakedriver"
	"k8s.io/minikube/pkg/libmachine/libmachine"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/host"
	"k8s.io/minikube/pkg/libmachine/libmachine/libmachinetest"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
	"github.com/stretchr/testify/assert"
)

func TestCmdProvision(t *testing.T) {
	testCases := []struct {
		commandLine CommandLine
		api         libmachine.API
		expectedErr error
	}{
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"foo", "bar"},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name:   "foo",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
							SwarmOptions:  &swarm.Options{},
						},
					},
					{
						Name:   "bar",
						Driver: &fakedriver.Driver{},
						HostOptions: &host.Options{
							EngineOptions: &engine.Options{},
							AuthOptions:   &auth.Options{},
							SwarmOptions:  &swarm.Options{},
						},
					},
				},
			},
			expectedErr: nil,
		},
	}

	provision.SetDetector(&provision.FakeDetector{
		Provisioner: provision.NewFakeProvisioner(nil),
	})

	// fakeprovisioner always returns "true" for compatible host, so we
	// just need to register it.
	provision.Register("fakeprovisioner", &provision.RegisteredProvisioner{
		New: provision.NewFakeProvisioner,
	})

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedErr, cmdProvision(tc.commandLine, tc.api))
	}
}
