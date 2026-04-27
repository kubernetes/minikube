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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/state"
)

const (
	defaultCPU                   = 1
	defaultMemory                = 1024
	defaultBoot2DockerURL        = ""
	defaultBoot2DockerImportVM   = ""
	defaultHostOnlyCIDR          = "192.168.99.1/24"
	defaultHostOnlyNictype       = "82540EM"
	defaultHostOnlyPromiscMode   = "deny"
	defaultUIType                = "headless"
	defaultHostOnlyNoDHCP        = false
	defaultDiskSize              = 20000
	defaultDNSProxy              = true
	defaultDNSResolver           = false
	defaultHostLoopbackReachable = true
)

// hostOnlyNicIndex is the NIC slot used for the host-only adapter/network
// on the guest VM. NIC1 is NAT (for SSH port-forward); NIC2 is host-only
// (for host-to-guest and inter-node reachability).
const hostOnlyNicIndex = 2

//nolint:staticcheck // ST1005: error strings should not be capitalized
var (
	ErrUnableToGenerateRandomIP = errors.New("unable to generate random IP")
	ErrMustEnableVTX            = errors.New("This computer doesn't have VT-X/AMD-v enabled. Enabling it in the BIOS is mandatory")
	ErrNotCompatibleWithHyperV  = errors.New("This computer is running Hyper-V. VirtualBox won't boot a 64bits VM when Hyper-V is activated. Either use Hyper-V as a driver, or disable the Hyper-V hypervisor. (To skip this check, use --virtualbox-no-vtx-check)")
	ErrNetworkAddrCidr          = errors.New("host-only cidr must be specified with a host address, not a network address")
	ErrNetworkAddrCollision     = errors.New("host-only cidr conflicts with the network address of a host interface")
)

type Driver struct {
	*drivers.BaseDriver
	VBoxManager
	HostInterfaces
	b2dUpdater            B2DUpdater
	sshKeyGenerator       SSHKeyGenerator
	diskCreator           DiskCreator
	logsReader            LogsReader
	ipWaiter              IPWaiter
	randomInter           RandomInter
	sleeper               Sleeper
	version               int
	CPU                   int
	Memory                int
	DiskSize              int
	NatNicType            string
	Boot2DockerURL        string
	Boot2DockerImportVM   string
	HostDNSResolver       bool
	HostLoopbackReachable bool
	HostOnlyCIDR          string
	HostOnlyNicType       string
	HostOnlyPromiscMode   string
	UIType                string
	HostOnlyNoDHCP        bool
	NoShare               bool
	DNSProxy              bool
	NoVTXCheck            bool
	NoAccelerate3DOff     bool
	ShareFolder           string
}

// NewDriver creates a new VirtualBox driver with default settings.
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		VBoxManager:           NewVBoxManager(),
		b2dUpdater:            NewB2DUpdater(),
		sshKeyGenerator:       NewSSHKeyGenerator(),
		diskCreator:           NewDiskCreator(),
		logsReader:            NewLogsReader(),
		ipWaiter:              NewIPWaiter(),
		randomInter:           NewRandomInter(),
		sleeper:               NewSleeper(),
		HostInterfaces:        NewHostInterfaces(),
		Memory:                defaultMemory,
		CPU:                   defaultCPU,
		DiskSize:              defaultDiskSize,
		NatNicType:            defaultHostOnlyNictype,
		HostOnlyCIDR:          defaultHostOnlyCIDR,
		HostOnlyNicType:       defaultHostOnlyNictype,
		HostOnlyPromiscMode:   defaultHostOnlyPromiscMode,
		UIType:                defaultUIType,
		HostOnlyNoDHCP:        defaultHostOnlyNoDHCP,
		DNSProxy:              defaultDNSProxy,
		HostDNSResolver:       defaultDNSResolver,
		HostLoopbackReachable: defaultHostLoopbackReachable,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.IntFlag{
			Name:   "virtualbox-memory",
			Usage:  "Size of memory for host in MB",
			Value:  defaultMemory,
			EnvVar: "VIRTUALBOX_MEMORY_SIZE",
		},
		mcnflag.IntFlag{
			Name:   "virtualbox-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  defaultCPU,
			EnvVar: "VIRTUALBOX_CPU_COUNT",
		},
		mcnflag.IntFlag{
			Name:   "virtualbox-disk-size",
			Usage:  "Size of disk for host in MB",
			Value:  defaultDiskSize,
			EnvVar: "VIRTUALBOX_DISK_SIZE",
		},
		mcnflag.StringFlag{
			Name:   "virtualbox-boot2docker-url",
			Usage:  "The URL of the boot2docker image. Defaults to the latest available version",
			Value:  defaultBoot2DockerURL,
			EnvVar: "VIRTUALBOX_BOOT2DOCKER_URL",
		},
		mcnflag.StringFlag{
			Name:   "virtualbox-import-boot2docker-vm",
			Usage:  "The name of a Boot2Docker VM to import",
			Value:  defaultBoot2DockerImportVM,
			EnvVar: "VIRTUALBOX_BOOT2DOCKER_IMPORT_VM",
		},
		mcnflag.BoolFlag{
			Name:   "virtualbox-host-dns-resolver",
			Usage:  "Use the host DNS resolver",
			EnvVar: "VIRTUALBOX_HOST_DNS_RESOLVER",
		},
		mcnflag.BoolFlag{
			Name:   "virtualbox-host-loopback-reachable",
			Usage:  "Enable host loopback interface accessibility",
			EnvVar: "VIRTUALBOX_HOST_LOOPBACK_REACHABLE",
		},
		mcnflag.StringFlag{
			Name:   "virtualbox-nat-nictype",
			Usage:  "Specify the Network Adapter Type",
			Value:  defaultHostOnlyNictype,
			EnvVar: "VIRTUALBOX_NAT_NICTYPE",
		},
		mcnflag.StringFlag{
			Name:   "virtualbox-hostonly-cidr",
			Usage:  "Specify the Host Only CIDR",
			Value:  defaultHostOnlyCIDR,
			EnvVar: "VIRTUALBOX_HOSTONLY_CIDR",
		},
		mcnflag.StringFlag{
			Name:   "virtualbox-hostonly-nictype",
			Usage:  "Specify the Host Only Network Adapter Type",
			Value:  defaultHostOnlyNictype,
			EnvVar: "VIRTUALBOX_HOSTONLY_NIC_TYPE",
		},
		mcnflag.StringFlag{
			Name:   "virtualbox-hostonly-nicpromisc",
			Usage:  "Specify the Host Only Network Adapter Promiscuous Mode",
			Value:  defaultHostOnlyPromiscMode,
			EnvVar: "VIRTUALBOX_HOSTONLY_NIC_PROMISC",
		},
		mcnflag.StringFlag{
			Name:   "virtualbox-ui-type",
			Usage:  "Specify the UI Type: (gui|sdl|headless|separate)",
			Value:  defaultUIType,
			EnvVar: "VIRTUALBOX_UI_TYPE",
		},
		mcnflag.BoolFlag{
			Name:   "virtualbox-hostonly-no-dhcp",
			Usage:  "Disable the Host Only DHCP Server",
			EnvVar: "VIRTUALBOX_HOSTONLY_NO_DHCP",
		},
		mcnflag.BoolFlag{
			Name:   "virtualbox-no-share",
			Usage:  "Disable the mount of your home directory",
			EnvVar: "VIRTUALBOX_NO_SHARE",
		},
		mcnflag.BoolFlag{
			Name:   "virtualbox-no-dns-proxy",
			Usage:  "Disable proxying all DNS requests to the host",
			EnvVar: "VIRTUALBOX_NO_DNS_PROXY",
		},
		mcnflag.BoolFlag{
			Name:   "virtualbox-no-vtx-check",
			Usage:  "Disable checking for the availability of hardware virtualization before the vm is started",
			EnvVar: "VIRTUALBOX_NO_VTX_CHECK",
		},
		mcnflag.BoolFlag{
			Name:   "virtualbox-no-accelerate3d-off",
			Usage:  "Disable turning off the possibly missing 3D graphics acceleration before the vm is started",
			EnvVar: "VIRTUALBOX_NO_ACCELERATE3D_OFF",
		},
		mcnflag.StringFlag{
			EnvVar: "VIRTUALBOX_SHARE_FOLDER",
			Name:   "virtualbox-share-folder",
			Usage:  "Mount the specified directory instead of the default home location. Format: dir:name",
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return "127.0.0.1", nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "virtualbox"
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.CPU = flags.Int("virtualbox-cpu-count")
	d.Memory = flags.Int("virtualbox-memory")
	d.DiskSize = flags.Int("virtualbox-disk-size")
	d.Boot2DockerURL = flags.String("virtualbox-boot2docker-url")
	d.SetSwarmConfigFromFlags(flags)
	d.SSHUser = "docker"
	d.Boot2DockerImportVM = flags.String("virtualbox-import-boot2docker-vm")
	d.HostDNSResolver = flags.Bool("virtualbox-host-dns-resolver")
	d.HostLoopbackReachable = flags.Bool("virtualbox-host-loopback-reachable")
	d.NatNicType = flags.String("virtualbox-nat-nictype")
	d.HostOnlyCIDR = flags.String("virtualbox-hostonly-cidr")
	d.HostOnlyNicType = flags.String("virtualbox-hostonly-nictype")
	d.HostOnlyPromiscMode = flags.String("virtualbox-hostonly-nicpromisc")
	d.UIType = flags.String("virtualbox-ui-type")
	d.HostOnlyNoDHCP = flags.Bool("virtualbox-hostonly-no-dhcp")
	d.NoShare = flags.Bool("virtualbox-no-share")
	d.DNSProxy = !flags.Bool("virtualbox-no-dns-proxy")
	d.NoVTXCheck = flags.Bool("virtualbox-no-vtx-check")
	d.NoAccelerate3DOff = flags.Bool("virtualbox-no-accelerate3d-off")
	d.ShareFolder = flags.String("virtualbox-share-folder")

	return nil
}

// PreCreateCheck checks that VBoxManage exists and works
func (d *Driver) PreCreateCheck() error {
	// Check that VBoxManage exists and works
	version, err := d.vbmOut("--version")
	if err != nil {
		return err
	}

	// Check that VBoxManage is of a supported version
	if err = checkVBoxManageVersion(strings.TrimSpace(version)); err != nil {
		return err
	}

	d.version, _, _ = parseVersion(strings.TrimSpace(version))

	if !d.NoVTXCheck {
		if isHyperVInstalled() {
			return ErrNotCompatibleWithHyperV
		}

		if d.IsVTXDisabled() {
			return ErrMustEnableVTX
		}
	}

	// Downloading boot2docker to cache should be done here to make sure
	// that a download failure will not leave a machine half created.
	if err := d.b2dUpdater.UpdateISOCache(d.StorePath, d.Boot2DockerURL); err != nil {
		return err
	}

	// Check that Host-only interfaces are ok
	if _, err = listHostOnlyAdapters(d.VBoxManager); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Create() error {
	if err := d.CreateVM(); err != nil {
		return err
	}

	log.Info("Starting the VM...")
	return d.Start()
}

func importBoot2Docker(d *Driver, name string) error {
	// make sure vm is stopped
	_ = d.vbm("controlvm", name, "poweroff")

	diskInfo, err := getVMDiskInfo(name, d.VBoxManager)
	if err != nil {
		return err
	}

	if _, err := os.Stat(diskInfo.Path); err != nil {
		return err
	}

	if err := d.vbm("clonehd", diskInfo.Path, d.diskPath()); err != nil {
		return err
	}

	log.Debugf("Importing VM settings...")
	vmInfo, err := getVMInfo(name, d.VBoxManager)
	if err != nil {
		return err
	}

	d.CPU = vmInfo.CPUs
	d.Memory = vmInfo.Memory

	return nil
}

func (d *Driver) CreateVM() error {
	if err := d.b2dUpdater.CopyIsoToMachineDir(d.StorePath, d.MachineName, d.Boot2DockerURL); err != nil {
		return err
	}

	log.Info("Creating VirtualBox VM...")

	// import b2d VM if requested
	if d.Boot2DockerImportVM != "" {
		name := d.Boot2DockerImportVM

		if err := importBoot2Docker(d, name); err != nil {
			return err
		}

		log.Debugf("Importing SSH key...")
		keyPath := filepath.Join(mcnutils.GetHomeDir(), ".ssh", "id_boot2docker")
		if err := mcnutils.CopyFile(keyPath, d.GetSSHKeyPath()); err != nil {
			return err
		}
	} else {
		log.Infof("Creating SSH key...")
		if err := d.sshKeyGenerator.Generate(d.GetSSHKeyPath()); err != nil {
			return err
		}

		log.Debugf("Creating disk image...")
		if err := d.diskCreator.Create(d.DiskSize, d.publicSSHKeyPath(), d.diskPath()); err != nil {
			return err
		}
	}

	if err := d.vbm("createvm",
		"--basefolder", d.ResolveStorePath("."),
		"--name", d.MachineName,
		"--register"); err != nil {
		return err
	}

	log.Debugf("VM CPUS: %d", d.CPU)
	log.Debugf("VM Memory: %d", d.Memory)

	cpus := d.CPU
	if cpus < 1 {
		cpus = runtime.NumCPU()
	}
	if cpus > 32 {
		cpus = 32
	}

	hostDNSResolver := "off"
	if d.HostDNSResolver {
		hostDNSResolver = "on"
	}

	hostLoopbackReachable := "off"
	if d.HostLoopbackReachable {
		hostLoopbackReachable = "on"
	}

	dnsProxy := "off"
	if d.DNSProxy {
		dnsProxy = "on"
	}

	if err := d.vbm(d.buildModifyVMFlags(cpus, hostDNSResolver, dnsProxy, hostLoopbackReachable)...); err != nil {
		return err
	}

	if err := d.vbm("modifyvm", d.MachineName,
		"--nic1", "nat",
		"--nictype1", d.NatNicType,
		"--cableconnected1", "on"); err != nil {
		return err
	}

	if err := d.vbm("storagectl", d.MachineName,
		"--name", "SATA",
		"--add", "sata",
		"--hostiocache", "on"); err != nil {
		return err
	}

	if err := d.vbm("storageattach", d.MachineName,
		"--storagectl", "SATA",
		"--port", "0",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", d.ResolveStorePath("boot2docker.iso")); err != nil {
		return err
	}

	if err := d.vbm("storageattach", d.MachineName,
		"--storagectl", "SATA",
		"--port", "1",
		"--device", "0",
		"--type", "hdd",
		"--medium", d.diskPath()); err != nil {
		return err
	}

	// let VBoxService do nice magic automounting (when it's used)
	if err := d.vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/"); err != nil {
		return err
	}
	if err := d.vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountDir", "/"); err != nil {
		return err
	}

	shareName, shareDir := getShareDriveAndName()

	if d.ShareFolder != "" {
		shareDir, shareName = parseShareFolder(d.ShareFolder)
	}

	if shareDir != "" && !d.NoShare {
		if err := setupShareDir(d, shareDir, shareName); err != nil {
			return err
		}
	}

	return nil
}

// buildModifyVMFlags returns the `VBoxManage modifyvm` argv used by CreateVM.
// The darwin/arm64 path uses the VBox 7.x kebab-case flags and ARM-appropriate
// chipset/firmware/graphics. Other platforms keep the legacy x86 flags.
func (d *Driver) buildModifyVMFlags(cpus int, hostDNSResolver, dnsProxy, hostLoopbackReachable string) []string {
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		// darwin/arm64 (Apple Silicon): VirtualBox 7.1+ creates ARM guests via
		// Apple's Hypervisor.framework. Several x86-only flags (--hwvirtex,
		// --nested-paging, --large-pages, --pae, --hpet, --vtxvpid) are not
		// accepted on ARM VMs. Flag names also use the kebab-case VBox 7.x
		// form, the firmware must be EFI (no BIOS on ARM), the chipset must
		// be armv8virtual, and the graphics controller must be qemuramfb
		// because the default vboxvga is x86-only and fails VM start with
		// VERR_PGM_RAM_CONFLICT on ARM guests.
		flags := []string{
			"modifyvm", d.MachineName,
			"--chipset", "armv8virtual",
			"--firmware", "efi64",
			"--graphicscontroller", "qemuramfb",
			"--firmware-logo-fade-in", "off",
			"--firmware-logo-fade-out", "off",
			"--firmware-logo-display-time", "0",
			"--firmware-boot-menu", "disabled",
			"--ostype", "Linux_arm64",
			"--cpus", fmt.Sprintf("%d", cpus),
			"--memory", fmt.Sprintf("%d", d.Memory),
			"--acpi", "on",
			"--ioapic", "on",
			"--rtc-use-utc", "on",
			"--natdnshostresolver1", hostDNSResolver,
			"--natdnsproxy1", dnsProxy,
			"--cpu-hotplug", "off",
		}
		if !d.NoAccelerate3DOff {
			flags = append(flags, "--accelerate-3d", "off")
		}
		flags = append(flags, "--boot1", "dvd")
		if d.version > 6 {
			flags = append(flags, "--natlocalhostreachable1", hostLoopbackReachable)
		}
		return flags
	}

	flags := []string{
		"modifyvm", d.MachineName,
		"--firmware", "bios",
		"--bioslogofadein", "off",
		"--bioslogofadeout", "off",
		"--bioslogodisplaytime", "0",
		"--biosbootmenu", "disabled",
		"--ostype", "Linux26_64",
		"--cpus", fmt.Sprintf("%d", cpus),
		"--memory", fmt.Sprintf("%d", d.Memory),
		"--acpi", "on",
		"--ioapic", "on",
		"--rtcuseutc", "on",
		"--natdnshostresolver1", hostDNSResolver,
		"--natdnsproxy1", dnsProxy,
		"--cpuhotplug", "off",
		"--pae", "on",
		"--hpet", "on",
		"--hwvirtex", "on",
		"--nestedpaging", "on",
		"--largepages", "on",
		"--vtxvpid", "on",
	}
	if !d.NoAccelerate3DOff {
		flags = append(flags, "--accelerate3d", "off")
	}
	flags = append(flags, "--boot1", "dvd")
	if d.version > 6 {
		flags = append(flags, "--natlocalhostreachable1", hostLoopbackReachable)
	}
	if runtime.GOOS == "windows" && runtime.GOARCH == "386" {
		flags = append(flags, "--longmode", "on")
	}
	return flags
}

func parseShareFolder(shareFolder string) (string, string) {
	split := strings.Split(shareFolder, ":")
	shareDir := strings.Join(split[:len(split)-1], ":")
	shareName := split[len(split)-1]
	return shareDir, shareName
}

func setupShareDir(d *Driver, shareDir, shareName string) error {
	log.Debugf("setting up shareDir '%s' -> '%s'", shareDir, shareName)
	if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
		return err
	} else if !os.IsNotExist(err) {
		if shareName == "" {
			// parts of the VBox internal code are buggy with share names that start with "/"
			shareName = strings.TrimLeft(shareDir, "/")
			// TODO do some basic Windows -> MSYS path conversion
			// ie, s!^([a-z]+):[/\\]+!\1/!; s!\\!/!g
		}

		// woo, shareDir exists!  let's carry on!
		if err := d.vbm("sharedfolder", "add", d.MachineName, "--name", shareName, "--hostpath", shareDir, "--automount"); err != nil {
			return err
		}

		// enable symlinks
		if err := d.vbm("setextradata", d.MachineName, "VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+shareName, "1"); err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) hostOnlyIPAvailable() bool {
	ip, err := d.GetIP()
	if err != nil {
		log.Debugf("ERROR getting IP: %s", err)
		return false
	}
	if ip == "" {
		log.Debug("Strangely, there was no error attempting to get the IP, but it was still empty.")
		return false
	}

	log.Debugf("IP is %s", ip)
	return true
}

func (d *Driver) Start() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	var hostOnlyAdapter *hostOnlyNetwork
	if s == state.Stopped {
		log.Infof("Check network to re-create if needed...")

		if hostOnlyAdapter, err = d.setupHostOnlyNetwork(d.MachineName); err != nil {
			//nolint:staticcheck // ST1005: error strings should not be capitalized
			return fmt.Errorf("Error setting up host only network on machine start: %s", err)
		}
	}

	switch s {
	case state.Stopped, state.Saved:
		d.SSHPort, err = setPortForwarding(d, 1, "ssh", "tcp", 22, d.SSHPort)
		if err != nil {
			return err
		}

		if err := d.vbm("startvm", d.MachineName, "--type", d.UIType); err != nil {
			if lines, readErr := d.readVBoxLog(); readErr == nil && len(lines) > 0 {
				//nolint:staticcheck // ST1005: error strings should not be capitalized
				return fmt.Errorf("Unable to start the VM: %s\nDetails: %s", err, lines[len(lines)-1])
			}
			//nolint:staticcheck // ST1005: error strings should not be capitalized
			return fmt.Errorf("Unable to start the VM: %s", err)
		}
	case state.Paused:
		if err := d.vbm("controlvm", d.MachineName, "resume", "--type", d.UIType); err != nil {
			return err
		}
		log.Infof("Resuming VM ...")
	default:
		log.Infof("VM not in restartable state")
	}

	if !d.NoVTXCheck {
		// Verify that VT-X is not disabled in the started VM
		vtxIsDisabled, err := d.IsVTXDisabledInTheVM()
		if err != nil {
			//nolint:staticcheck // ST1005: error strings should not be capitalized
			return fmt.Errorf("Checking if hardware virtualization is enabled failed: %s", err)
		}

		if vtxIsDisabled {
			return ErrMustEnableVTX
		}
	}

	log.Infof("Waiting for an IP...")
	if err := d.ipWaiter.Wait(d); err != nil {
		return err
	}

	if hostOnlyAdapter == nil {
		return nil
	}

	// On darwin/arm64 the hostonlynet API manages the network; the legacy
	// adapter corruption check is not applicable and listHostOnlyAdapters
	// would fail without /dev/vboxnetctl.
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return nil
	}

	// Check that the host-only adapter we just created can still be found
	// Sometimes it is corrupted after the VM is started.
	nets, err := listHostOnlyAdapters(d.VBoxManager)
	if err != nil {
		return err
	}

	ip, network, err := parseAndValidateCIDR(d.HostOnlyCIDR)
	if err != nil {
		return err
	}

	err = validateNoIPCollisions(d.HostInterfaces, network, nets)
	if err != nil {
		return err
	}

	hostOnlyNet := getHostOnlyAdapter(nets, ip, network.Mask)
	if hostOnlyNet != nil {
		// OK, we found a valid host-only adapter
		return nil
	}

	// This happens a lot on windows. The adapter has an invalid IP and the VM has the same IP
	log.Warn("The host-only adapter is corrupted. Let's stop the VM, fix the host-only adapter and restart the VM")
	if err := d.Stop(); err != nil {
		return err
	}

	// We have to be sure the host-only adapter is not used by the VM
	d.sleeper.Sleep(5 * time.Second)

	log.Debugf("Fixing %+v...", hostOnlyAdapter)
	if err := hostOnlyAdapter.SaveIPv4(d.VBoxManager); err != nil {
		return err
	}

	// We have to be sure the adapter is updated before starting the VM
	d.sleeper.Sleep(5 * time.Second)

	if err := d.vbm("startvm", d.MachineName, "--type", d.UIType); err != nil {
		//nolint:staticcheck // ST1005: error strings should not be capitalized
		return fmt.Errorf("Unable to start the VM: %s", err)
	}

	log.Infof("Waiting for an IP...")
	return d.ipWaiter.Wait(d)
}

func (d *Driver) Stop() error {
	currentState, err := d.GetState()
	if err != nil {
		return err
	}

	if currentState == state.Paused {
		if err := d.vbm("controlvm", d.MachineName, "resume"); err != nil { // , "--type", d.UIType
			return err
		}
		log.Infof("Resuming VM ...")
	}

	if err := d.vbm("controlvm", d.MachineName, "acpipowerbutton"); err != nil {
		return err
	}
	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}
		if s == state.Running {
			d.sleeper.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	d.IPAddress = ""

	return nil
}

// Restart restarts a machine which is known to be running.
func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		//nolint:staticcheck // ST1005: error strings should not be capitalized
		return fmt.Errorf("Problem stopping the VM: %s", err)
	}

	if err := d.Start(); err != nil {
		//nolint:staticcheck // ST1005: error strings should not be capitalized
		return fmt.Errorf("Problem starting the VM: %s", err)
	}

	d.IPAddress = ""

	return d.ipWaiter.Wait(d)
}

func (d *Driver) Kill() error {
	return d.vbm("controlvm", d.MachineName, "poweroff")
}

func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err == ErrMachineNotExist {
		return nil
	}
	if err != nil {
		return err
	}

	if s != state.Stopped && s != state.Saved {
		if err := d.Kill(); err != nil {
			return err
		}
	}

	return d.vbm("unregistervm", "--delete", d.MachineName)
}

func (d *Driver) GetState() (state.State, error) {
	stdout, stderr, err := d.vbmOutErr("showvminfo", d.MachineName, "--machinereadable")
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return state.Error, ErrMachineNotExist
		}
		return state.Error, err
	}
	re := regexp.MustCompile(`(?m)^VMState="(\w+)"`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 1 {
		return state.None, nil
	}
	switch groups[1] {
	case "running":
		return state.Running, nil
	case "paused":
		return state.Paused, nil
	case "saved":
		return state.Saved, nil
	case "poweroff", "aborted":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) getHostOnlyMACAddress() (string, error) {
	// Return the MAC address of the host-only adapter
	// assigned to this machine. The returned address
	// is lower-cased and does not contain colons.

	stdout, stderr, err := d.vbmOutErr("showvminfo", d.MachineName, "--machinereadable")
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return "", ErrMachineNotExist
		}
		return "", err
	}

	// First, we get the number of the host-only interface. VBox 7.x ARM
	// (darwin/arm64) uses the hostonly-network<N> field name (new hostonlynet
	// API); older/x86 configurations use hostonlyadapter<N>. The regex accepts
	// any NIC index defensively; in practice it matches hostOnlyNicIndex.
	re := regexp.MustCompile(`(?m)^(?:hostonlyadapter|hostonly-network)([\d]+)`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 2 {
		//nolint:staticcheck // ST1005: error strings should not be capitalized
		return "", errors.New("Machine does not have a host-only adapter")
	}

	// Then we grab the MAC address based on that number
	adapterNumber := groups[1]
	re = regexp.MustCompile(fmt.Sprintf("(?m)^macaddress%s=\"(.*)\"", adapterNumber))
	groups = re.FindStringSubmatch(stdout)
	if len(groups) < 2 {
		//nolint:staticcheck // ST1005: error strings should not be capitalized
		return "", fmt.Errorf("Could not find MAC address for adapter %v", adapterNumber)
	}

	return strings.ToLower(groups[1]), nil
}

// HostInterfaceIP returns the host's IPv4 address on the host-only network
// attached to this VM. On darwin/arm64 (VBox 7.x hostonlynet API) the host
// address is auto-assigned by VirtualBox and not discoverable from VBox
// metadata directly, so it is resolved by enumerating local interfaces for
// one whose subnet matches the hostonlynet. On other platforms the legacy
// hostonlyif path returns the IPAddress exposed by `list hostonlyifs`.
func (d *Driver) HostInterfaceIP() (net.IP, error) {
	stdout, _, err := d.vbmOutErr("showvminfo", d.MachineName, "--machinereadable")
	if err != nil {
		return nil, fmt.Errorf("showvminfo: %w", err)
	}

	// VBox 7.x ARM (darwin/arm64) uses the hostonlynet API, which exposes
	// the attached network as hostonly-network<N>= in showvminfo output
	// (rather than the legacy hostonlyadapter<N>= for hostonlyif).
	// `list hostonlyifs` is empty on that path, so look up `list
	// hostonlynets` and find the matching host-side IP by enumerating
	// local interfaces whose subnet overlaps the hostonlynet.
	if m := regexp.MustCompile(`hostonly-network\d+="(.*?)"`).FindStringSubmatch(stdout); m != nil {
		netName := m[1]
		netList, _, err := d.vbmOutErr("list", "hostonlynets")
		if err != nil {
			return nil, fmt.Errorf("list hostonlynets: %w", err)
		}
		mask, netAddr, err := parseHostOnlyNet(netList, netName)
		if err != nil {
			return nil, fmt.Errorf("hostonlynet %q: %w", netName, err)
		}
		if mask == nil {
			return nil, fmt.Errorf("hostonlynet %q not found in `VBoxManage list hostonlynets` output", netName)
		}
		return findHostIPInSubnet(d.HostInterfaces, netAddr, mask)
	}

	// Legacy hostonlyif path: find the adapter named in showvminfo and
	// extract the IPAddress from `list hostonlyifs`.
	m := regexp.MustCompile(`hostonlyadapter\d+="(.*?)"`).FindStringSubmatch(stdout)
	if m == nil {
		return nil, fmt.Errorf("VM %q has no host-only adapter", d.MachineName)
	}
	iface := m[1]
	list, _, err := d.vbmOutErr("list", "hostonlyifs")
	if err != nil {
		return nil, fmt.Errorf("list hostonlyifs: %w", err)
	}
	re := regexp.MustCompile(`(?sm)Name:\s*` + regexp.QuoteMeta(iface) + `\s*$.+?IPAddress:\s*(\S+)`)
	im := re.FindStringSubmatch(list)
	if im == nil {
		return nil, fmt.Errorf("IP not found for host-only adapter %q", iface)
	}
	ip := net.ParseIP(im[1])
	if ip == nil {
		return nil, fmt.Errorf("unable to parse host-only IPAddress %q", im[1])
	}
	return ip, nil
}

// findHostIPInSubnet returns the first IPv4 address on any local interface
// that lies in the given subnet. Used on darwin/arm64 with the hostonlynet
// API where the host-side IP is auto-assigned by VirtualBox and the actual
// address is only discoverable by enumerating local interfaces.
func findHostIPInSubnet(hif HostInterfaces, netAddr net.IP, mask net.IPMask) (net.IP, error) {
	ints, err := hif.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, in := range ints {
		addrs, err := hif.Addrs(&in)
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil {
				continue
			}
			if ip4.Mask(mask).Equal(netAddr) {
				return ip4, nil
			}
		}
	}
	return nil, fmt.Errorf("no local interface found with an IPv4 address in %s/%d", netAddr, maskOnes(mask))
}

func maskOnes(mask net.IPMask) int {
	ones, _ := mask.Size()
	return ones
}

// parseHostOnlyNet extracts the NetworkMask and LowerIP for the named
// hostonlynet from the output of `VBoxManage list hostonlynets`. The
// function is order-tolerant within a record and tolerant of whitespace
// variation. Returns nil, nil if the name isn't found.
//
// Records are delimited by a line starting with "Name:". Blank lines within
// a record (VBoxManage emits one between GUID and State) are ignored.
func parseHostOnlyNet(listOutput, name string) (mask net.IPMask, network net.IP, err error) {
	var blockName, maskStr, lowerStr string
	// finish evaluates the current record against `name` and, if matched,
	// parses and returns the mask/network. It is called when a new record
	// begins (next "Name:" line) and at end-of-input.
	finish := func() (bool, net.IPMask, net.IP, error) {
		if blockName != name {
			return false, nil, nil, nil
		}
		if maskStr == "" || lowerStr == "" {
			return true, nil, nil, fmt.Errorf("hostonlynet %q record missing NetworkMask or LowerIP", name)
		}
		mip := net.ParseIP(maskStr).To4()
		if mip == nil {
			return true, nil, nil, fmt.Errorf("hostonlynet %q: unable to parse NetworkMask %q", name, maskStr)
		}
		lip := net.ParseIP(lowerStr).To4()
		if lip == nil {
			return true, nil, nil, fmt.Errorf("hostonlynet %q: unable to parse LowerIP %q", name, lowerStr)
		}
		m := net.IPMask(mip)
		return true, m, lip.Mask(m), nil
	}

	for _, line := range strings.Split(listOutput, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "Name":
			// Starting a new record: evaluate the previous one.
			if blockName != "" {
				if matched, m, n, err := finish(); matched {
					return m, n, err
				}
			}
			blockName, maskStr, lowerStr = val, "", ""
		case "NetworkMask":
			maskStr = val
		case "LowerIP":
			lowerStr = val
		}
	}
	// Evaluate the last record.
	if blockName != "" {
		if matched, m, n, err := finish(); matched {
			return m, n, err
		}
	}
	return nil, nil, nil
}

func (d *Driver) parseIPForMACFromIPAddr(ipAddrOutput string, macAddress string) (string, error) {
	// Given the output of "ip addr show" on the VM, return the IPv4 address
	// of the interface with the given MAC address.

	lines := strings.Split(ipAddrOutput, "\n")
	returnNextIP := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "link") { // line contains MAC address
			vals := strings.Split(line, " ")
			if len(vals) >= 2 {
				macBlock := vals[1]
				macWithoutColons := strings.ReplaceAll(macBlock, ":", "")
				if macWithoutColons == macAddress { // we are in the correct device block
					returnNextIP = true
				}
			}
		} else if strings.HasPrefix(line, "inet") && !strings.HasPrefix(line, "inet6") && returnNextIP {
			vals := strings.Split(line, " ")
			if len(vals) >= 2 {
				idx := strings.Index(vals[1], "/")
				if idx != -1 {
					return vals[1][:idx], nil
				}
				return vals[1], nil
			}
		}
	}

	//nolint:staticcheck // ST1005: error strings should not be capitalized
	return "", fmt.Errorf("Could not find matching IP for MAC address %v", macAddress)
}

func (d *Driver) GetIP() (string, error) {
	// DHCP is used to get the IP, so virtualbox hosts don't have IPs unless
	// they are running
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	macAddress, err := d.getHostOnlyMACAddress()
	if err != nil {
		return "", err
	}

	log.Debugf("Host-only MAC: %s\n", macAddress)

	output, err := drivers.RunSSHCommandFromDriver(d, "ip addr show")
	if err != nil {
		return "", err
	}

	log.Debugf("SSH returned: %s\nEND SSH\n", output)

	ipAddress, err := d.parseIPForMACFromIPAddr(output, macAddress)
	if err != nil {
		return "", err
	}

	return ipAddress, nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return d.ResolveStorePath("disk.vmdk")
}

func (d *Driver) setupHostOnlyNetwork(machineName string) (*hostOnlyNetwork, error) {
	hostOnlyCIDR := d.HostOnlyCIDR

	// This is to assist in migrating from version 0.2 to 0.3 format
	// it should be removed in a later release
	if hostOnlyCIDR == "" {
		hostOnlyCIDR = defaultHostOnlyCIDR
	}

	ip, network, err := parseAndValidateCIDR(hostOnlyCIDR)
	if err != nil {
		return nil, err
	}

	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return d.setupHostOnlyNetworkVBox7(machineName, ip, network)
	}

	nets, err := listHostOnlyAdapters(d.VBoxManager)
	if err != nil {
		return nil, err
	}

	err = validateNoIPCollisions(d.HostInterfaces, network, nets)
	if err != nil {
		return nil, err
	}

	log.Debugf("Searching for hostonly interface for IPv4: %s and Mask: %s", ip, network.Mask)
	hostOnlyAdapter, err := getOrCreateHostOnlyNetwork(ip, network.Mask, nets, d.VBoxManager)
	if err != nil {
		return nil, err
	}

	if err := removeOrphanDHCPServers(d.VBoxManager); err != nil {
		return nil, err
	}

	dhcpAddr, err := getRandomIPinSubnet(d, ip)
	if err != nil {
		return nil, err
	}

	lowerIP, upperIP := getDHCPAddressRange(dhcpAddr, network)

	log.Debugf("Adding/Modifying DHCP server %q with address range %q - %q...", dhcpAddr, lowerIP, upperIP)

	dhcp := dhcpServer{}
	dhcp.IPv4.IP = dhcpAddr
	dhcp.IPv4.Mask = network.Mask
	dhcp.LowerIP = lowerIP
	dhcp.UpperIP = upperIP
	dhcp.Enabled = !d.HostOnlyNoDHCP
	if err := addHostOnlyDHCPServer(hostOnlyAdapter.Name, dhcp, d.VBoxManager); err != nil {
		return nil, err
	}

	if err := d.vbm("modifyvm", machineName,
		fmt.Sprintf("--nic%d", hostOnlyNicIndex), "hostonly",
		fmt.Sprintf("--nictype%d", hostOnlyNicIndex), d.HostOnlyNicType,
		fmt.Sprintf("--nicpromisc%d", hostOnlyNicIndex), d.HostOnlyPromiscMode,
		fmt.Sprintf("--hostonlyadapter%d", hostOnlyNicIndex), hostOnlyAdapter.Name,
		fmt.Sprintf("--cableconnected%d", hostOnlyNicIndex), "on"); err != nil {
		return nil, err
	}

	return hostOnlyAdapter, nil
}

// setupHostOnlyNetworkVBox7 implements the host-only network setup using the
// VBox 7.x hostonlynet API. This is required on darwin/arm64 because the
// legacy hostonlyif mechanism depends on the /dev/vboxnetctl kernel extension
// which does not exist on Apple Silicon.
//
// Unlike the legacy path, hostonlynet bundles IP-range/DHCP config into the
// `hostonlynet add` command, so no separate DHCP server management is needed.
func (d *Driver) setupHostOnlyNetworkVBox7(machineName string, ip net.IP, network *net.IPNet) (*hostOnlyNetwork, error) {
	nets, err := listHostOnlyNets(d.VBoxManager)
	if err != nil {
		return nil, err
	}

	// Shape the hostonlynet list the way validateNoIPCollisions expects, so
	// the host-side interface that VBox creates for each active hostonlynet
	// (e.g. bridge100 on darwin via vmnet.framework) is excluded from the
	// collision scan on a subsequent start that reuses the network.
	//
	// The hostonlynet API does not report the host-side address directly, so
	// it is discovered by scanning live interfaces for one whose IPv4 falls
	// in the net's subnet. Nets with no live host-side interface (e.g. the
	// associated VM is not running) are skipped; there is nothing to exclude.
	legacyShape := map[string]*hostOnlyNetwork{}
	for k, n := range nets {
		netAddr := n.LowerIP.Mask(n.NetworkMask)
		hostIP, err := findHostIPInSubnet(d.HostInterfaces, netAddr, n.NetworkMask)
		if err != nil {
			continue
		}
		legacyShape[k] = &hostOnlyNetwork{
			IPv4: net.IPNet{
				IP:   hostIP,
				Mask: n.NetworkMask,
			},
		}
	}
	if err := validateNoIPCollisions(d.HostInterfaces, network, legacyShape); err != nil {
		return nil, err
	}

	lowerIP, upperIP := getDHCPAddressRange(ip, network)

	// Reuse an existing hostonlynet if its subnet matches the requested one;
	// otherwise create a new one.
	hon := findHostOnlyNetByCIDR(nets, ip, network.Mask)
	if hon == nil {
		// NOTE: the name format "minikube-hostonly-<ip>" is read back by
		// pkg/minikube/cluster/ip.go's HostIP() when resolving the host's
		// address. Keep the two sites in sync if this format changes.
		name := fmt.Sprintf("minikube-hostonly-%s", ip.String())
		hon, err = createHostOnlyNet(d.VBoxManager, name, network.Mask, lowerIP, upperIP)
		if err != nil {
			return nil, err
		}
	}

	// Attach NIC2 to the hostonlynet. Note the flag names differ from the
	// legacy path: --nic2 hostonlynet (not hostonly) and --host-only-net2
	// (not --hostonlyadapter2). --nictype2 and --nicpromisc2 and
	// --cableconnected2 are still accepted.
	if err := d.vbm("modifyvm", machineName,
		fmt.Sprintf("--nic%d", hostOnlyNicIndex), "hostonlynet",
		fmt.Sprintf("--nictype%d", hostOnlyNicIndex), d.HostOnlyNicType,
		fmt.Sprintf("--nicpromisc%d", hostOnlyNicIndex), d.HostOnlyPromiscMode,
		fmt.Sprintf("--host-only-net%d", hostOnlyNicIndex), hon.Name,
		fmt.Sprintf("--cableconnected%d", hostOnlyNicIndex), "on"); err != nil {
		return nil, err
	}

	// Return a hostOnlyNetwork-shaped struct so the caller's type expectation
	// holds. Only Name is used downstream on this path.
	return &hostOnlyNetwork{
		Name:        hon.Name,
		IPv4:        net.IPNet{IP: ip, Mask: network.Mask},
		NetworkName: hon.VBoxNetworkName,
	}, nil
}

func getDHCPAddressRange(dhcpAddr net.IP, network *net.IPNet) (lowerIP net.IP, upperIP net.IP) {
	nAddr := network.IP.To4()
	ones, bits := network.Mask.Size()

	if ones <= 24 {
		// For a /24 subnet, use the original behavior of allowing the address range
		// between x.x.x.100 and x.x.x.254.
		lowerIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(100))
		upperIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(254))
		return
	}

	// Start the lowerIP range one address above the selected DHCP address.
	lowerIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], dhcpAddr.To4()[3]+1)

	// The highest D-part of the address A.B.C.D in this subnet is at 2^n - 1,
	// where n is the number of available bits in the subnet. Since the highest
	// address is reserved for subnet broadcast, the highest *assignable* address
	// is at (2^n - 1) - 1 == 2^n - 2.
	maxAssignableSubnetAddress := (byte)((1 << (uint)(bits-ones)) - 2)
	upperIP = net.IPv4(nAddr[0], nAddr[1], nAddr[2], maxAssignableSubnetAddress)
	return
}

func parseAndValidateCIDR(hostOnlyCIDR string) (net.IP, *net.IPNet, error) {
	ip, network, err := net.ParseCIDR(hostOnlyCIDR)
	if err != nil {
		return nil, nil, err
	}

	networkAddress := network.IP.To4()
	if ip.Equal(networkAddress) {
		return nil, nil, ErrNetworkAddrCidr
	}

	return ip, network, nil
}

// validateNoIPCollisions ensures no conflicts between the host's network interfaces and the vbox host-only network that
// will be used for machine vm instances.
func validateNoIPCollisions(hif HostInterfaces, hostOnlyNet *net.IPNet, currHostOnlyNets map[string]*hostOnlyNetwork) error {
	hostOnlyByCIDR := map[string]*hostOnlyNetwork{}
	// listHostOnlyAdapters returns a map w/ virtualbox net names as key.  Rekey to CIDRs
	for _, n := range currHostOnlyNets {
		ipnet := net.IPNet{IP: n.IPv4.IP, Mask: n.IPv4.Mask}
		hostOnlyByCIDR[ipnet.String()] = n
	}

	m, err := listHostInterfaces(hif, hostOnlyByCIDR)
	if err != nil {
		return err
	}

	collision := checkIPNetCollision(hostOnlyNet, m)

	if collision {
		return ErrNetworkAddrCollision
	}
	return nil
}

// Select an available port, trying the specified
// port first, falling back on an OS selected port.
func getAvailableTCPPort(port int) (int, error) {
	for i := 0; i <= 10; i++ {
		ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			return 0, err
		}
		defer ln.Close()
		addr := ln.Addr().String()
		addrParts := strings.SplitN(addr, ":", 2)
		p, err := strconv.Atoi(addrParts[1])
		if err != nil {
			return 0, err
		}
		if p != 0 {
			port = p
			return port, nil
		}
		port = 0 // Throw away the port hint before trying again
		time.Sleep(1 * time.Second)
	}
	return 0, errors.New("unable to allocate tcp port")
}

// Setup a NAT port forwarding entry.
func setPortForwarding(d *Driver, interfaceNum int, mapName, protocol string, guestPort, desiredHostPort int) (int, error) {
	actualHostPort, err := getAvailableTCPPort(desiredHostPort)
	if err != nil {
		return -1, err
	}
	if desiredHostPort != actualHostPort && desiredHostPort != 0 {
		log.Debugf("NAT forwarding host port for guest port %d (%s) changed from %d to %d",
			guestPort, mapName, desiredHostPort, actualHostPort)
	}
	cmd := fmt.Sprintf("--natpf%d", interfaceNum)
	_ = d.vbm("modifyvm", d.MachineName, cmd, "delete", mapName)
	if err := d.vbm("modifyvm", d.MachineName,
		cmd, fmt.Sprintf("%s,%s,127.0.0.1,%d,,%d", mapName, protocol, actualHostPort, guestPort)); err != nil {
		return -1, err
	}
	return actualHostPort, nil
}

// getRandomIPinSubnet returns a pseudo-random net.IP in the same
// subnet as the IP passed
func getRandomIPinSubnet(d *Driver, baseIP net.IP) (net.IP, error) {
	var dhcpAddr net.IP

	nAddr := baseIP.To4()
	// select pseudo-random DHCP addr; make sure not to clash with the host
	// only try 5 times and bail if no random received
	for i := 0; i < 5; i++ {
		n := d.randomInter.RandomInt(24) + 1
		if byte(n) != nAddr[3] {
			dhcpAddr = net.IPv4(nAddr[0], nAddr[1], nAddr[2], byte(n))
			break
		}
	}

	if dhcpAddr == nil {
		return nil, ErrUnableToGenerateRandomIP
	}

	return dhcpAddr, nil
}

func detectVBoxManageCmdInPath() string {
	cmd := "VBoxManage"
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}

func (d *Driver) readVBoxLog() ([]string, error) {
	logPath := filepath.Join(d.ResolveStorePath(d.MachineName), "Logs", "VBox.log")
	log.Debugf("Checking vm logs: %s", logPath)

	return d.logsReader.Read(logPath)
}
