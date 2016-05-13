package commands

import (
	"os"
	"testing"

	"time"

	"errors"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcndockerclient"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/stretchr/testify/assert"
)

func TestParseFiltersErrorsGivenInvalidFilter(t *testing.T) {
	_, err := parseFilters([]string{"foo=bar"})
	assert.EqualError(t, err, "Unsupported filter key 'foo'")
}

func TestParseFiltersSwarm(t *testing.T) {
	actual, _ := parseFilters([]string{"swarm=foo"})
	assert.Equal(t, actual, FilterOptions{SwarmName: []string{"foo"}})
}

func TestParseFiltersDriver(t *testing.T) {
	actual, _ := parseFilters([]string{"driver=bar"})
	assert.Equal(t, actual, FilterOptions{DriverName: []string{"bar"}})
}

func TestParseFiltersState(t *testing.T) {
	actual, _ := parseFilters([]string{"state=Running"})
	assert.Equal(t, actual, FilterOptions{State: []string{"Running"}})
}

func TestParseFiltersName(t *testing.T) {
	actual, _ := parseFilters([]string{"name=dev"})
	assert.Equal(t, actual, FilterOptions{Name: []string{"dev"}})
}

func TestParseFiltersLabel(t *testing.T) {
	actual, err := parseFilters([]string{"label=com.example.foo=bar"})
	assert.EqualValues(t, actual, FilterOptions{Labels: []string{"com.example.foo=bar"}})
	assert.Nil(t, err, "returned err value must be Nil")
}

func TestParseFiltersAll(t *testing.T) {
	actual, _ := parseFilters([]string{"swarm=foo", "driver=bar", "state=Stopped", "name=dev"})
	assert.Equal(t, actual, FilterOptions{SwarmName: []string{"foo"}, DriverName: []string{"bar"}, State: []string{"Stopped"}, Name: []string{"dev"}})
}

func TestParseFiltersAllCase(t *testing.T) {
	actual, err := parseFilters([]string{"sWarM=foo", "DrIver=bar", "StaTe=Stopped", "NAMe=dev", "LABEL=com=foo"})
	assert.Equal(t, actual, FilterOptions{SwarmName: []string{"foo"}, DriverName: []string{"bar"}, State: []string{"Stopped"}, Name: []string{"dev"}, Labels: []string{"com=foo"}})
	assert.Nil(t, err, "err should be nil")
}

func TestParseFiltersDuplicates(t *testing.T) {
	actual, _ := parseFilters([]string{"swarm=foo", "driver=bar", "name=mark", "swarm=baz", "driver=qux", "state=Running", "state=Starting", "name=time"})
	assert.Equal(t, actual, FilterOptions{SwarmName: []string{"foo", "baz"}, DriverName: []string{"bar", "qux"}, State: []string{"Running", "Starting"}, Name: []string{"mark", "time"}})
}

func TestParseFiltersValueWithEqual(t *testing.T) {
	actual, _ := parseFilters([]string{"driver=bar=baz"})
	assert.Equal(t, actual, FilterOptions{DriverName: []string{"bar=baz"}})
}

func TestFilterHostsReturnsFiltersValuesCaseInsensitive(t *testing.T) {
	opts := FilterOptions{
		SwarmName:  []string{"fOo"},
		DriverName: []string{"ViRtUaLboX"},
		State:      []string{"StOPpeD"},
		Labels:     []string{"com.EXAMPLE.app=FOO"},
	}
	hosts := []*host.Host{}
	actual := filterHosts(hosts, opts)
	assert.EqualValues(t, actual, hosts)
}
func TestFilterHostsReturnsSameGivenNoFilters(t *testing.T) {
	opts := FilterOptions{}
	hosts := []*host.Host{
		{
			Name:       "testhost",
			DriverName: "fakedriver",
		},
	}
	actual := filterHosts(hosts, opts)
	assert.EqualValues(t, actual, hosts)
}

func TestFilterHostsReturnSetLabel(t *testing.T) {
	opts := FilterOptions{
		Labels: []string{"com.class.foo=bar"},
	}
	hosts := []*host.Host{
		{
			Name:       "testhost",
			DriverName: "fakedriver",
			HostOptions: &host.Options{
				EngineOptions: &engine.Options{
					Labels: []string{"com.class.foo=bar"},
				},
			},
		},
	}
	actual := filterHosts(hosts, opts)
	assert.EqualValues(t, actual, hosts)
}

func TestFilterHostsReturnsEmptyGivenEmptyHosts(t *testing.T) {
	opts := FilterOptions{
		SwarmName: []string{"foo"},
	}
	hosts := []*host.Host{}
	assert.Empty(t, filterHosts(hosts, opts))
}

func TestFilterHostsReturnsEmptyGivenNonMatchingFilters(t *testing.T) {
	opts := FilterOptions{
		SwarmName: []string{"foo"},
	}
	hosts := []*host.Host{
		{
			Name:       "testhost",
			DriverName: "fakedriver",
		},
	}
	assert.Empty(t, filterHosts(hosts, opts))
}

func TestFilterHostsBySwarmName(t *testing.T) {
	opts := FilterOptions{
		SwarmName: []string{"master"},
	}
	master :=
		&host.Host{
			Name: "master",
			HostOptions: &host.Options{
				SwarmOptions: &swarm.Options{Master: true, Discovery: "foo"},
			},
		}
	node1 :=
		&host.Host{
			Name: "node1",
			HostOptions: &host.Options{
				SwarmOptions: &swarm.Options{Master: false, Discovery: "foo"},
			},
		}
	othermaster :=
		&host.Host{
			Name: "othermaster",
			HostOptions: &host.Options{
				SwarmOptions: &swarm.Options{Master: true, Discovery: "bar"},
			},
		}
	hosts := []*host.Host{master, node1, othermaster}
	expected := []*host.Host{master, node1}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsByDriverName(t *testing.T) {
	opts := FilterOptions{
		DriverName: []string{"fakedriver"},
	}
	node1 :=
		&host.Host{
			Name:       "node1",
			DriverName: "fakedriver",
		}
	node2 :=
		&host.Host{
			Name:       "node2",
			DriverName: "virtualbox",
		}
	node3 :=
		&host.Host{
			Name:       "node3",
			DriverName: "fakedriver",
		}
	hosts := []*host.Host{node1, node2, node3}
	expected := []*host.Host{node1, node3}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsByState(t *testing.T) {
	opts := FilterOptions{
		State: []string{"Paused", "Saved", "Stopped"},
	}
	node1 :=
		&host.Host{
			Name:       "node1",
			DriverName: "fakedriver",
			Driver:     &fakedriver.Driver{MockState: state.Paused},
		}
	node2 :=
		&host.Host{
			Name:       "node2",
			DriverName: "virtualbox",
			Driver:     &fakedriver.Driver{MockState: state.Stopped},
		}
	node3 :=
		&host.Host{
			Name:       "node3",
			DriverName: "fakedriver",
			Driver:     &fakedriver.Driver{MockState: state.Running},
		}
	hosts := []*host.Host{node1, node2, node3}
	expected := []*host.Host{node1, node2}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsByName(t *testing.T) {
	opts := FilterOptions{
		Name: []string{"fire", "ice", "earth", "a.?r"},
	}
	node1 :=
		&host.Host{
			Name:       "fire",
			DriverName: "fakedriver",
			Driver:     &fakedriver.Driver{MockState: state.Paused, MockName: "fire"},
		}
	node2 :=
		&host.Host{
			Name:       "ice",
			DriverName: "adriver",
			Driver:     &fakedriver.Driver{MockState: state.Paused, MockName: "ice"},
		}
	node3 :=
		&host.Host{
			Name:       "air",
			DriverName: "nodriver",
			Driver:     &fakedriver.Driver{MockState: state.Paused, MockName: "air"},
		}
	node4 :=
		&host.Host{
			Name:       "water",
			DriverName: "falsedriver",
			Driver:     &fakedriver.Driver{MockState: state.Paused, MockName: "water"},
		}
	hosts := []*host.Host{node1, node2, node3, node4}
	expected := []*host.Host{node1, node2, node3}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsMultiFlags(t *testing.T) {
	opts := FilterOptions{
		SwarmName:  []string{},
		DriverName: []string{"fakedriver", "virtualbox"},
	}
	node1 :=
		&host.Host{
			Name:       "node1",
			DriverName: "fakedriver",
		}
	node2 :=
		&host.Host{
			Name:       "node2",
			DriverName: "virtualbox",
		}
	node3 :=
		&host.Host{
			Name:       "node3",
			DriverName: "softlayer",
		}
	hosts := []*host.Host{node1, node2, node3}
	expected := []*host.Host{node1, node2}

	assert.EqualValues(t, filterHosts(hosts, opts), expected)
}

func TestFilterHostsDifferentFlagsProduceAND(t *testing.T) {
	opts := FilterOptions{
		DriverName: []string{"virtualbox"},
		State:      []string{"Running"},
	}

	hosts := []*host.Host{
		{
			Name:       "node1",
			DriverName: "fakedriver",
			Driver:     &fakedriver.Driver{MockState: state.Paused},
		},
		{
			Name:       "node2",
			DriverName: "virtualbox",
			Driver:     &fakedriver.Driver{MockState: state.Stopped},
		},
		{
			Name:       "node3",
			DriverName: "fakedriver",
			Driver:     &fakedriver.Driver{MockState: state.Running},
		},
	}

	assert.Empty(t, filterHosts(hosts, opts))
}

func TestGetHostListItems(t *testing.T) {
	defer func(versioner mcndockerclient.DockerVersioner) { mcndockerclient.CurrentDockerVersioner = versioner }(mcndockerclient.CurrentDockerVersioner)
	mcndockerclient.CurrentDockerVersioner = &mcndockerclient.FakeDockerVersioner{Version: "1.9"}

	// TODO: Ideally this would mockable via interface instead.
	defer func(host string) { os.Setenv("DOCKER_HOST", host) }(os.Getenv("DOCKER_HOST"))
	os.Setenv("DOCKER_HOST", "tcp://active.host.com:2376")

	hosts := []*host.Host{
		{
			Name: "foo",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
				MockIP:    "active.host.com",
			},
		},
		{
			Name: "bar100",
			Driver: &fakedriver.Driver{
				MockState: state.Stopped,
			},
		},
		{
			Name: "bar10",
			Driver: &fakedriver.Driver{
				MockState: state.Error,
			},
		},
	}

	expected := []struct {
		name    string
		state   state.State
		active  bool
		version string
		error   string
	}{
		{"bar10", state.Error, false, "Unknown", "Unable to get ip"},
		{"bar100", state.Stopped, false, "Unknown", ""},
		{"foo", state.Running, true, "v1.9", ""},
	}

	items := getHostListItems(hosts, map[string]error{}, 10*time.Second)

	for i := range expected {
		assert.Equal(t, expected[i].name, items[i].Name)
		assert.Equal(t, expected[i].state, items[i].State)
		assert.Equal(t, expected[i].active, items[i].ActiveHost)
		assert.Equal(t, expected[i].version, items[i].DockerVersion)
		assert.Equal(t, expected[i].error, items[i].Error)
	}
}

func TestGetHostListItemsEnvDockerHostUnset(t *testing.T) {
	defer func(versioner mcndockerclient.DockerVersioner) { mcndockerclient.CurrentDockerVersioner = versioner }(mcndockerclient.CurrentDockerVersioner)
	mcndockerclient.CurrentDockerVersioner = &mcndockerclient.FakeDockerVersioner{Version: "1.9"}

	defer func(host string) { os.Setenv("DOCKER_HOST", host) }(os.Getenv("DOCKER_HOST"))
	os.Unsetenv("DOCKER_HOST")

	hosts := []*host.Host{
		{
			Name: "foo",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
				MockIP:    "120.0.0.1",
			},
		},
		{
			Name: "bar",
			Driver: &fakedriver.Driver{
				MockState: state.Stopped,
			},
		},
		{
			Name: "baz",
			Driver: &fakedriver.Driver{
				MockState: state.Saved,
			},
		},
	}

	expected := map[string]struct {
		state  state.State
		active bool
	}{
		"foo": {state.Running, false},
		"bar": {state.Stopped, false},
		"baz": {state.Saved, false},
	}

	items := getHostListItems(hosts, map[string]error{}, 10*time.Second)

	for _, item := range items {
		expected := expected[item.Name]

		assert.Equal(t, expected.state, item.State)
		assert.Equal(t, expected.active, item.ActiveHost)
	}
}

func TestIsActive(t *testing.T) {
	cases := []struct {
		dockerHost string
		state      state.State
		expected   bool
	}{
		{"", state.Running, false},
		{"tcp://5.6.7.8:2376", state.Running, false},
		{"tcp://1.2.3.4:2376", state.Stopped, false},
		{"tcp://1.2.3.4:2376", state.Running, true},
		{"tcp://1.2.3.4:3376", state.Running, false},
	}

	for _, c := range cases {
		os.Unsetenv("DOCKER_HOST")
		if c.dockerHost != "" {
			os.Setenv("DOCKER_HOST", c.dockerHost)
		}

		actual := isActive(c.state, "tcp://1.2.3.4:2376")

		assert.Equal(t, c.expected, actual)
	}
}

func TestIsSwarmActive(t *testing.T) {
	cases := []struct {
		dockerHost string
		state      state.State
		isMaster   bool
		expected   bool
	}{
		{"", state.Running, false, false},
		{"tcp://5.6.7.8:3376", state.Running, true, false},
		{"tcp://1.2.3.4:3376", state.Stopped, true, false},
		{"tcp://1.2.3.4:3376", state.Running, true, true},
		{"tcp://1.2.3.4:3376", state.Running, false, false},
		{"tcp://1.2.3.4:2376", state.Running, true, false},
	}

	for _, c := range cases {
		os.Unsetenv("DOCKER_HOST")
		if c.dockerHost != "" {
			os.Setenv("DOCKER_HOST", c.dockerHost)
		}

		actual := isSwarmActive(c.state, "tcp://1.2.3.4:2376", c.isMaster, "tcp://0.0.0.0:3376")

		assert.Equal(t, c.expected, actual)
	}
}

func TestGetHostStateTimeout(t *testing.T) {
	hosts := []*host.Host{
		{
			Name: "foo",
			Driver: &fakedriver.Driver{
				MockState: state.Timeout,
			},
		},
	}

	hostItem := getHostListItems(hosts, nil, time.Millisecond)[0]

	assert.Equal(t, "foo", hostItem.Name)
	assert.Equal(t, state.Timeout, hostItem.State)
	assert.Equal(t, "Driver", hostItem.DriverName)
	assert.Equal(t, time.Millisecond, hostItem.ResponseTime)
}

func TestGetHostStateError(t *testing.T) {
	hosts := []*host.Host{
		{
			Name: "foo",
			Driver: &fakedriver.Driver{
				MockState: state.Error,
			},
		},
	}

	hostItem := getHostListItems(hosts, nil, 10*time.Second)[0]

	assert.Equal(t, "foo", hostItem.Name)
	assert.Equal(t, state.Error, hostItem.State)
	assert.Equal(t, "Driver", hostItem.DriverName)
	assert.Empty(t, hostItem.URL)
	assert.Equal(t, "Unable to get ip", hostItem.Error)
	assert.Nil(t, hostItem.SwarmOptions)
}

func TestGetSomeHostInError(t *testing.T) {
	defer func(versioner mcndockerclient.DockerVersioner) { mcndockerclient.CurrentDockerVersioner = versioner }(mcndockerclient.CurrentDockerVersioner)
	mcndockerclient.CurrentDockerVersioner = &mcndockerclient.FakeDockerVersioner{Version: "1.9"}

	hosts := []*host.Host{
		{
			Name: "foo",
			Driver: &fakedriver.Driver{
				MockState: state.Running,
			},
		},
	}
	hostsInError := map[string]error{
		"bar": errors.New("invalid memory address or nil pointer dereference"),
	}

	hostItems := getHostListItems(hosts, hostsInError, 10*time.Second)
	assert.Equal(t, 2, len(hostItems))

	hostItem := hostItems[0]
	assert.Equal(t, "bar", hostItem.Name)
	assert.Equal(t, state.Error, hostItem.State)
	assert.Equal(t, "not found", hostItem.DriverName)
	assert.Empty(t, hostItem.URL)
	assert.Equal(t, "invalid memory address or nil pointer dereference", hostItem.Error)
	assert.Nil(t, hostItem.SwarmOptions)

	hostItem = hostItems[1]
	assert.Equal(t, "foo", hostItem.Name)
	assert.Equal(t, state.Running, hostItem.State)
}

func TestOnErrorWithMultilineComment(t *testing.T) {
	err := errors.New("MissingParameter: The request must contain the parameter InstanceId\n	status code: 400, request id:")

	itemInError := newHostListItemInError("foo", err)

	assert.Equal(t, itemInError.Error, "MissingParameter: The request must contain the parameter InstanceId 	status code: 400, request id:")
}
