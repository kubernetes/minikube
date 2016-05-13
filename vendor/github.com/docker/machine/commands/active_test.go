package commands

import (
	"testing"

	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestCmdActiveNone(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:        "host1",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Running,
		},
		{
			Name:        "host2",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Running,
		},
		{
			Name:        "host3",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Running,
		},
	}
	_, err := activeHost(hostListItems)
	assert.Equal(t, err, errNoActiveHost)
}

func TestCmdActiveHost(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:        "host1",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Timeout,
		},
		{
			Name:        "host2",
			ActiveHost:  true,
			ActiveSwarm: false,
			State:       state.Running,
		},
		{
			Name:        "host3",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Running,
		},
	}
	active, err := activeHost(hostListItems)
	assert.Equal(t, err, nil)
	assert.Equal(t, active.Name, "host2")
}

func TestCmdActiveSwarm(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:        "host1",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Running,
		},
		{
			Name:        "host2",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Running,
		},
		{
			Name:        "host3",
			ActiveHost:  false,
			ActiveSwarm: true,
			State:       state.Running,
		},
	}
	active, err := activeHost(hostListItems)
	assert.Equal(t, err, nil)
	assert.Equal(t, active.Name, "host3")
}

func TestCmdActiveTimeout(t *testing.T) {
	hostListItems := []HostListItem{
		{
			Name:        "host1",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Running,
		},
		{
			Name:        "host2",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Running,
		},
		{
			Name:        "host3",
			ActiveHost:  false,
			ActiveSwarm: false,
			State:       state.Timeout,
		},
	}
	_, err := activeHost(hostListItems)
	assert.Equal(t, err, errActiveTimeout)
}
