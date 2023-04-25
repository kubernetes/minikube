package virtualbox

import (
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	stdOutOneHostOnlyNetwork = `
Name:            vboxnet0
GUID:            786f6276-656e-4074-8000-0a0027000000
DHCP:            Disabled
IPAddress:       192.168.99.1
NetworkMask:     255.255.255.0
IPV6Address:
IPV6NetworkMaskPrefixLength: 0
HardwareAddress: 0a:00:27:00:00:00
MediumType:      Ethernet
Status:          Up
VBoxNetworkName: HostInterfaceNetworking-vboxnet0

`
	stdOutTwoHostOnlyNetwork = `
Name:            vboxnet0
GUID:            786f6276-656e-4074-8000-0a0027000000
DHCP:            Disabled
IPAddress:       192.168.99.1
NetworkMask:     255.255.255.0
IPV6Address:
IPV6NetworkMaskPrefixLength: 0
HardwareAddress: 0a:00:27:00:00:00
MediumType:      Ethernet
Status:          Up
VBoxNetworkName: HostInterfaceNetworking-vboxnet0

Name:            vboxnet1
GUID:            786f6276-656e-4174-8000-0a0027000001
DHCP:            Disabled
IPAddress:       169.254.37.187
NetworkMask:     255.255.255.0
IPV6Address:
IPV6NetworkMaskPrefixLength: 0
HardwareAddress: 0a:00:27:00:00:01
MediumType:      Ethernet
Status:          Up
VBoxNetworkName: HostInterfaceNetworking-vboxnet1
`
	stdOutListTwoDHCPServers = `
NetworkName:    HostInterfaceNetworking-vboxnet0
IP:             192.168.99.6
NetworkMask:    255.255.255.0
lowerIPAddress: 192.168.99.100
upperIPAddress: 192.168.99.254
Enabled:        Yes

NetworkName:    HostInterfaceNetworking-vboxnet1
IP:             192.168.99.7
NetworkMask:    255.255.255.0
lowerIPAddress: 192.168.99.100
upperIPAddress: 192.168.99.254
Enabled:        No
`
)

type mockHostInterfaces struct {
	mockIfaces []net.Interface
	mockAddrs  map[string]net.Addr
}

func newMockHostInterfaces() *mockHostInterfaces {
	return &mockHostInterfaces{
		mockAddrs: make(map[string]net.Addr),
	}
}

func (mhi *mockHostInterfaces) Interfaces() ([]net.Interface, error) {
	return mhi.mockIfaces, nil
}

func (mhi *mockHostInterfaces) Addrs(iface *net.Interface) ([]net.Addr, error) {
	return []net.Addr{mhi.mockAddrs[iface.Name]}, nil
}

func (mhi *mockHostInterfaces) addMockIface(ip string, mask int, iplen int, name string, flags net.Flags) (*net.IPNet, error) {
	iface := &net.Interface{Name: name, Flags: flags}
	mhi.mockIfaces = append(mhi.mockIfaces, *iface)

	ipnet := &net.IPNet{IP: net.ParseIP(ip), Mask: net.CIDRMask(mask, 8*iplen)}
	if ipnet.IP == nil {
		return nil, &net.ParseError{Type: "IP address", Text: ip}
	}
	mhi.mockAddrs[name] = ipnet
	return ipnet, nil
}

// Tests that when we have a host only network which matches our expectations,
// it gets returned correctly.
func TestGetHostOnlyNetworkHappy(t *testing.T) {
	cidr := "192.168.99.0/24"
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("Error parsing cidr: %s", err)
	}
	expectedHostOnlyNetwork := &hostOnlyNetwork{
		IPv4: *ipnet,
	}
	vboxNets := map[string]*hostOnlyNetwork{
		"HostInterfaceNetworking-vboxnet0": expectedHostOnlyNetwork,
	}

	n := getHostOnlyAdapter(vboxNets, ip, ipnet.Mask)
	if !reflect.DeepEqual(n, expectedHostOnlyNetwork) {
		t.Fatalf("Expected result of calling getHostOnlyNetwork to be the same as expected but it was not:\nexpected: %+v\nactual: %+v\n", expectedHostOnlyNetwork, n)
	}
}

// Tests that we are able to properly detect when a host only network which
// matches our expectations can not be found.
func TestGetHostOnlyNetworkNotFound(t *testing.T) {
	// Note that this has a different ip is different from "ip" below.
	cidr := "192.168.99.0/24"
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("Error parsing cidr: %s", err)
	}

	ip = net.ParseIP("192.168.59.0").To4()

	// Suppose a vbox net is created, but it doesn't align with our
	// expectation.
	vboxNet := &hostOnlyNetwork{
		IPv4: *ipnet,
	}
	vboxNets := map[string]*hostOnlyNetwork{
		"HostInterfaceNetworking-vboxnet0": vboxNet,
	}

	n := getHostOnlyAdapter(vboxNets, ip, ipnet.Mask)
	if n != nil {
		t.Fatalf("Expected vbox net to be nil but it has a value: %+v\n", n)
	}
}

// Tests a special case where Virtualbox creates the host only network
// successfully but mis-reports the netmask.
func TestGetHostOnlyNetworkWindows10Bug(t *testing.T) {
	cidr := "192.168.99.0/24"
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("Error parsing cidr: %s", err)
	}

	// This is a faulty netmask: a VirtualBox bug causes it to be
	// misreported.
	ipnet.Mask = net.IPMask(net.ParseIP("15.0.0.0").To4())

	expectedHostOnlyNetwork := &hostOnlyNetwork{
		IPv4: *ipnet,
	}

	vboxNets := map[string]*hostOnlyNetwork{
		"HostInterfaceNetworking-vboxnet0": expectedHostOnlyNetwork,
	}

	// The Mask that we are passing in will be the "legitimate" mask, so it
	// must differ from the magic buggy mask.
	n := getHostOnlyAdapter(vboxNets, ip, net.IPMask(net.ParseIP("255.255.255.0").To4()))
	if !reflect.DeepEqual(n, expectedHostOnlyNetwork) {
		t.Fatalf("Expected result of calling getHostOnlyNetwork to be the same as expected but it was not:\nexpected: %+v\nactual: %+v\n", expectedHostOnlyNetwork, n)
	}
}

func TestListHostOnlyNetworks(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "list hostonlyifs",
		stdOut: stdOutOneHostOnlyNetwork,
	}

	nets, err := listHostOnlyAdapters(vbox)

	assert.Equal(t, 1, len(nets))
	assert.NoError(t, err)

	net, present := nets["HostInterfaceNetworking-vboxnet0"]

	assert.True(t, present)
	assert.Equal(t, "vboxnet0", net.Name)
	assert.Equal(t, "786f6276-656e-4074-8000-0a0027000000", net.GUID)
	assert.False(t, net.DHCP)
	assert.Equal(t, "192.168.99.1", net.IPv4.IP.String())
	assert.Equal(t, "ffffff00", net.IPv4.Mask.String())
	assert.Equal(t, "0a:00:27:00:00:00", net.HwAddr.String())
	assert.Equal(t, "Ethernet", net.Medium)
	assert.Equal(t, "Up", net.Status)
	assert.Equal(t, "HostInterfaceNetworking-vboxnet0", net.NetworkName)
}

func TestListTwoHostOnlyNetworks(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "list hostonlyifs",
		stdOut: stdOutTwoHostOnlyNetwork,
	}

	nets, err := listHostOnlyAdapters(vbox)

	assert.Equal(t, 2, len(nets))
	assert.NoError(t, err)

	net, present := nets["HostInterfaceNetworking-vboxnet1"]

	assert.True(t, present)
	assert.Equal(t, "vboxnet1", net.Name)
	assert.Equal(t, "786f6276-656e-4174-8000-0a0027000001", net.GUID)
	assert.False(t, net.DHCP)
	assert.Equal(t, "169.254.37.187", net.IPv4.IP.String())
	assert.Equal(t, "ffffff00", net.IPv4.Mask.String())
	assert.Equal(t, "0a:00:27:00:00:01", net.HwAddr.String())
	assert.Equal(t, "Ethernet", net.Medium)
	assert.Equal(t, "Up", net.Status)
	assert.Equal(t, "HostInterfaceNetworking-vboxnet1", net.NetworkName)
}

func TestListHostOnlyNetworksDontRelyOnEmptyLinesForParsing(t *testing.T) {
	vbox := &VBoxManagerMock{
		args: "list hostonlyifs",
		stdOut: `Name:            vboxnet0
VBoxNetworkName: HostInterfaceNetworking-vboxnet0
Name:            vboxnet1
VBoxNetworkName: HostInterfaceNetworking-vboxnet1`,
	}

	nets, err := listHostOnlyAdapters(vbox)

	assert.Equal(t, 2, len(nets))
	assert.NoError(t, err)

	net, present := nets["HostInterfaceNetworking-vboxnet1"]
	assert.True(t, present)
	assert.Equal(t, "vboxnet1", net.Name)

	net, present = nets["HostInterfaceNetworking-vboxnet0"]
	assert.True(t, present)
	assert.Equal(t, "vboxnet0", net.Name)
}

func TestGetHostOnlyNetwork(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "list hostonlyifs",
		stdOut: stdOutOneHostOnlyNetwork,
	}
	nets, err := listHostOnlyAdapters(vbox)
	assert.NoError(t, err)

	net, err := getOrCreateHostOnlyNetwork(net.ParseIP("192.168.99.1"), parseIPv4Mask("255.255.255.0"), nets, vbox)

	assert.NotNil(t, net)
	assert.Equal(t, "HostInterfaceNetworking-vboxnet0", net.NetworkName)
	assert.NoError(t, err)
}

func TestFailIfTwoNetworksHaveSameIP(t *testing.T) {
	vbox := &VBoxManagerMock{
		args: "list hostonlyifs",
		stdOut: `Name:            vboxnet0
IPAddress:       192.168.99.1
NetworkMask:     255.255.255.0
VBoxNetworkName: HostInterfaceNetworking-vboxnet0
Name:            vboxnet1
IPAddress:       192.168.99.1
NetworkMask:     255.255.255.0
VBoxNetworkName: HostInterfaceNetworking-vboxnet1`,
	}
	nets, err := listHostOnlyAdapters(vbox)
	assert.Nil(t, nets)
	assert.EqualError(t, err, `VirtualBox is configured with multiple host-only adapters with the same IP "192.168.99.1". Please remove one`)
}

func TestFailIfTwoNetworksHaveSameName(t *testing.T) {
	vbox := &VBoxManagerMock{
		args: "list hostonlyifs",
		stdOut: `Name:            vboxnet0
VBoxNetworkName: HostInterfaceNetworking-vboxnet0
Name:            vboxnet0
VBoxNetworkName: HostInterfaceNetworking-vboxnet0`,
	}
	nets, err := listHostOnlyAdapters(vbox)
	assert.Nil(t, nets)
	assert.EqualError(t, err, `VirtualBox is configured with multiple host-only adapters with the same name "HostInterfaceNetworking-vboxnet0". Please remove one`)
}

func TestGetDHCPServers(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "list dhcpservers",
		stdOut: stdOutListTwoDHCPServers,
	}

	servers, err := listDHCPServers(vbox)

	assert.Equal(t, 2, len(servers))
	assert.NoError(t, err)

	server, present := servers["HostInterfaceNetworking-vboxnet0"]
	assert.True(t, present)
	assert.Equal(t, "HostInterfaceNetworking-vboxnet0", server.NetworkName)
	assert.Equal(t, "192.168.99.6", server.IPv4.IP.String())
	assert.Equal(t, "192.168.99.100", server.LowerIP.String())
	assert.Equal(t, "192.168.99.254", server.UpperIP.String())
	assert.Equal(t, "ffffff00", server.IPv4.Mask.String())
	assert.True(t, server.Enabled)

	server, present = servers["HostInterfaceNetworking-vboxnet1"]
	assert.True(t, present)
	assert.Equal(t, "HostInterfaceNetworking-vboxnet1", server.NetworkName)
	assert.Equal(t, "192.168.99.7", server.IPv4.IP.String())
	assert.Equal(t, "192.168.99.100", server.LowerIP.String())
	assert.Equal(t, "192.168.99.254", server.UpperIP.String())
	assert.Equal(t, "ffffff00", server.IPv4.Mask.String())
	assert.False(t, server.Enabled)
}

// Tests detection of a conflict between prospective vbox host-only network and an IPV6 host interface
func TestCheckIPNetCollisionIPv6(t *testing.T) {
	m := map[string]*net.IPNet{}
	_, vboxHostOnly, err := net.ParseCIDR("2607:f8b0:400e:c04:ffff:ffff:ffff:ffff/64")
	assert.Nil(t, err)

	hostIP, hostNet, err := net.ParseCIDR("2001:4998:c:a06::2:4008/64")
	assert.Nil(t, err)
	m[hostIP.String()] = &net.IPNet{IP: hostIP, Mask: hostNet.Mask}

	result, err := checkIPNetCollision(vboxHostOnly, m)
	assert.Nil(t, err)
	assert.False(t, result)

	hostIP, hostNet, err = net.ParseCIDR("2607:f8b0:400e:c04::6a/64")
	assert.Nil(t, err)
	m[hostIP.String()] = &net.IPNet{IP: hostIP, Mask: hostNet.Mask}

	result, err = checkIPNetCollision(vboxHostOnly, m)
	assert.Nil(t, err)
	assert.True(t, result)
}

// Tests detection of a conflict between prospective vbox host-only network and an IPV4 host interface
func TestCheckIPNetCollisionIPv4(t *testing.T) {
	m := map[string]*net.IPNet{}
	_, vboxHostOnly, err := net.ParseCIDR("192.168.99.1/24")
	assert.NoError(t, err)

	hostIP, hostNet, err := net.ParseCIDR("10.10.10.42/24")
	assert.NoError(t, err)
	m[hostIP.String()] = &net.IPNet{IP: hostIP, Mask: hostNet.Mask}

	result, err := checkIPNetCollision(vboxHostOnly, m)
	assert.NoError(t, err)
	assert.False(t, result)

	hostIP, hostNet, err = net.ParseCIDR("192.168.99.22/24")
	assert.NoError(t, err)
	m[hostIP.String()] = &net.IPNet{IP: hostIP, Mask: hostNet.Mask}

	result, err = checkIPNetCollision(vboxHostOnly, m)
	assert.NoError(t, err)
	assert.True(t, result)
}

// Tests functionality of listHostInterfaces and verifies only non-loopback, active and non-excluded interfaces are returned
func TestListHostInterfaces(t *testing.T) {
	mhi := newMockHostInterfaces()
	excludes := map[string]*hostOnlyNetwork{}

	en0, err := mhi.addMockIface("10.10.0.22", 24, net.IPv4len, "en0", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)
	_, err = mhi.addMockIface("10.10.1.11", 24, net.IPv4len, "en1", net.FlagBroadcast /*not up*/)
	assert.NoError(t, err)
	_, err = mhi.addMockIface("127.0.0.1", 24, net.IPv4len, "lo0", net.FlagUp|net.FlagLoopback)
	assert.NoError(t, err)
	en0ipv6, err := mhi.addMockIface("2001:4998:c:a06::2:4008", 64, net.IPv6len, "en0ipv6", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)
	vboxnet0, err := mhi.addMockIface("192.168.99.1", 24, net.IPv4len, "vboxnet0", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)
	notvboxnet0, err := mhi.addMockIface("192.168.99.42", 24, net.IPv4len, "en2", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)

	excludes["192.168.99.1/24"] = &hostOnlyNetwork{IPv4: *vboxnet0, Name: "HostInterfaceNetworking-vboxnet0"}

	m, err := listHostInterfaces(mhi, excludes)
	assert.NoError(t, err)
	assert.NotEmpty(t, m)

	assert.Contains(t, m, "10.10.0.22/24")
	assert.Equal(t, en0, m["10.10.0.22/24"])

	assert.Contains(t, m, "2001:4998:c:a06::2:4008/64")
	assert.Equal(t, en0ipv6, m["2001:4998:c:a06::2:4008/64"])

	assert.Contains(t, m, "192.168.99.42/24")
	assert.Equal(t, notvboxnet0, m["192.168.99.42/24"])

	assert.NotContains(t, m, "10.10.1.11/24")
	assert.NotContains(t, m, "127.0.0.1/24")
	assert.NotContains(t, m, "192.168.99.1/24")
}
