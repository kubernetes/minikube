package commands

import (
	"testing"

	"errors"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/stretchr/testify/assert"
)

func TestCmdRmMissingMachineName(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{}
	api := &libmachinetest.FakeAPI{}

	err := cmdRm(commandLine, api)

	assert.Equal(t, ErrNoMachineSpecified, err)
	assert.True(t, commandLine.HelpShown)
}

func TestCmdRm(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1", "machineToRemove2"},
		LocalFlags: &commandstest.FakeFlagger{
			Data: map[string]interface{}{
				"y": true,
			},
		},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name:   "machineToRemove1",
				Driver: &fakedriver.Driver{},
			},
			{
				Name:   "machineToRemove2",
				Driver: &fakedriver.Driver{},
			},
			{
				Name:   "machine",
				Driver: &fakedriver.Driver{},
			},
		},
	}

	err := cmdRm(commandLine, api)
	assert.NoError(t, err)

	assert.False(t, libmachinetest.Exists(api, "machineToRemove1"))
	assert.False(t, libmachinetest.Exists(api, "machineToRemove2"))
	assert.True(t, libmachinetest.Exists(api, "machine"))
}

func TestCmdRmforcefully(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1", "machineToRemove2"},
		LocalFlags: &commandstest.FakeFlagger{
			Data: map[string]interface{}{
				"force": true,
			},
		},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name:   "machineToRemove1",
				Driver: &fakedriver.Driver{},
			},
			{
				Name:   "machineToRemove2",
				Driver: &fakedriver.Driver{},
			},
		},
	}

	err := cmdRm(commandLine, api)
	assert.NoError(t, err)

	assert.False(t, libmachinetest.Exists(api, "machineToRemove1"))
	assert.False(t, libmachinetest.Exists(api, "machineToRemove2"))
}

func TestCmdRmforceDoesAutoConfirm(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1", "machineToRemove2"},
		LocalFlags: &commandstest.FakeFlagger{
			Data: map[string]interface{}{
				"y":     false,
				"force": true,
			},
		},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name:   "machineToRemove1",
				Driver: &fakedriver.Driver{},
			},
			{
				Name:   "machineToRemove2",
				Driver: &fakedriver.Driver{},
			},
		},
	}

	err := cmdRm(commandLine, api)
	assert.NoError(t, err)

	assert.False(t, libmachinetest.Exists(api, "machineToRemove1"))
	assert.False(t, libmachinetest.Exists(api, "machineToRemove2"))
}

func TestCmdRmforceConfirmUnset(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1"},
		LocalFlags: &commandstest.FakeFlagger{
			Data: map[string]interface{}{
				"y":     false,
				"force": false,
			},
		},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name:   "machineToRemove1",
				Driver: &fakedriver.Driver{},
			},
		},
	}

	err := cmdRm(commandLine, api)
	assert.NoError(t, err)

	assert.True(t, libmachinetest.Exists(api, "machineToRemove1"))
}

type DriverWithRemoveWhichFail struct {
	fakedriver.Driver
}

func (d *DriverWithRemoveWhichFail) Remove() error {
	return errors.New("unknown error")
}

func TestDontStopWhenADriverRemovalFails(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1", "machineToRemove2", "machineToRemove3"},
		LocalFlags: &commandstest.FakeFlagger{
			Data: map[string]interface{}{
				"y": true,
			},
		},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name:   "machineToRemove1",
				Driver: &fakedriver.Driver{},
			},
			{
				Name:   "machineToRemove2",
				Driver: &DriverWithRemoveWhichFail{},
			},
			{
				Name:   "machineToRemove3",
				Driver: &fakedriver.Driver{},
			},
		},
	}

	err := cmdRm(commandLine, api)
	assert.EqualError(t, err, "Error removing host \"machineToRemove2\": unknown error")

	assert.False(t, libmachinetest.Exists(api, "machineToRemove1"))
	assert.True(t, libmachinetest.Exists(api, "machineToRemove2"))
	assert.False(t, libmachinetest.Exists(api, "machineToRemove3"))
}

func TestForceRemoveEvenWhenItFails(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1"},
		LocalFlags: &commandstest.FakeFlagger{
			Data: map[string]interface{}{
				"y":     true,
				"force": true,
			},
		},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name:   "machineToRemove1",
				Driver: &DriverWithRemoveWhichFail{},
			},
		},
	}

	err := cmdRm(commandLine, api)
	assert.NoError(t, err)

	assert.False(t, libmachinetest.Exists(api, "machineToRemove1"))
}

func TestDontRemoveMachineIsRemovalFailsAndNotForced(t *testing.T) {
	commandLine := &commandstest.FakeCommandLine{
		CliArgs: []string{"machineToRemove1"},
		LocalFlags: &commandstest.FakeFlagger{
			Data: map[string]interface{}{
				"y":     true,
				"force": false,
			},
		},
	}
	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{
			{
				Name:   "machineToRemove1",
				Driver: &DriverWithRemoveWhichFail{},
			},
		},
	}

	err := cmdRm(commandLine, api)
	assert.EqualError(t, err, "Error removing host \"machineToRemove1\": unknown error")

	assert.True(t, libmachinetest.Exists(api, "machineToRemove1"))
}
