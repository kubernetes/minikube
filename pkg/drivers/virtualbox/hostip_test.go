/*
Copyright 2026 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package virtualbox

import (
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// multiCmdVBoxMock implements VBoxManager and dispatches different responses
// for different argv strings. Needed for HostInterfaceIP which issues two
// distinct commands (showvminfo + list hostonlynets or list hostonlyifs).
type multiCmdVBoxMock struct {
	responses map[string]vbmResponse
}

type vbmResponse struct {
	stdOut string
	stdErr string
	err    error
}

func (m *multiCmdVBoxMock) vbm(args ...string) error {
	_, _, err := m.vbmOutErr(args...)
	return err
}

func (m *multiCmdVBoxMock) vbmOut(args ...string) (string, error) {
	stdout, _, err := m.vbmOutErr(args...)
	return stdout, err
}

func (m *multiCmdVBoxMock) vbmOutErr(args ...string) (string, string, error) {
	key := strings.Join(args, " ")
	if r, ok := m.responses[key]; ok {
		return r.stdOut, r.stdErr, r.err
	}
	return "", "", errors.New("multiCmdVBoxMock: unexpected args " + key)
}

func TestFindHostIPInSubnet_Match(t *testing.T) {
	mhi := newMockHostInterfaces()
	_, err := mhi.addMockIface("10.0.0.5", 24, net.IPv4len, "en0", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)
	_, err = mhi.addMockIface("192.168.59.100", 24, net.IPv4len, "bridge100", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)

	_, network, err := net.ParseCIDR("192.168.59.0/24")
	assert.NoError(t, err)

	ip, err := findHostIPInSubnet(mhi, network.IP, network.Mask)
	assert.NoError(t, err)
	assert.Equal(t, "192.168.59.100", ip.String())
}

func TestFindHostIPInSubnet_NoMatch(t *testing.T) {
	mhi := newMockHostInterfaces()
	_, err := mhi.addMockIface("10.0.0.5", 24, net.IPv4len, "en0", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)

	_, network, err := net.ParseCIDR("192.168.59.0/24")
	assert.NoError(t, err)

	ip, err := findHostIPInSubnet(mhi, network.IP, network.Mask)
	assert.Error(t, err)
	assert.Nil(t, ip)
}

func TestFindHostIPInSubnet_IPv6Skipped(t *testing.T) {
	mhi := newMockHostInterfaces()
	_, err := mhi.addMockIface("fd00::1", 64, net.IPv6len, "en0", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)

	_, network, err := net.ParseCIDR("192.168.59.0/24")
	assert.NoError(t, err)

	ip, err := findHostIPInSubnet(mhi, network.IP, network.Mask)
	assert.Error(t, err)
	assert.Nil(t, ip)
}

func TestHostInterfaceIP_HostonlyNet(t *testing.T) {
	driver := NewDriver("default", "")
	driver.VBoxManager = &multiCmdVBoxMock{
		responses: map[string]vbmResponse{
			"showvminfo default --machinereadable": {
				stdOut: `hostonly-network2="minikube-hostonly-192.168.59.1"`,
			},
			"list hostonlynets": {
				stdOut: `Name:            minikube-hostonly-192.168.59.1
GUID:            00000000-0000-0000-0000-000000000001
State:           Enabled
NetworkMask:     255.255.255.0
LowerIP:         192.168.59.100
UpperIP:         192.168.59.254
VBoxNetworkName: hostonly-minikube-hostonly-192.168.59.1`,
			},
		},
	}
	mhi := newMockHostInterfaces()
	_, err := mhi.addMockIface("192.168.59.100", 24, net.IPv4len, "bridge100", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)
	driver.HostInterfaces = mhi

	ip, err := driver.HostInterfaceIP()
	assert.NoError(t, err)
	assert.Equal(t, "192.168.59.100", ip.String())
}

func TestHostInterfaceIP_HostonlyAdapter(t *testing.T) {
	driver := NewDriver("default", "")
	driver.VBoxManager = &multiCmdVBoxMock{
		responses: map[string]vbmResponse{
			"showvminfo default --machinereadable": {
				stdOut: `hostonlyadapter2="vboxnet0"`,
			},
			"list hostonlyifs": {
				stdOut: `Name:            vboxnet0
GUID:            786f6276-656e-4074-8000-0a0027000000
DHCP:            Disabled
IPAddress:       192.168.99.1
NetworkMask:     255.255.255.0
HardwareAddress: 0a:00:27:00:00:00
Status:          Up
VBoxNetworkName: HostInterfaceNetworking-vboxnet0`,
			},
		},
	}
	driver.HostInterfaces = newMockHostInterfaces()

	ip, err := driver.HostInterfaceIP()
	assert.NoError(t, err)
	assert.Equal(t, "192.168.99.1", ip.String())
}

func TestHostInterfaceIP_NoHostOnlyAdapter(t *testing.T) {
	driver := NewDriver("default", "")
	driver.VBoxManager = &multiCmdVBoxMock{
		responses: map[string]vbmResponse{
			"showvminfo default --machinereadable": {
				stdOut: `VMState="poweroff"`,
			},
		},
	}
	driver.HostInterfaces = newMockHostInterfaces()

	ip, err := driver.HostInterfaceIP()
	assert.Error(t, err)
	assert.Nil(t, ip)
}
