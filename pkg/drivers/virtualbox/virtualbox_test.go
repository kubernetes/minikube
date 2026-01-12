/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"net"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/state"
)

type VBoxManagerMock struct {
	args   string
	stdOut string
	stdErr string
	err    error
}

func (v *VBoxManagerMock) vbm(args ...string) error {
	_, _, err := v.vbmOutErr(args...)
	return err
}

func (v *VBoxManagerMock) vbmOut(args ...string) (string, error) {
	stdout, _, err := v.vbmOutErr(args...)
	return stdout, err
}

func (v *VBoxManagerMock) vbmOutErr(args ...string) (string, string, error) {
	if strings.Join(args, " ") == v.args {
		return v.stdOut, v.stdErr, v.err
	}
	return "", "", errors.New("Invalid args")
}

func newTestDriver() *Driver {
	return NewDriver("default", "")
}

func TestDriverName(t *testing.T) {
	driverName := newTestDriver().DriverName()

	assert.Equal(t, "virtualbox", driverName)
}

func TestSSHHostname(t *testing.T) {
	hostname, err := newTestDriver().GetSSHHostname()

	assert.Equal(t, "127.0.0.1", hostname)
	assert.NoError(t, err)
}

func TestDefaultSSHUsername(t *testing.T) {
	username := newTestDriver().GetSSHUsername()

	assert.Equal(t, "docker", username)
}

var parseShareFolderTestCases = []struct {
	shareFolder       string
	expectedShareDir  string
	expectedShareName string
}{
	{"dir:name", "dir", "name"},
	{"C:\\dir:name", "C:\\dir", "name"},
	{"C:\\:name", "C:\\", "name"},
}

func TestParseShareFolder(t *testing.T) {
	for _, parseShareFolderTestCase := range parseShareFolderTestCases {
		shareDir, shareName := parseShareFolder(parseShareFolderTestCase.shareFolder)

		assert.Equal(t, shareDir, parseShareFolderTestCase.expectedShareDir)
		assert.Equal(t, shareName, parseShareFolderTestCase.expectedShareName)
	}
}

func TestState(t *testing.T) {
	var tests = []struct {
		stdOut string
		state  state.State
	}{
		{`VMState="running"`, state.Running},
		{`VMState="paused"`, state.Paused},
		{`VMState="saved"`, state.Saved},
		{`VMState="poweroff"`, state.Stopped},
		{`VMState="aborted"`, state.Stopped},
		{`VMState="whatever"`, state.None},
		{`VMState=`, state.None},
	}

	for _, expected := range tests {
		driver := newTestDriver()
		driver.VBoxManager = &VBoxManagerMock{
			args:   "showvminfo default --machinereadable",
			stdOut: expected.stdOut,
		}

		machineState, err := driver.GetState()

		assert.NoError(t, err)
		assert.Equal(t, expected.state, machineState)
	}
}

func TestStateErrors(t *testing.T) {
	var tests = []struct {
		stdErr   string
		err      error
		finalErr error
	}{
		{"Could not find a registered machine named 'unknown'", errors.New("Bug"), errors.New("machine does not exist")},
		{"", errors.New("Unexpected error"), errors.New("Unexpected error")},
	}

	for _, expected := range tests {
		driver := newTestDriver()
		driver.VBoxManager = &VBoxManagerMock{
			args:   "showvminfo default --machinereadable",
			stdErr: expected.stdErr,
			err:    expected.err,
		}

		machineState, err := driver.GetState()

		assert.Equal(t, err, expected.finalErr)
		assert.Equal(t, state.Error, machineState)
	}
}

func TestGetRandomIPinSubnet(t *testing.T) {
	driver := newTestDriver()

	// test IP 1.2.3.4
	testIP := net.IPv4(byte(1), byte(2), byte(3), byte(4))
	newIP, err := getRandomIPinSubnet(driver, testIP)
	if err != nil {
		t.Fatal(err)
	}

	if testIP.Equal(newIP) {
		t.Fatalf("expected different IP (source %s); received %s", testIP.String(), newIP.String())
	}

	if newIP[0] != testIP[0] {
		t.Fatalf("expected first octet of %d; received %d", testIP[0], newIP[0])
	}

	if newIP[1] != testIP[1] {
		t.Fatalf("expected second octet of %d; received %d", testIP[1], newIP[1])
	}

	if newIP[2] != testIP[2] {
		t.Fatalf("expected third octet of %d; received %d", testIP[2], newIP[2])
	}
}

func TestGetHostOnlyMACAddress(t *testing.T) {
	driver := newTestDriver()
	driver.VBoxManager = &VBoxManagerMock{
		args:   "showvminfo default --machinereadable",
		stdOut: "unrelatedfield=whatever\nhostonlyadapter2=\"vboxnet1\"\nmacaddress2=\"004488AABBCC\"\n",
	}

	result, err := driver.getHostOnlyMACAddress()
	expected := "004488aabbcc"
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestGetHostOnlyMACAddressWhenNoHostOnlyAdapter(t *testing.T) {
	driver := newTestDriver()
	driver.VBoxManager = &VBoxManagerMock{
		args:   "showvminfo default --machinereadable",
		stdOut: "unrelatedfield=whatever\n",
	}

	result, err := driver.getHostOnlyMACAddress()
	assert.Empty(t, result)
	assert.Equal(t, err, errors.New("Machine does not have a host-only adapter"))
}

func TestParseIPForMACFromIPAddr(t *testing.T) {
	driver := newTestDriver()

	ipAddrOutput := "1: eth0:\n    link/ether 00:44:88:aa:bb:cc\n    inet 1.2.3.4/24\n2: eth1:\n    link/ether 11:55:99:dd:ee:ff\n   inet 5.6.7.8/24"

	result, err := driver.parseIPForMACFromIPAddr(ipAddrOutput, "004488aabbcc")
	assert.NoError(t, err)
	assert.Equal(t, result, "1.2.3.4")

	result, err = driver.parseIPForMACFromIPAddr(ipAddrOutput, "115599ddeeff")
	assert.NoError(t, err)
	assert.Equal(t, result, "5.6.7.8")

	result, err = driver.parseIPForMACFromIPAddr(ipAddrOutput, "000000000000")
	assert.Empty(t, result)
	assert.Equal(t, err, errors.New("Could not find matching IP for MAC address 000000000000"))
}

func TestGetIPErrors(t *testing.T) {
	var tests = []struct {
		stdOut   string
		err      error
		finalErr error
	}{
		{`VMState="poweroff"`, nil, errors.New("Host is not running")},
		{"", errors.New("Unable to get state"), errors.New("Unable to get state")},
	}

	for _, expected := range tests {
		driver := newTestDriver()
		driver.VBoxManager = &VBoxManagerMock{
			args:   "showvminfo default --machinereadable",
			stdOut: expected.stdOut,
			err:    expected.err,
		}

		ip, err := driver.GetIP()

		assert.Empty(t, ip)
		assert.Equal(t, err, expected.finalErr)

		url, err := driver.GetURL()

		assert.Empty(t, url)
		assert.Equal(t, err, expected.finalErr)
	}
}

func TestParseValidCIDR(t *testing.T) {
	ip, network, err := parseAndValidateCIDR("192.168.100.1/24")

	assert.Equal(t, "192.168.100.1", ip.String())
	assert.Equal(t, "192.168.100.0", network.IP.String())
	assert.Equal(t, "ffffff00", network.Mask.String())
	assert.NoError(t, err)
}

func TestInvalidCIDR(t *testing.T) {
	ip, network, err := parseAndValidateCIDR("192.168.100.1")

	assert.EqualError(t, err, "invalid CIDR address: 192.168.100.1")
	assert.Nil(t, ip)
	assert.Nil(t, network)
}

func TestInvalidNetworkIpCIDR(t *testing.T) {
	ip, network, err := parseAndValidateCIDR("192.168.100.0/24")

	assert.Equal(t, ErrNetworkAddrCidr, err)
	assert.Nil(t, ip)
	assert.Nil(t, network)
}

// Tests detection of a conflict between an existing vbox host-only network and a host network interface.  This
// scenario would happen if the docker-machine was created with the host on one network, and then the host gets
// moved to another network (e.g. different wifi routers)
func TestCIDRHostIFaceCollisionExisting(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "list hostonlyifs",
		stdOut: stdOutTwoHostOnlyNetwork,
	}
	mhi := newMockHostInterfaces()
	_, err := mhi.addMockIface("192.168.99.42", 24, net.IPv4len, "en0", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)

	nets, err := listHostOnlyAdapters(vbox)
	assert.NoError(t, err)
	m, listErr := listHostInterfaces(mhi, nets)
	assert.Nil(t, listErr)
	assert.NotEmpty(t, m)

	_, network, cidrErr := net.ParseCIDR("192.168.99.1/24")
	assert.Nil(t, cidrErr)
	err = validateNoIPCollisions(mhi, network, nets)
	assert.Equal(t, ErrNetworkAddrCollision, err)
}

// Tests operation of validateNoIPCollisions when no conflicts exist.
func TestCIDRHostIFaceNoCollision(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "list hostonlyifs",
		stdOut: stdOutTwoHostOnlyNetwork,
	}
	mhi := newMockHostInterfaces()
	_, err := mhi.addMockIface("10.10.0.22", 24, net.IPv4len, "en0", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)

	nets, err := listHostOnlyAdapters(vbox)
	assert.NoError(t, err)
	m, listErr := listHostInterfaces(mhi, nets)
	assert.Nil(t, listErr)
	assert.NotEmpty(t, m)

	_, network, cidrErr := net.ParseCIDR("192.168.99.1/24")
	assert.Nil(t, cidrErr)
	err = validateNoIPCollisions(mhi, network, nets)
	assert.NoError(t, err)
}

// Tests detection of a conflict between a potential vbox host-only network and a host network interface.
func TestCIDRHostIFaceCollision(t *testing.T) {
	vbox := &VBoxManagerMock{
		args:   "list hostonlyifs",
		stdOut: "",
	}
	mhi := newMockHostInterfaces()
	_, err := mhi.addMockIface("192.168.99.42", 24, net.IPv4len, "en0", net.FlagUp|net.FlagBroadcast)
	assert.NoError(t, err)

	nets, err := listHostOnlyAdapters(vbox)
	assert.NoError(t, err)
	m, listErr := listHostInterfaces(mhi, nets)
	assert.Nil(t, listErr)
	assert.NotEmpty(t, m)

	_, network, cidrErr := net.ParseCIDR("192.168.99.1/24")
	assert.Nil(t, cidrErr)
	err = validateNoIPCollisions(mhi, network, nets)
	assert.Equal(t, ErrNetworkAddrCollision, err)
}

// Tests the behavior of getDHCPAddressRange with a variety of subnets.
func TestGetDHCPAddressRange(t *testing.T) {
	tests := []struct {
		name            string
		dhcpAddrCIDR    string
		expectedLowerIP net.IP
		expectedUpperIP net.IP
	}{
		{
			"Test /8 CIDR",
			"10.0.0.14/8",
			net.ParseIP("10.0.0.100"),
			net.ParseIP("10.0.0.254"),
		},
		{
			"Test /24 CIDR",
			"192.168.99.7/24",
			net.ParseIP("192.168.99.100"),
			net.ParseIP("192.168.99.254"),
		},
		{
			"Test /25 CIDR",
			"100.121.20.19/25",
			net.ParseIP("100.121.20.20"),
			net.ParseIP("100.121.20.126"),
		},
		{
			"Test /28 CIDR",
			"100.121.10.8/28",
			net.ParseIP("100.121.10.9"),
			net.ParseIP("100.121.10.14"),
		},
	}

	getTestArgsFromCIDR := func(cidr string) (dhcpAddr net.IP, network *net.IPNet) {
		var err error
		dhcpAddr, network, err = net.ParseCIDR(cidr)
		assert.NoError(t, err, "Invalid CIDR %s", cidr)
		return
	}

	for _, tt := range tests {
		dhcpAddr, network := getTestArgsFromCIDR(tt.dhcpAddrCIDR)
		t.Run(tt.name, func(t *testing.T) {
			lowerIP, upperIP := getDHCPAddressRange(dhcpAddr, network)
			if !reflect.DeepEqual(lowerIP, tt.expectedLowerIP) {
				t.Errorf("getDHCPAddressRange() lowerIP = %v, want %v", lowerIP, tt.expectedLowerIP)
			}
			if !reflect.DeepEqual(upperIP, tt.expectedUpperIP) {
				t.Errorf("getDHCPAddressRange() upperIP = %v, want %v", upperIP, tt.expectedUpperIP)
			}
		})
	}
}

func TestSetConfigFromFlags(t *testing.T) {
	driver := newTestDriver()

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

type MockCreateOperations struct {
	test          *testing.T
	expectedCalls []Call
	call          int
}

type Call struct {
	signature string
	output    string
	err       error
}

func (v *MockCreateOperations) vbm(args ...string) error {
	_, _, err := v.vbmOutErr(args...)
	return err
}

func (v *MockCreateOperations) vbmOut(args ...string) (string, error) {
	stdout, _, err := v.vbmOutErr(args...)
	return stdout, err
}

func (v *MockCreateOperations) vbmOutErr(args ...string) (string, string, error) {
	output, err := v.doCall("vbm " + strings.Join(args, " "))
	return output, "", err
}

func (v *MockCreateOperations) UpdateISOCache(storePath, isoURL string) error {
	_, err := v.doCall("UpdateISOCache " + storePath + " " + isoURL)
	return err
}

func (v *MockCreateOperations) CopyIsoToMachineDir(storePath, machineName, isoURL string) error {
	_, err := v.doCall("CopyIsoToMachineDir " + storePath + " " + machineName + " " + isoURL)
	return err
}

func (v *MockCreateOperations) Generate(path string) error {
	_, err := v.doCall("Generate " + path)
	return err
}

func (v *MockCreateOperations) Create(size int, publicSSHKeyPath, diskPath string) error {
	_, err := v.doCall("Create " + fmt.Sprintf("%d %s %s", size, publicSSHKeyPath, diskPath))
	return err
}

func (v *MockCreateOperations) Read(path string) ([]string, error) {
	_, err := v.doCall("Read " + path)
	return []string{}, err
}

func (v *MockCreateOperations) Wait(d *Driver) error {
	_, err := v.doCall("WaitIP")
	return err
}

func (v *MockCreateOperations) RandomInt(n int) int {
	return 5
}

func (v *MockCreateOperations) Sleep(d time.Duration) {
	_, err := v.doCall("Sleep " + fmt.Sprintf("%v", d))
	if err != nil {
		v.test.Fatal(err)
	}
}

func (v *MockCreateOperations) Interfaces() ([]net.Interface, error) {
	_, err := v.doCall("Interfaces")
	return []net.Interface{}, err
}

func (v *MockCreateOperations) Addrs(iface *net.Interface) ([]net.Addr, error) {
	_, err := v.doCall("Addrs " + fmt.Sprintf("%v", iface))
	return []net.Addr{}, err
}

func (v *MockCreateOperations) doCall(callSignature string) (string, error) {
	if v.call >= len(v.expectedCalls) {
		v.test.Fatal("Unexpected call", callSignature)

	}

	call := v.expectedCalls[v.call]
	if call.signature != "IGNORE CALL" && (callSignature != call.signature) {
		v.test.Fatal("Unexpected call", callSignature)
	}

	v.call++

	return call.output, call.err
}

func mockCalls(t *testing.T, driver *Driver, expectedCalls []Call) {
	mockOperations := &MockCreateOperations{
		test:          t,
		expectedCalls: expectedCalls,
	}

	driver.Boot2DockerURL = "http://b2d.org"
	driver.VBoxManager = mockOperations
	driver.b2dUpdater = mockOperations
	driver.sshKeyGenerator = mockOperations
	driver.diskCreator = mockOperations
	driver.logsReader = mockOperations
	driver.ipWaiter = mockOperations
	driver.randomInter = mockOperations
	driver.sleeper = mockOperations
	driver.HostInterfaces = mockOperations
}

func TestCreateVM(t *testing.T) {
	shareName, shareDir := getShareDriveAndName()

	modifyVMcommand := "vbm modifyvm default --firmware bios --bioslogofadein off --bioslogofadeout off --bioslogodisplaytime 0 --biosbootmenu disabled --ostype Linux26_64 --cpus 1 --memory 1024 --acpi on --ioapic on --rtcuseutc on --natdnshostresolver1 off --natdnsproxy1 on --cpuhotplug off --pae on --hpet on --hwvirtex on --nestedpaging on --largepages on --vtxvpid on --accelerate3d off --boot1 dvd"
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		modifyVMcommand += " --longmode on"
	}

	driver := NewDriver("default", "path")
	mockCalls(t, driver, []Call{
		{"CopyIsoToMachineDir path default http://b2d.org", "", nil},
		{"Generate path/machines/default/id_rsa", "", nil},
		{"Create 20000 path/machines/default/id_rsa.pub path/machines/default/disk.vmdk", "", nil},
		{"vbm createvm --basefolder path/machines/default --name default --register", "", nil},
		{modifyVMcommand, "", nil},
		{"vbm modifyvm default --nic1 nat --nictype1 82540EM --cableconnected1 on", "", nil},
		{"vbm storagectl default --name SATA --add sata --hostiocache on", "", nil},
		{"vbm storageattach default --storagectl SATA --port 0 --device 0 --type dvddrive --medium path/machines/default/boot2docker.iso", "", nil},
		{"vbm storageattach default --storagectl SATA --port 1 --device 0 --type hdd --medium path/machines/default/disk.vmdk", "", nil},
		{"vbm guestproperty set default /VirtualBox/GuestAdd/SharedFolders/MountPrefix /", "", nil},
		{"vbm guestproperty set default /VirtualBox/GuestAdd/SharedFolders/MountDir /", "", nil},
		{"vbm sharedfolder add default --name " + shareName + " --hostpath " + shareDir + " --automount", "", nil},
		{"vbm setextradata default VBoxInternal2/SharedFoldersEnableSymlinksCreate/" + shareName + " 1", "", nil},
	})

	err := driver.CreateVM()

	assert.NoError(t, err)
}

func TestCreateVMWithSpecificNatNicType(t *testing.T) {
	shareName, shareDir := getShareDriveAndName()

	modifyVMcommand := "vbm modifyvm default --firmware bios --bioslogofadein off --bioslogofadeout off --bioslogodisplaytime 0 --biosbootmenu disabled --ostype Linux26_64 --cpus 1 --memory 1024 --acpi on --ioapic on --rtcuseutc on --natdnshostresolver1 off --natdnsproxy1 on --cpuhotplug off --pae on --hpet on --hwvirtex on --nestedpaging on --largepages on --vtxvpid on --accelerate3d off --boot1 dvd"
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		modifyVMcommand += " --longmode on"
	}

	driver := NewDriver("default", "path")
	driver.NatNicType = "Am79C973"
	mockCalls(t, driver, []Call{
		{"CopyIsoToMachineDir path default http://b2d.org", "", nil},
		{"Generate path/machines/default/id_rsa", "", nil},
		{"Create 20000 path/machines/default/id_rsa.pub path/machines/default/disk.vmdk", "", nil},
		{"vbm createvm --basefolder path/machines/default --name default --register", "", nil},
		{modifyVMcommand, "", nil},
		{"vbm modifyvm default --nic1 nat --nictype1 Am79C973 --cableconnected1 on", "", nil},
		{"vbm storagectl default --name SATA --add sata --hostiocache on", "", nil},
		{"vbm storageattach default --storagectl SATA --port 0 --device 0 --type dvddrive --medium path/machines/default/boot2docker.iso", "", nil},
		{"vbm storageattach default --storagectl SATA --port 1 --device 0 --type hdd --medium path/machines/default/disk.vmdk", "", nil},
		{"vbm guestproperty set default /VirtualBox/GuestAdd/SharedFolders/MountPrefix /", "", nil},
		{"vbm guestproperty set default /VirtualBox/GuestAdd/SharedFolders/MountDir /", "", nil},
		{"vbm sharedfolder add default --name " + shareName + " --hostpath " + shareDir + " --automount", "", nil},
		{"vbm setextradata default VBoxInternal2/SharedFoldersEnableSymlinksCreate/" + shareName + " 1", "", nil},
	})

	err := driver.CreateVM()

	assert.NoError(t, err)
}

func TestCreateVMWithoutAccelerate3D(t *testing.T) {
	shareName, shareDir := getShareDriveAndName()

	modifyVMcommand := "vbm modifyvm default --firmware bios --bioslogofadein off --bioslogofadeout off --bioslogodisplaytime 0 --biosbootmenu disabled --ostype Linux26_64 --cpus 1 --memory 1024 --acpi on --ioapic on --rtcuseutc on --natdnshostresolver1 off --natdnsproxy1 on --cpuhotplug off --pae on --hpet on --hwvirtex on --nestedpaging on --largepages on --vtxvpid on --boot1 dvd"
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		modifyVMcommand += " --longmode on"
	}

	driver := NewDriver("default", "path")
	driver.NoAccelerate3DOff = true
	mockCalls(t, driver, []Call{
		{"CopyIsoToMachineDir path default http://b2d.org", "", nil},
		{"Generate path/machines/default/id_rsa", "", nil},
		{"Create 20000 path/machines/default/id_rsa.pub path/machines/default/disk.vmdk", "", nil},
		{"vbm createvm --basefolder path/machines/default --name default --register", "", nil},
		{modifyVMcommand, "", nil},
		{"vbm modifyvm default --nic1 nat --nictype1 82540EM --cableconnected1 on", "", nil},
		{"vbm storagectl default --name SATA --add sata --hostiocache on", "", nil},
		{"vbm storageattach default --storagectl SATA --port 0 --device 0 --type dvddrive --medium path/machines/default/boot2docker.iso", "", nil},
		{"vbm storageattach default --storagectl SATA --port 1 --device 0 --type hdd --medium path/machines/default/disk.vmdk", "", nil},
		{"vbm guestproperty set default /VirtualBox/GuestAdd/SharedFolders/MountPrefix /", "", nil},
		{"vbm guestproperty set default /VirtualBox/GuestAdd/SharedFolders/MountDir /", "", nil},
		{"vbm sharedfolder add default --name " + shareName + " --hostpath " + shareDir + " --automount", "", nil},
		{"vbm setextradata default VBoxInternal2/SharedFoldersEnableSymlinksCreate/" + shareName + " 1", "", nil},
	})

	err := driver.CreateVM()

	assert.NoError(t, err)
}

func TestStart(t *testing.T) {
	driver := NewDriver("default", "path")
	mockCalls(t, driver, []Call{
		{"vbm showvminfo default --machinereadable", `VMState="poweroff"`, nil},
		{"vbm list hostonlyifs", "", nil},
		{"Interfaces", "", nil},
		{"vbm hostonlyif create", "Interface 'VirtualBox Host-Only Ethernet Adapter' was successfully created", nil},
		{"vbm list hostonlyifs", `
Name:            VirtualBox Host-Only Ethernet Adapter
GUID:            786f6276-656e-4074-8000-0a0027000000
DHCP:            Disabled
IPAddress:       192.168.99.1
NetworkMask:     255.255.255.0
IPV6Address:
IPV6NetworkMaskPrefixLength: 0
HardwareAddress: 0a:00:27:00:00:00
MediumType:      Ethernet
Status:          Up
VBoxNetworkName: HostInterfaceNetworking-VirtualBox Host-Only Ethernet Adapter`, nil},
		{"vbm hostonlyif ipconfig VirtualBox Host-Only Ethernet Adapter --ip 192.168.99.1 --netmask 255.255.255.0", "", nil},
		{"vbm list dhcpservers", "", nil},
		{"vbm list dhcpservers", "", nil},
		{"vbm dhcpserver add --netname HostInterfaceNetworking-VirtualBox Host-Only Ethernet Adapter --ip 192.168.99.6 --netmask 255.255.255.0 --lowerip 192.168.99.100 --upperip 192.168.99.254 --enable", "", nil},
		{"vbm modifyvm default --nic2 hostonly --nictype2 82540EM --nicpromisc2 deny --hostonlyadapter2 VirtualBox Host-Only Ethernet Adapter --cableconnected2 on", "", nil},
		{"IGNORE CALL", "", nil},
		{"IGNORE CALL", "", nil},
		{"vbm startvm default --type headless", "", nil},
		{"Read path/machines/default/default/Logs/VBox.log", "", nil},
		{"WaitIP", "", nil},
		{"vbm list hostonlyifs", `
Name:            VirtualBox Host-Only Ethernet Adapter
GUID:            786f6276-656e-4074-8000-0a0027000000
DHCP:            Disabled
IPAddress:       192.168.99.1
NetworkMask:     255.255.255.0
IPV6Address:
IPV6NetworkMaskPrefixLength: 0
HardwareAddress: 0a:00:27:00:00:00
MediumType:      Ethernet
Status:          Up
VBoxNetworkName: HostInterfaceNetworking-VirtualBox Host-Only Ethernet Adapter`, nil},
		{"Interfaces", "", nil},
	})

	err := driver.Start()

	assert.NoError(t, err)
}

func TestStartWithHostOnlyAdapterCreationBug(t *testing.T) {
	driver := NewDriver("default", "path")
	mockCalls(t, driver, []Call{
		{"vbm showvminfo default --machinereadable", `VMState="poweroff"`, nil},
		{"vbm list hostonlyifs", "", nil},
		{"Interfaces", "", nil},
		{"vbm hostonlyif create", "", errors.New("error: Failed to create the host-only adapter")},
		{"vbm list hostonlyifs", "", nil},
		{"vbm list hostonlyifs", `
Name:            VirtualBox Host-Only Ethernet Adapter
GUID:            786f6276-656e-4074-8000-0a0027000000
DHCP:            Disabled
IPAddress:       192.168.99.1
NetworkMask:     255.255.255.0
IPV6Address:
IPV6NetworkMaskPrefixLength: 0
HardwareAddress: 0a:00:27:00:00:00
MediumType:      Ethernet
Status:          Up
VBoxNetworkName: HostInterfaceNetworking-VirtualBox Host-Only Ethernet Adapter`, nil},
		{"vbm hostonlyif ipconfig VirtualBox Host-Only Ethernet Adapter --ip 192.168.99.1 --netmask 255.255.255.0", "", nil},
		{"vbm list dhcpservers", "", nil},
		{"vbm list dhcpservers", "", nil},
		{"vbm dhcpserver add --netname HostInterfaceNetworking-VirtualBox Host-Only Ethernet Adapter --ip 192.168.99.6 --netmask 255.255.255.0 --lowerip 192.168.99.100 --upperip 192.168.99.254 --enable", "", nil},
		{"vbm modifyvm default --nic2 hostonly --nictype2 82540EM --nicpromisc2 deny --hostonlyadapter2 VirtualBox Host-Only Ethernet Adapter --cableconnected2 on", "", nil},
		{"IGNORE CALL", "", nil},
		{"IGNORE CALL", "", nil},
		{"vbm startvm default --type headless", "", nil},
		{"Read path/machines/default/default/Logs/VBox.log", "", nil},
		{"WaitIP", "", nil},
		{"vbm list hostonlyifs", `
Name:            VirtualBox Host-Only Ethernet Adapter
GUID:            786f6276-656e-4074-8000-0a0027000000
DHCP:            Disabled
IPAddress:       192.168.99.100
NetworkMask:     255.255.255.0
IPV6Address:
IPV6NetworkMaskPrefixLength: 0
HardwareAddress: 0a:00:27:00:00:00
MediumType:      Ethernet
Status:          Up
VBoxNetworkName: HostInterfaceNetworking-VirtualBox Host-Only Ethernet Adapter`, nil},
		{"Interfaces", "", nil},
		{"vbm showvminfo default --machinereadable", `VMState="running"`, nil},
		{"vbm controlvm default acpipowerbutton", "", nil},
		{"vbm showvminfo default --machinereadable", `VMState="stopped"`, nil},
		{"Sleep 5s", "", nil},
		{"vbm hostonlyif ipconfig VirtualBox Host-Only Ethernet Adapter --ip 192.168.99.1 --netmask 255.255.255.0", "", nil},
		{"Sleep 5s", "", nil},
		{"vbm startvm default --type headless", "", nil},
		{"WaitIP", "", nil},
	})

	err := driver.Start()

	assert.NoError(t, err)
}

func TestRemoveStopped(t *testing.T) {
	driver := NewDriver("default", "path")
	mockCalls(t, driver, []Call{
		{"vbm showvminfo default --machinereadable", `VMState="poweroff"`, nil},
		{"vbm unregistervm --delete default", "", nil},
	})

	err := driver.Remove()

	assert.NoError(t, err)
}

func TestRemoveStarted(t *testing.T) {
	driver := NewDriver("default", "path")
	mockCalls(t, driver, []Call{
		{"vbm showvminfo default --machinereadable", `VMState="running"`, nil},
		{"vbm controlvm default poweroff", "", nil},
		{"vbm unregistervm --delete default", "", nil},
	})

	err := driver.Remove()

	assert.NoError(t, err)
}

func TestRemoveSaved(t *testing.T) {
	driver := NewDriver("default", "path")
	mockCalls(t, driver, []Call{
		{"vbm showvminfo default --machinereadable", `VMState="saved"`, nil},
		{"vbm unregistervm --delete default", "", nil},
	})

	err := driver.Remove()

	assert.NoError(t, err)
}

func TestRemovePaused(t *testing.T) {
	driver := NewDriver("default", "path")
	mockCalls(t, driver, []Call{
		{"vbm showvminfo default --machinereadable", `VMState="running"`, nil},
		{"vbm controlvm default poweroff", "", nil},
		{"vbm unregistervm --delete default", "", nil},
	})

	err := driver.Remove()

	assert.NoError(t, err)
}
