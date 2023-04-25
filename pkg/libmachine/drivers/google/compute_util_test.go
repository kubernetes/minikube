package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
	raw "google.golang.org/api/compute/v1"
)

func TestDefaultTag(t *testing.T) {
	tags := parseTags(&Driver{Tags: ""})

	assert.Equal(t, []string{"docker-machine"}, tags)
}

func TestAdditionalTag(t *testing.T) {
	tags := parseTags(&Driver{Tags: "tag1"})

	assert.Equal(t, []string{"docker-machine", "tag1"}, tags)
}

func TestAdditionalTags(t *testing.T) {
	tags := parseTags(&Driver{Tags: "tag1,tag2"})

	assert.Equal(t, []string{"docker-machine", "tag1", "tag2"}, tags)
}

func TestPortsUsed(t *testing.T) {
	var tests = []struct {
		description   string
		computeUtil   *ComputeUtil
		expectedPorts []string
		expectedError error
	}{
		{"use docker port", &ComputeUtil{}, []string{"2376/tcp"}, nil},
		{"use docker and swarm port", &ComputeUtil{SwarmMaster: true, SwarmHost: "tcp://host:3376"}, []string{"2376/tcp", "3376/tcp"}, nil},
		{"use docker and non default swarm port", &ComputeUtil{SwarmMaster: true, SwarmHost: "tcp://host:4242"}, []string{"2376/tcp", "4242/tcp"}, nil},
		{"include additional ports", &ComputeUtil{openPorts: []string{"80", "2377/udp"}}, []string{"2376/tcp", "80/tcp", "2377/udp"}, nil},
	}

	for _, test := range tests {
		ports, err := test.computeUtil.portsUsed()

		assert.Equal(t, test.expectedPorts, ports)
		assert.Equal(t, test.expectedError, err)
	}
}

func TestMissingOpenedPorts(t *testing.T) {
	var tests = []struct {
		description     string
		allowed         []*raw.FirewallAllowed
		ports           []string
		expectedMissing map[string][]string
	}{
		{"no port opened", []*raw.FirewallAllowed{}, []string{"2376"}, map[string][]string{"tcp": {"2376"}}},
		{"docker port opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376"}}}, []string{"2376"}, map[string][]string{}},
		{"missing swarm port", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376"}}}, []string{"2376", "3376"}, map[string][]string{"tcp": {"3376"}}},
		{"missing docker port", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"3376"}}}, []string{"2376", "3376"}, map[string][]string{"tcp": {"2376"}}},
		{"both ports opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "3376"}}}, []string{"2376", "3376"}, map[string][]string{}},
		{"more ports opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "3376", "22", "1024-2048"}}}, []string{"2376", "3376"}, map[string][]string{}},
		{"additional missing", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "2377/tcp"}}}, []string{"2377/udp", "80/tcp", "2376"}, map[string][]string{"tcp": {"80"}, "udp": {"2377"}}},
	}

	for _, test := range tests {
		firewall := &raw.Firewall{Allowed: test.allowed}

		missingPorts := missingOpenedPorts(firewall, test.ports)

		assert.Equal(t, test.expectedMissing, missingPorts, test.description)
	}
}
