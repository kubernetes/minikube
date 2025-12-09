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

/*
Copyright (c) 2015-2020 Parallels IP Holdings GmbH.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package parallels

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/ssh"
	"k8s.io/minikube/pkg/libmachine/state"
)

const (
	isoFilename                 = "boot2docker.iso"
	shareFolderNamePrefix       = "docker_machine_share_"
	minDiskSize                 = 32
	defaultCPU                  = 1
	defaultMemory               = 1024
	defaultVideoSize            = 64
	defaultBoot2DockerURL       = ""
	defaultShareFolder          = "/Users"
	defaultNoShare              = false
	defaultDiskSize             = 20000
	defaultSSHPort              = 22
	defaultSSHUser              = "docker"
	defaultNestedVirtualization = false
)

var (
	reMachineNotFound  = regexp.MustCompile(`Failed to get VM config: The virtual machine could not be found..*`)
	reParallelsVersion = regexp.MustCompile(`.* (\d+\.\d+\.\d+).*`)
	reParallelsEdition = regexp.MustCompile(`edition="(.+)"`)
	reSharedAdapterIP  = regexp.MustCompile(`\s*IPv4 address:\s*(\d+\.\d+\.\d+\.\d+)`)
	reSharedFolder     = regexp.MustCompile(`\s*(.+) \(\+\) path='(.+)' mode=.+`)

	errMachineNotExist           = errors.New("machine does not exist")
	errSharedNetworkNotConnected = errors.New("Your Mac host is not connected to Shared network. Please, ensure this option is set: 'Parallels Desktop' -> 'Preferences' -> 'Network' -> 'Shared' -> 'Connect Mac to this network'")

	v11, _ = version.NewVersion("11.0.0")
)

// Driver for Parallels Desktop
type Driver struct {
	*drivers.BaseDriver
	CPU                  int
	Memory               int
	VideoSize            int
	DiskSize             int
	Boot2DockerURL       string
	NoShare              bool
	ShareFolders         []string
	NestedVirtualization bool
}

// NewDriver creates a new Parallels Desktop driver with default settings
func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
			SSHUser:     defaultSSHUser,
			SSHPort:     defaultSSHPort,
		},
		CPU:                  defaultCPU,
		Memory:               defaultMemory,
		VideoSize:            defaultVideoSize,
		DiskSize:             defaultDiskSize,
		Boot2DockerURL:       defaultBoot2DockerURL,
		NoShare:              defaultNoShare,
		NestedVirtualization: defaultNestedVirtualization,
	}
}

// Create a host using the driver's config
//
//nolint:gocyclo
func (d *Driver) Create() error {
	var (
		err error
	)

	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err = b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Creating SSH key...")
	sshKeyPath := d.GetSSHKeyPath()
	log.Debugf("SSH key: %s", sshKeyPath)
	if err = ssh.GenerateSSHKey(sshKeyPath); err != nil {
		return err
	}

	log.Infof("Creating Parallels Desktop VM...")

	ver, err := getParallelsVersion()
	if err != nil {
		return err
	}

	absStorePath, _ := filepath.Abs(d.ResolveStorePath("."))
	if err = prlctl("create", d.MachineName,
		"--distribution", "linux",
		"--dst", absStorePath,
		"--no-hdd"); err != nil {
		return err
	}

	cpus := d.CPU
	if cpus < 1 {
		cpus = runtime.NumCPU()
	}
	if cpus > 32 {
		cpus = 32
	}

	videoSize := d.VideoSize
	if videoSize < 2 {
		videoSize = defaultVideoSize
	}

	if err = prlctl("set", d.MachineName,
		"--select-boot-device", "off",
		"--cpus", fmt.Sprintf("%d", cpus),
		"--memsize", fmt.Sprintf("%d", d.Memory),
		"--videosize", fmt.Sprintf("%d", videoSize),
		"--cpu-hotplug", "off",
		"--on-window-close", "keep-running",
		"--longer-battery-life", "on",
		"--3d-accelerate", "off",
		"--device-bootorder", "cdrom0"); err != nil {
		return err
	}

	if d.NestedVirtualization {
		if err = prlctl("set", d.MachineName,
			"--nested-virt", "on"); err != nil {
			return err
		}
	}

	absISOPath, _ := filepath.Abs(d.ResolveStorePath(isoFilename))
	if err = prlctl("set", d.MachineName,
		"--device-set", "cdrom0",
		"--iface", "sata",
		"--position", "0",
		"--image", absISOPath,
		"--connect"); err != nil {
		return err
	}

	initialDiskSize := minDiskSize

	// Fix for [GH-67]. Create a bigger disk on Parallels Desktop 13.0.*
	constraints, _ := version.NewConstraint(">= 13.0.0, < 13.1.0")
	if constraints.Check(ver) {
		initialDiskSize = 1891
	}

	// Create a small plain disk. It will be converted and expanded later
	if err = prlctl("set", d.MachineName,
		"--device-add", "hdd",
		"--iface", "sata",
		"--position", "1",
		"--image", d.diskPath(),
		"--type", "plain",
		"--size", fmt.Sprintf("%d", initialDiskSize)); err != nil {
		return err
	}

	if err = d.generateDiskImage(d.DiskSize); err != nil {
		return err
	}

	// Enable headless mode
	if err = prlctl("set", d.MachineName,
		"--startup-view", "headless"); err != nil {
		return err
	}

	// Don't share any additional folders
	if err = prlctl("set", d.MachineName,
		"--shf-host-defined", "off"); err != nil {
		return err
	}

	// Enable time sync, don't touch timezone (this part is buggy)
	if err = prlctl("set", d.MachineName, "--time-sync", "on"); err != nil {
		return err
	}
	if err = prlctl("set", d.MachineName,
		"--disable-timezone-sync", "on"); err != nil {
		return err
	}

	// Configure Shared Folders
	if err = prlctl("set", d.MachineName,
		"--shf-host", "on",
		"--shared-cloud", "off",
		"--shared-profile", "off",
		"--smart-mount", "off"); err != nil {
		return err
	}

	// Configure Serial Console as a socket
	if err = prlctl("set", d.MachineName,
		"--device-add", "serial",
		"--socket", fmt.Sprintf("parallels-serial-%s", d.MachineName)); err != nil {
		return err
	}

	if !d.NoShare {
		for i, f := range d.ShareFolders {
			// Ensure the path is absolute and is available
			fAbs, err := filepath.Abs(f)
			if err != nil {
				return err
			}
			if _, err := os.Stat(fAbs); err != nil {
				if os.IsNotExist(err) {
					log.Infof("Host path '%s' does not exist. Skipping sharing it with the machine...", fAbs)
					continue
				}
				return err
			}

			if err = prlctl("set", d.MachineName,
				"--shf-host-add", fmt.Sprintf("%s%d", shareFolderNamePrefix, i),
				"--path", fAbs); err != nil {
				return err
			}
		}
	}

	log.Infof("Starting Parallels Desktop VM...")

	// Don't use Start() since it expects to have a dhcp lease already
	if err = prlctl("start", d.MachineName); err != nil {
		return err
	}

	var ip string

	log.Infof("Waiting for VM to come online...")
	for i := 1; i <= 60; i++ {
		ip, err = d.getIPfromDHCPLease()
		if err != nil {
			log.Debugf("Not there yet %d/%d, error: %s", i, 60, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if ip != "" {
			log.Debugf("Got an ip: %s", ip)
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, d.SSHPort), 2*time.Second)
			if err != nil {
				log.Debugf("SSH Daemon not responding yet: %s", err)
				time.Sleep(2 * time.Second)
				continue
			}
			conn.Close()
			break
		}
	}

	if ip == "" {
		return fmt.Errorf("Machine didn't return an IP after 120 seconds, aborting")
	}

	d.IPAddress = ip

	if err := d.Start(); err != nil {
		return err
	}

	return nil
}

// DriverName returns the name of the driver as it is registered
func (d *Driver) DriverName() string {
	return "parallels"
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	// Assume that Parallels Desktop hosts don't have IPs unless they are running
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	ip, err := d.getIPfromDHCPLease()
	if err != nil {
		return "", err
	}

	return ip, nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
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

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	stdout, stderr, err := prlctlOutErr("list", d.MachineName, "--output", "status", "--no-header")
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return state.Error, errMachineNotExist
		}
		return state.Error, err
	}

	switch stdout {
	case "running\n":
		return state.Running, nil
	case "paused\n":
		return state.Paused, nil
	case "suspended\n":
		return state.Saved, nil
	case "stopping\n":
		return state.Stopping, nil
	case "stopped\n":
		return state.Stopped, nil
	}
	return state.None, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return prlctl("stop", d.MachineName, "--kill")
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	// Check platform type
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("Driver \"parallels\" works only on macOS!")
	}

	// Check Parallels Desktop version
	ver, err := getParallelsVersion()
	if err != nil {
		return err
	}

	if ver.LessThan(v11) {
		return fmt.Errorf("Driver \"parallels\" supports only Parallels Desktop 11 and higher. You use: Parallels Desktop %s.", ver)
	}

	// Check Parallels Desktop edition
	edit, err := getParallelsEdition()
	if err != nil {
		return err
	}

	log.Debugf("Found Parallels Desktop version: %d, edition: %s", ver, edit)

	switch edit {
	case "pro", "business":
		break
	default:
		return fmt.Errorf("Docker Machine can be used only with Parallels Desktop Pro or Business edition. You use: %s edition", edit)
	}

	// Check whether the host is connected to Shared network
	if err := checkSharedNetworkConnected(); err != nil {
		return err
	}

	// Downloading boot2docker to cache should be done here to make sure
	// that a download failure will not leave a machine half created.
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.UpdateISOCache(d.Boot2DockerURL); err != nil {
		return err
	}

	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		if err == errMachineNotExist {
			log.Infof("machine does not exist, assuming it has been removed already")
			return nil
		}
		return err
	}
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}
	return prlctl("delete", d.MachineName)
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.IntFlag{
			EnvVar: "PARALLELS_MEMORY_SIZE",
			Name:   "parallels-memory",
			Usage:  "Size of memory for host in MB",
			Value:  defaultMemory,
		},
		mcnflag.IntFlag{
			EnvVar: "PARALLELS_CPU_COUNT",
			Name:   "parallels-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  defaultCPU,
		},
		mcnflag.IntFlag{
			EnvVar: "PARALLELS_VIDEO_SIZE",
			Name:   "parallels-video-size",
			Usage:  "Size of video memory for host in MB",
			Value:  defaultVideoSize,
		},
		mcnflag.IntFlag{
			EnvVar: "PARALLELS_DISK_SIZE",
			Name:   "parallels-disk-size",
			Usage:  "Size of disk for host in MB",
			Value:  defaultDiskSize,
		},
		mcnflag.StringFlag{
			EnvVar: "PARALLELS_BOOT2DOCKER_URL",
			Name:   "parallels-boot2docker-url",
			Usage:  "The URL of the boot2docker image. Defaults to the latest available version",
			Value:  defaultBoot2DockerURL,
		},
		mcnflag.BoolFlag{
			Name:  "parallels-no-share",
			Usage: "Disable the mount of shared folder",
		},
		mcnflag.StringSliceFlag{
			Name:  "parallels-share-folder",
			Usage: "Path to the directory which should be shared with the machine. Default: /Users",
			Value: []string{defaultShareFolder},
		},
		mcnflag.BoolFlag{
			Name:  "parallels-nested-virtualization",
			Usage: "Enable nested virtualization",
		},
	}
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.CPU = opts.Int("parallels-cpu-count")
	d.Memory = opts.Int("parallels-memory")
	d.VideoSize = opts.Int("parallels-video-size")
	d.DiskSize = opts.Int("parallels-disk-size")
	d.Boot2DockerURL = opts.String("parallels-boot2docker-url")
	d.SetSwarmConfigFromFlags(opts)
	d.SSHUser = defaultSSHUser
	d.SSHPort = defaultSSHPort
	d.NoShare = opts.Bool("parallels-no-share")
	d.ShareFolders = opts.StringSlice("parallels-share-folder")
	d.NestedVirtualization = opts.Bool("parallels-nested-virtualization")

	return nil
}

// Start a host
func (d *Driver) Start() error {
	// Check whether the host is connected to Shared network
	if err := checkSharedNetworkConnected(); err != nil {
		return err
	}

	s, err := d.GetState()
	if err != nil {
		return err
	}

	switch s {
	case state.Stopped, state.Saved, state.Paused:
		if err = prlctl("start", d.MachineName); err != nil {
			return err
		}
		log.Infof("Waiting for VM to start...")
	case state.Running:
		break
	default:
		log.Infof("VM not in restartable state")
	}

	if err = drivers.WaitForSSH(d); err != nil {
		return err
	}

	d.IPAddress, err = d.GetIP()
	if err != nil {
		return err
	}

	// Mount Share Folder
	shareFoldersMap, err := d.getShareFolders()
	if err != nil {
		return err
	}

	if !d.NoShare {
		for shareName, sharePath := range shareFoldersMap {
			if err := d.mountShareFolder(shareName, sharePath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	if err := prlctl("stop", d.MachineName); err != nil {
		return err
	}
	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}
		if s == state.Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return nil
}

func (d *Driver) getIPfromDHCPLease() (string, error) {

	DHCPLeaseFile := "/Library/Preferences/Parallels/parallels_dhcp_leases"

	stdout, _, err := prlctlOutErr("list", "-i", d.MachineName)
	if err != nil {
		return "", err
	}
	macRe := regexp.MustCompile("net0.* mac=([0-9A-F]{12}) card=.*")
	macMatch := macRe.FindAllStringSubmatch(stdout, 1)

	if len(macMatch) != 1 {
		return "", fmt.Errorf("MAC address for NIC: nic0 on Virtual Machine: %s not found!\n", d.MachineName)
	}
	mac := macMatch[0][1]

	if len(mac) != 12 {
		return "", fmt.Errorf("Not a valid MAC address: %s. It should be exactly 12 digits.", mac)
	}

	leases, err := ioutil.ReadFile(DHCPLeaseFile)
	if err != nil {
		return "", err
	}

	ipRe := regexp.MustCompile("(.*)=\"(.*),(.*)," + strings.ToLower(mac) + ",.*\"")
	mostRecentIP := ""
	mostRecentLease := uint64(0)
	for _, l := range ipRe.FindAllStringSubmatch(string(leases), -1) {
		ip := l[1]
		expiry, _ := strconv.ParseUint(l[2], 10, 64)
		leaseTime, _ := strconv.ParseUint(l[3], 10, 32)
		log.Debugf("Found lease: %s for MAC: %s, expiring at %d, leased for %d s.\n", ip, mac, expiry, leaseTime)
		if mostRecentLease <= expiry-leaseTime {
			mostRecentIP = ip
			mostRecentLease = expiry - leaseTime
		}
	}

	if len(mostRecentIP) == 0 {
		return "", fmt.Errorf("IP lease not found for MAC address %s in: %s\n", mac, DHCPLeaseFile)
	}
	log.Debugf("Found IP lease: %s for MAC address %s\n", mostRecentIP, mac)

	return mostRecentIP, nil
}

func (d *Driver) diskPath() string {
	absDiskPath, _ := filepath.Abs(d.ResolveStorePath("disk.hdd"))
	return absDiskPath
}

func (d *Driver) getShareFolders() (map[string]string, error) {
	stdout, _, err := prlctlOutErr("list", "--info", d.MachineName)
	if err != nil {
		return nil, err
	}

	// Parse Shared Folder name (ID) and path
	res := make(map[string]string)
	for _, match := range reSharedFolder.FindAllStringSubmatch(stdout, -1) {
		sName := match[1]
		sPath := match[2]
		log.Debugf("Found the configured shared folder. Name: %q, Path: %q\n", sName, sPath)
		res[sName] = sPath
	}
	return res, nil
}

// Mounts shared folder to the specified guest path. It is assumed that host and guest paths are the same
func (d *Driver) mountShareFolder(shareName string, mountPoint string) error {
	// Check the host path is available
	if _, err := os.Stat(mountPoint); err != nil {
		if os.IsNotExist(err) {
			log.Infof("Host path %q does not exist. Skipping mount to VM...", mountPoint)
			return nil
		}
		return err
	}

	// Ensure that the share is available on the guest side
	checkCmd := fmt.Sprintf("sudo modprobe prl_fs && grep -w %q /proc/fs/prl_fs/sf_list", shareName)
	if _, err := drivers.RunSSHCommandFromDriver(d, checkCmd); err != nil {
		log.Infof("Shared folder %q is unavailable. Skipping mount to VM...", shareName)
		return nil
	}

	// Mount the shared folder
	log.Infof("Mounting shared folder %q ...", mountPoint)
	mountCmd := fmt.Sprintf("sudo mkdir -p %q && sudo mount -t prl_fs %q %q", mountPoint, shareName, mountPoint)
	if _, err := drivers.RunSSHCommandFromDriver(d, mountCmd); err != nil {
		return fmt.Errorf("Error mounting shared folder: %s", err)
	}

	return nil
}

// Make a boot2docker VM disk image.
func (d *Driver) generateDiskImage(size int) error {
	tarBuf, err := mcnutils.MakeDiskImage(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	minSizeBytes := int64(minDiskSize) << 20 // usually won't fit in 32-bit int (max 2GB)

	//Expand the initial image if needed
	if bufLen := int64(tarBuf.Len()); bufLen > minSizeBytes {
		bufLenMBytes := bufLen>>20 + 1
		if err = prldisktool("resize",
			"--hdd", d.diskPath(),
			"--size", fmt.Sprintf("%d", bufLenMBytes)); err != nil {
			return err
		}
	}

	// Find hds file
	hdsList, err := filepath.Glob(d.diskPath() + "/*.hds")
	if err != nil {
		return err
	}
	if len(hdsList) == 0 {
		return fmt.Errorf("Could not find *.hds image in %s", d.diskPath())
	}
	hdsPath := hdsList[0]
	log.Debugf("HDS image path: %s", hdsPath)

	// Write tar to the hds file
	hds, err := os.OpenFile(hdsPath, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer hds.Close()
	hds.Seek(0, os.SEEK_SET)
	_, err = hds.Write(tarBuf.Bytes())
	if err != nil {
		return err
	}
	hds.Close()

	// Convert image to expanding type and resize it
	if err := prldisktool("convert", "--expanding", "--merge",
		"--hdd", d.diskPath()); err != nil {
		return err
	}

	if err := prldisktool("resize",
		"--hdd", d.diskPath(),
		"--size", fmt.Sprintf("%d", size)); err != nil {
		return err
	}

	return nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// Detects Parallels Desktop major version
func getParallelsVersion() (*version.Version, error) {
	stdout, _, err := prlctlOutErr("--version")
	if err != nil {
		return nil, err
	}

	// Parse Parallels Desktop version
	verRaw := reParallelsVersion.FindStringSubmatch(stdout)
	if verRaw == nil {
		return nil, fmt.Errorf("Parallels Desktop version could not be fetched: %s", stdout)
	}

	ver, err := version.NewVersion(verRaw[1])
	if err != nil {
		return nil, err
	}

	return ver, nil
}

// Detects Parallels Desktop edition
func getParallelsEdition() (string, error) {
	stdout, _, err := prlsrvctlOutErr("info", "--license")
	if err != nil {
		return "", err
	}

	// Parse Parallels Desktop version
	res := reParallelsEdition.FindStringSubmatch(stdout)
	if res == nil {
		return "", fmt.Errorf("Driver \"parallels\" requires Parallels Desktop license to be activated. More info: https://kb.parallels.com/en/124225")
	}

	return res[1], nil
}

// Checks whether the host is connected to Shared network
func checkSharedNetworkConnected() error {
	stdout, _, err := prlsrvctlOutErr("net", "info", "Shared")
	if err != nil {
		return err
	}

	// Parse the IPv4 of Shared network adapter
	res := reSharedAdapterIP.FindStringSubmatch(stdout)
	if res == nil {
		return errSharedNetworkNotConnected
	}

	sharedNetworkIP := net.ParseIP(res[1])
	log.Debugf("IP address of Shared network adapter: %s", sharedNetworkIP)

	hostAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	log.Debugf("All host interface addresses: %v", hostAddrs)

	// Check if the there is an interface with the Shared network adapter's IP assigned
	for _, netAddr := range hostAddrs {
		ipAddr := netAddr.(*net.IPNet).IP
		if ipAddr.Equal(sharedNetworkIP) {
			log.Debugf("Parallels Shared network adapter is connected")
			return nil
		}
	}

	return errSharedNetworkNotConnected
}
