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

package hyperv

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/ssh"
	"k8s.io/minikube/pkg/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	Boot2DockerURL       string
	VSwitch              string
	DiskSize             int
	MemSize              int
	CPU                  int
	MacAddr              string
	VLanID               int
	DisableDynamicMemory bool
}

const (
	defaultDiskSize             = 20000
	defaultMemory               = 1024
	defaultCPU                  = 1
	defaultVLanID               = 0
	defaultDisableDynamicMemory = false
	defaultSwitchID             = "c08cb7b8-9b3c-408e-8e30-5e16a3aeb444"
)

// NewDriver creates a new Hyper-v driver with default settings.
func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		DiskSize:             defaultDiskSize,
		MemSize:              defaultMemory,
		CPU:                  defaultCPU,
		DisableDynamicMemory: defaultDisableDynamicMemory,
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
		mcnflag.StringFlag{
			Name:   "hyperv-boot2docker-url",
			Usage:  "URL of the boot2docker ISO. Defaults to the latest available version.",
			EnvVar: "HYPERV_BOOT2DOCKER_URL",
		},
		mcnflag.StringFlag{
			Name:   "hyperv-virtual-switch",
			Usage:  "Virtual switch name. Defaults to first found.",
			EnvVar: "HYPERV_VIRTUAL_SWITCH",
		},
		mcnflag.IntFlag{
			Name:   "hyperv-disk-size",
			Usage:  "Maximum size of dynamically expanding disk in MB.",
			Value:  defaultDiskSize,
			EnvVar: "HYPERV_DISK_SIZE",
		},
		mcnflag.IntFlag{
			Name:   "hyperv-memory",
			Usage:  "Memory size for host in MB.",
			Value:  defaultMemory,
			EnvVar: "HYPERV_MEMORY",
		},
		mcnflag.IntFlag{
			Name:   "hyperv-cpu-count",
			Usage:  "number of CPUs for the machine",
			Value:  defaultCPU,
			EnvVar: "HYPERV_CPU_COUNT",
		},
		mcnflag.StringFlag{
			Name:   "hyperv-static-macaddress",
			Usage:  "Hyper-V network adapter's static MAC address.",
			EnvVar: "HYPERV_STATIC_MACADDRESS",
		},
		mcnflag.IntFlag{
			Name:   "hyperv-vlan-id",
			Usage:  "Hyper-V network adapter's VLAN ID if any",
			Value:  defaultVLanID,
			EnvVar: "HYPERV_VLAN_ID",
		},
		mcnflag.BoolFlag{
			Name:   "hyperv-disable-dynamic-memory",
			Usage:  "Disable dynamic memory management setting",
			EnvVar: "HYPERV_DISABLE_DYNAMIC_MEMORY",
		},
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Boot2DockerURL = flags.String("hyperv-boot2docker-url")
	d.VSwitch = flags.String("hyperv-virtual-switch")
	d.DiskSize = flags.Int("hyperv-disk-size")
	d.MemSize = flags.Int("hyperv-memory")
	d.CPU = flags.Int("hyperv-cpu-count")
	d.MacAddr = flags.String("hyperv-static-macaddress")
	d.VLanID = flags.Int("hyperv-vlan-id")
	d.SSHUser = "docker"
	d.DisableDynamicMemory = flags.Bool("hyperv-disable-dynamic-memory")
	d.SetSwarmConfigFromFlags(flags)

	return nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "hyperv"
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	stdout, err := cmdOut("(", "Hyper-V\\Get-VM", d.MachineName, ").state")
	if err != nil {
		return state.None, errors.New("Failed to find the VM status")
	}

	resp := parseLines(stdout)
	if len(resp) < 1 {
		return state.None, nil
	}

	switch resp[0] {
	case "Running":
		return state.Running, nil
	case "Off":
		return state.Stopped, nil
	default:
		return state.None, nil
	}
}

// PreCreateCheck checks that the machine creation process can be started safely.
func (d *Driver) PreCreateCheck() error {
	// Check that powershell was found
	if powershell == "" {
		return ErrPowerShellNotFound
	}

	// Check that hyperv is installed
	if err := hypervAvailable(); err != nil {
		return err
	}

	// Check that the user is an Administrator
	isAdmin, err := isAdministrator()
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotAdministrator
	}

	// Check that there is a virtual switch already configured
	if _, err := d.chooseVirtualSwitch(); err != nil {
		return err
	}

	// Downloading boot2docker to cache should be done here to make sure
	// that a download failure will not leave a machine half created.
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	err = b2dutils.UpdateISOCache(d.Boot2DockerURL)
	return err
}

func (d *Driver) Create() error {
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")
	if d.VSwitch == "" {
		defaultVSwitch, err := d.chooseVirtualSwitch()
		if err != nil {
			return err
		}
		d.VSwitch = defaultVSwitch
	}
	log.Infof("Using switch %q", d.VSwitch)

	diskImage, err := d.generateDiskImage()
	if err != nil {
		return err
	}

	if err := cmd("Hyper-V\\New-VM",
		d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.ResolveStorePath(".")),
		"-SwitchName", quote(d.VSwitch),
		"-MemoryStartupBytes", toMb(d.MemSize)); err != nil {
		return err
	}
	if d.DisableDynamicMemory {
		if err := cmd("Hyper-V\\Set-VMMemory",
			"-VMName", d.MachineName,
			"-DynamicMemoryEnabled", "$false"); err != nil {
			return err
		}
	}

	if d.CPU > 1 {
		if err := cmd("Hyper-V\\Set-VMProcessor",
			d.MachineName,
			"-Count", fmt.Sprintf("%d", d.CPU)); err != nil {
			return err
		}
	}

	if d.MacAddr != "" {
		if err := cmd("Hyper-V\\Set-VMNetworkAdapter",
			"-VMName", d.MachineName,
			"-StaticMacAddress", fmt.Sprintf("\"%s\"", d.MacAddr)); err != nil {
			return err
		}
	}

	if d.VLanID > 0 {
		if err := cmd("Hyper-V\\Set-VMNetworkAdapterVlan",
			"-VMName", d.MachineName,
			"-Access",
			"-VlanId", fmt.Sprintf("%d", d.VLanID)); err != nil {
			return err
		}
	}

	if err := cmd("Hyper-V\\Set-VMDvdDrive",
		"-VMName", d.MachineName,
		"-Path", quote(d.ResolveStorePath("boot2docker.iso"))); err != nil {
		return err
	}

	if err := cmd("Hyper-V\\Add-VMHardDiskDrive",
		"-VMName", d.MachineName,
		"-Path", quote(diskImage)); err != nil {
		return err
	}

	log.Infof("Starting VM...")
	return d.Start()
}

func (d *Driver) chooseVirtualSwitch() (string, error) {
	type Switch struct {
		ID         string
		Name       string
		SwitchType int
	}

	getHyperVSwitches := func(filters []string) ([]Switch, error) {
		cmd := []string{"Hyper-V\\Get-VMSwitch", "Select Id, Name, SwitchType"}
		cmd = append(cmd, filters...)
		stdout, err := cmdOut(fmt.Sprintf("[Console]::OutputEncoding = [Text.Encoding]::UTF8; ConvertTo-Json @(%s)", strings.Join(cmd, "|")))
		if err != nil {
			return nil, err
		}

		var switches []Switch
		err = json.Unmarshal([]byte(strings.NewReplacer("\r", "").Replace(stdout)), &switches)
		if err != nil {
			return nil, err
		}

		return switches, nil
	}

	if d.VSwitch == "" {
		// prefer Default Switch over external switches
		switches, err := getHyperVSwitches([]string{fmt.Sprintf("Where-Object {($_.SwitchType -eq 'External') -or ($_.Id -eq '%s')}", defaultSwitchID), "Sort-Object -Property SwitchType"})
		if err != nil {
			return "", errors.New("unable to get available hyperv switches")
		}

		if len(switches) < 1 {
			return "", errors.New("no External vswitch nor Default Switch found. A valid vswitch must be available for this command to run. Check https://docs.docker.com/machine/drivers/hyper-v/")
		}

		return switches[0].Name, nil
	}

	// prefer external switches (using descending order)
	switches, err := getHyperVSwitches([]string{fmt.Sprintf("Where-Object {$_.Name -eq '%s'}", d.VSwitch), "Sort-Object -Property SwitchType -Descending"})
	if err != nil {
		return "", errors.New("unable to get available hyperv switches")
	}

	if len(switches) < 1 {
		return "", fmt.Errorf("vswitch %q not found", d.VSwitch)
	}

	return switches[0].Name, nil
}

// waitForIP waits until the host has a valid IP
func (d *Driver) waitForIP() string {
	log.Infof("Waiting for host to start...")

	for {
		ip, _ := d.GetIP()
		if ip != "" {
			return ip
		}

		time.Sleep(1 * time.Second)
	}
}

// waitStopped waits until the host is stopped
func (d *Driver) waitStopped() error {
	log.Infof("Waiting for host to stop...")

	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}

		if s != state.Running {
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}

// Start starts an host
func (d *Driver) Start() error {
	if err := cmd("Hyper-V\\Start-VM", d.MachineName); err != nil {
		return err
	}

	ip := d.waitForIP()

	d.IPAddress = ip

	return nil
}

// Stop stops an host
func (d *Driver) Stop() error {
	if err := cmd("Hyper-V\\Stop-VM", d.MachineName); err != nil {
		return err
	}

	if err := d.waitStopped(); err != nil {
		return err
	}

	d.IPAddress = ""

	return nil
}

// Remove removes an host
func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}

	return cmd("Hyper-V\\Remove-VM", d.MachineName, "-Force")
}

// Restart stops and starts an host
func (d *Driver) Restart() error {
	err := d.Stop()
	if err != nil {
		return err
	}

	return d.Start()
}

// Kill force stops an host
func (d *Driver) Kill() error {
	if err := cmd("Hyper-V\\Stop-VM", d.MachineName, "-TurnOff"); err != nil {
		return err
	}

	if err := d.waitStopped(); err != nil {
		return err
	}

	d.IPAddress = ""

	return nil
}

func (d *Driver) GetIP() (string, error) {
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	stdout, err := cmdOut("((", "Hyper-V\\Get-VM", d.MachineName, ").networkadapters[0]).ipaddresses[0]")
	if err != nil {
		return "", err
	}

	resp := parseLines(stdout)
	if len(resp) < 1 {
		return "", errors.New("IP not found")
	}

	return resp[0], nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// generateDiskImage creates a small fixed vhd, put the tar in, convert to dynamic, then resize
func (d *Driver) generateDiskImage() (string, error) {
	diskImage := d.ResolveStorePath("disk.vhd")
	fixed := d.ResolveStorePath("fixed.vhd")

	// Resizing vhds requires administrator privileges
	// incase the user is only a hyper-v admin then create the disk at the target size to avoid resizing.
	isWindowsAdmin, err := isWindowsAdministrator()
	if err != nil {
		return "", err
	}
	fixedDiskSize := "10MB"
	if !isWindowsAdmin {
		fixedDiskSize = toMb(d.DiskSize)
	}

	log.Infof("Creating VHD")
	if err := cmd("Hyper-V\\New-VHD", "-Path", quote(fixed), "-SizeBytes", fixedDiskSize, "-Fixed"); err != nil {
		return "", err
	}

	tarBuf, err := mcnutils.MakeDiskImage(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	file, err := os.OpenFile(fixed, os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}
	_, err = file.Write(tarBuf.Bytes())
	if err != nil {
		return "", err
	}
	err = file.Close()
	if err != nil {
		return "", err
	}

	if err := cmd("Hyper-V\\Convert-VHD", "-Path", quote(fixed), "-DestinationPath", quote(diskImage), "-VHDType", "Dynamic", "-DeleteSource"); err != nil {
		return "", err
	}

	if isWindowsAdmin {
		if err := cmd("Hyper-V\\Resize-VHD", "-Path", quote(diskImage), "-SizeBytes", toMb(d.DiskSize)); err != nil {
			return "", err
		}
	}

	return diskImage, nil
}
