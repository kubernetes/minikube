package commands

import (
	"testing"

	"k8s.io/minikube/pkg/libmachine/commands/commandstest"
	"k8s.io/minikube/pkg/libmachine/drivers/fakedriver"
	"k8s.io/minikube/pkg/libmachine/libmachine"
	"k8s.io/minikube/pkg/libmachine/libmachine/host"
	"k8s.io/minikube/pkg/libmachine/libmachine/libmachinetest"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestCmdIPMissingMachineName(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{}
	api := &libmachinetest.FakeAPI{}

	err := cmdURL(commandLine, api)

	assert.Equal(t, err, ErrNoDefault)
}

func TestCmdIP(t *testing.T) {
	testCases := []struct {
		commandLine CommandLine
		api         libmachine.API
		expectedErr error
		expectedOut string
	}{
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{"machine"},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "machine",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
							MockIP:    "1.2.3.4",
						},
					},
				},
			},
			expectedErr: nil,
			expectedOut: "1.2.3.4\n",
		},
		{
			commandLine: &commandstest.FakeCommandLine{
				CliArgs: []string{},
			},
			api: &libmachinetest.FakeAPI{
				Hosts: []*host.Host{
					{
						Name: "default",
						Driver: &fakedriver.Driver{
							MockState: state.Running,
							MockIP:    "1.2.3.4",
						},
					},
				},
			},
			expectedErr: nil,
			expectedOut: "1.2.3.4\n",
		},
	}

	for _, tc := range testCases {
		stdoutGetter := commandstest.NewStdoutGetter()

		err := cmdIP(tc.commandLine, tc.api)

		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedOut, stdoutGetter.Output())

		stdoutGetter.Stop()
	}
}
