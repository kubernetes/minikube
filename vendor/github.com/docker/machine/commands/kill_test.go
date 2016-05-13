package commands

import (
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestCmdKillMissingMachineName(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{}
	api := &libmachinetest.FakeAPI{}

	err := cmdKill(commandLine, api)

	assert.Equal(t, ErrNoDefault, err)
}

func TestCmdKill(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToKill1", "machineToKill2"},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name: "machineToKill1",
				Driver: &fakedriver.Driver{
					MockState: state.Running,
				},
			},
			{
				Name: "machineToKill2",
				Driver: &fakedriver.Driver{
					MockState: state.Running,
				},
			},
			{
				Name: "machine",
				Driver: &fakedriver.Driver{
					MockState: state.Running,
				},
			},
		},
	}

	err := cmdKill(commandLine, api)
	assert.NoError(t, err)

	assert.Equal(t, state.Stopped, libmachinetest.State(api, "machineToKill1"))
	assert.Equal(t, state.Stopped, libmachinetest.State(api, "machineToKill2"))
	assert.Equal(t, state.Running, libmachinetest.State(api, "machine"))
}
