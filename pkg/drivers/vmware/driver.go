/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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
 * Copyright 2017 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmware

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"errors"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/ssh"
	"k8s.io/minikube/pkg/libmachine/state"
)

const (
	isoFilename    = "boot2docker.iso"
	isoConfigDrive = "configdrive.iso"
)

// Driver for VMware
type Driver struct {
	*Config
}

func NewDriver(hostname, storePath string) drivers.Driver {
	return &Driver{
		Config: NewConfig(hostname, storePath),
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "vmware"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("vmware-memory-size")
	d.CPU = flags.Int("vmware-cpu-count")
	d.DiskSize = flags.Int("vmware-disk-size")
	d.Boot2DockerURL = flags.String("vmware-boot2docker-url")
	d.ConfigDriveURL = flags.String("vmware-configdrive-url")
	d.ISO = d.ResolveStorePath(isoFilename)
	d.ConfigDriveISO = d.ResolveStorePath(isoConfigDrive)
	d.SetSwarmConfigFromFlags(flags)
	d.SSHUser = flags.String("vmware-ssh-user")
	d.SSHPassword = flags.String("vmware-ssh-password")
	d.SSHPort = 22
	d.NoShare = flags.Bool("vmware-no-share")
	d.WaitIP = flags.Int("vmware-wait-ip")
	d.NetworkType = flags.String("vmware-network-type")

	// We support a maximum of 16 cpu to be consistent with Virtual Hardware 10
	// specs.
	if d.CPU < 1 {
		d.CPU = runtime.NumCPU()
	}
	if d.CPU > 16 {
		d.CPU = 16
	}

	return nil
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

func (d *Driver) GetIP() (ip string, err error) {
	s, err := d.GetState()
	if err != nil {
		return "", err
	}

	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	// attempt to find the address from vmrun
	if ip, err := d.getIPfromVmrun(); err == nil {
		return ip, nil
	}

	// determine MAC address for VM
	macaddr, err := d.getMacAddressFromVmx()
	if err != nil {
		return "", err
	}

	// attempt to find the address in the vmnet configuration
	if ip, err = d.getIPfromVmnetConfiguration(macaddr); err == nil {
		return ip, nil
	}

	// address not found in vmnet so look for a DHCP lease
	ip, err = d.getIPfromDHCPLease(macaddr)
	if err != nil {
		return "", err
	}

	return ip, nil
}

func (d *Driver) GetState() (state.State, error) {
	// VMRUN only tells use if the vm is running or not
	vmxp, err := filepath.EvalSymlinks(d.vmxPath())
	if err != nil {
		return state.Error, err
	}

	if stdout, _, _ := vmrun("list"); strings.Contains(stdout, vmxp) {
		return state.Running, nil
	}
	return state.Stopped, nil
}

// PreCreateCheck checks that the machine creation process can be started safely.
func (d *Driver) PreCreateCheck() error {
	// Downloading boot2docker to cache should be done here to make sure
	// that a download failure will not leave a machine half created.
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	return b2dutils.UpdateISOCache(d.Boot2DockerURL)
}

func (d *Driver) Create() error {
	if err := os.MkdirAll(filepath.Join(d.StorePath, "machines", d.GetMachineName()), 0755); err != nil {
		return err
	}

	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	// download cloud-init config drive
	if d.ConfigDriveURL != "" {
		if err := b2dutils.DownloadISO(d.ResolveStorePath("."), isoConfigDrive, d.ConfigDriveURL); err != nil {
			return err
		}
	}

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")
	if err := os.MkdirAll(d.ResolveStorePath("."), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(d.vmxPath()); err == nil {
		return ErrMachineExist
	}

	// Generate vmx config file from template
	vmxt := template.Must(template.New("vmx").Parse(vmx))
	vmxfile, err := os.Create(d.vmxPath())
	if err != nil {
		return err
	}
	if err = vmxt.Execute(vmxfile, d); err != nil {
		return err
	}

	// Generate vmdk file
	diskImg := d.vmdkPath()
	if _, err := os.Stat(diskImg); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := d.generateDiskImage(); err != nil {
			return err
		}
	}

	return d.Start()
}

func (d *Driver) generateDiskImage() error {
	diskImg := d.vmdkPath()

	log.Infof("Creating %d MB hard disk image at %s...", d.DiskSize, diskImg)

	magicString := "boot2docker, please format-me"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// magicString first so the automount script knows to format the disk
	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(magicString)); err != nil {
		return err
	}
	// .ssh/key.pub => authorized_keys
	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	pubKey, err := os.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write(pubKey); err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write(pubKey); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	// we create a 1 MB temporary preallocated disk
	// this will create 2 files:
	// - ${name}-tmp.vmdk - a text file containing disk metadata
	// - ${name}-tmp-flat.vmdk - disk raw data, initially filled with zeroes
	// where ${name}.vmdk is the expected disk filename
	tmpDiskPath := strings.Replace(diskImg, ".vmdk", "-tmp.vmdk", 1)
	err = createDisk(tmpDiskPath, 1, diskTypePreallocated)
	if err != nil {
		return err
	}

	// we write the tar stream at the beginning of the temporary disk raw data
	tmpFlatPath := strings.Replace(tmpDiskPath, "-tmp.vmdk", "-tmp-flat.vmdk", 1)
	f, err := os.OpenFile(tmpFlatPath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, err = f.WriteAt(buf.Bytes(), 0)
	f.Close()
	if err != nil {
		return err
	}

	// we convert the temporary disk to a single, growable expected disk
	err = convertDisk(tmpDiskPath, diskImg, diskTypeGrowable)
	if err != nil {
		return err
	}
	// and grow it to the expected size
	err = growDisk(diskImg, d.DiskSize)
	if err != nil {
		return err
	}

	// finally, we clean up the temporary disk
	err = os.Remove(tmpFlatPath)
	if err != nil {
		return err
	}
	err = os.Remove(tmpDiskPath)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Start() error {
	var ip string
	var err error

	log.Infof("Starting %s...", d.MachineName)
	_, _, err = vmrun("start", d.vmxPath(), "nogui")
	if err != nil {
		return err
	}

	log.Infof("Waiting for VM to come online...")
	for i := 1; i <= 60; i++ {
		ip, err = d.GetIP()
		if err != nil {
			log.Debugf("Not there yet %d/%d, error: %s", i, 60, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if ip != "" {
			log.Debugf("Got an ip: %s", ip)
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, 22), 2*time.Second)
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
		return errors.New("machine didn't return an IP after 120 seconds, aborting")
	}

	// we got an IP
	d.IPAddress = ip

	return nil
}

func (d *Driver) Stop() error {
	_, _, err := vmrun("stop", d.vmxPath(), "nogui")
	return err
}

func (d *Driver) Restart() error {
	// Stop VM gracefully
	if err := d.Stop(); err != nil {
		return err
	}
	// Start it again and mount shared folder
	return d.Start()
}

func (d *Driver) Kill() error {
	_, _, err := vmrun("stop", d.vmxPath(), "hard nogui")
	return err
}

func (d *Driver) Remove() error {
	s, _ := d.GetState()
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return errors.New("error stopping VM before deletion")
		}
	}
	log.Infof("Deleting %s...", d.MachineName)
	if _, _, err := vmrun("deleteVM", d.vmxPath(), "nogui"); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Upgrade() error {
	return errors.New("VMware does not currently support the upgrade operation")
}

func (d *Driver) vmxPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.vmx", d.MachineName))
}

func (d *Driver) vmdkPath() string {
	return d.ResolveStorePath(fmt.Sprintf("%s.vmdk", d.MachineName))
}

func (d *Driver) getMacAddressFromVmx() (string, error) {
	var vmxfh *os.File
	var vmxcontent []byte
	var err error

	if vmxfh, err = os.Open(d.vmxPath()); err != nil {
		return "", err
	}
	defer vmxfh.Close()

	if vmxcontent, err = io.ReadAll(vmxfh); err != nil {
		return "", err
	}

	// Look for generatedAddress as we're passing a VMX with addressType = "generated".
	var macaddr string
	vmxparse := regexp.MustCompile(`^ethernet0.generatedAddress\s*=\s*"(.*?)"\s*$`)
	for _, line := range strings.Split(string(vmxcontent), "\n") {
		matches := vmxparse.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		macaddr = strings.ToLower(matches[1])
	}

	if macaddr == "" {
		return "", fmt.Errorf("couldn't find MAC address in VMX file %s", d.vmxPath())
	}

	log.Debugf("MAC address in VMX: %s", macaddr)

	return macaddr, nil
}

func (d *Driver) getIPfromVmrun() (string, error) {
	vmx := d.vmxPath()

	ip := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
	stdout, _, _ := vmrunWait(time.Duration(d.WaitIP)*time.Millisecond, "getGuestIPAddress", vmx, "-wait")
	if match := ip.FindString(stdout); match != "" {
		return match, nil
	}

	return "", errors.New("could not get IP from vmrun")
}

func (d *Driver) getIPfromVmnetConfiguration(macaddr string) (string, error) {

	// DHCP lease table for NAT vmnet interface
	confFiles, _ := filepath.Glob(DhcpConfigFiles())
	for _, conffile := range confFiles {
		log.Debugf("Trying to find IP address in configuration file: %s", conffile)
		if ipaddr, err := d.getIPfromVmnetConfigurationFile(conffile, macaddr); err == nil {
			return ipaddr, nil
		}
	}

	return "", fmt.Errorf("IP not found for MAC %s in vmnet configuration files", macaddr)
}

func (d *Driver) getIPfromVmnetConfigurationFile(conffile, macaddr string) (string, error) {
	var conffh *os.File
	var confcontent []byte

	var currentip string
	var lastipmatch string
	var lastmacmatch string

	var err error

	if conffh, err = os.Open(conffile); err != nil {
		return "", err
	}
	defer conffh.Close()

	if confcontent, err = io.ReadAll(conffh); err != nil {
		return "", err
	}

	// find all occurrences of 'host .* { .. }' and extract
	// out of the inner block the MAC and IP addresses

	// key = MAC, value = IP
	m := make(map[string]string)

	// Begin of a host block, that contains the IP, MAC
	hostbegin := regexp.MustCompile(`^host (.+?) {`)
	// End of a host block
	hostend := regexp.MustCompile(`^}`)

	// Get the IP address.
	ip := regexp.MustCompile(`^\s*fixed-address (.+?);\r?$`)
	// Get the MAC address associated.
	mac := regexp.MustCompile(`^\s*hardware ethernet (.+?);\r?$`)

	// we use a block depth so that just in case inner blocks exists
	// we are not being fooled by them
	blockdepth := 0
	for _, line := range strings.Split(string(confcontent), "\n") {

		if matches := hostbegin.FindStringSubmatch(line); matches != nil {
			blockdepth++
			continue
		}

		// we are only in interested in endings if we in a block. Otherwise we will count
		// ending of non host blocks as well
		if matches := hostend.FindStringSubmatch(line); blockdepth > 0 && matches != nil {
			blockdepth--

			if blockdepth == 0 {
				// add data
				m[lastmacmatch] = lastipmatch

				// reset all temp var holders
				lastipmatch = ""
				lastmacmatch = ""
			}

			continue
		}

		// only if we are within the first level of a block
		// we are looking for addresses to extract
		if blockdepth == 1 {
			if matches := ip.FindStringSubmatch(line); matches != nil {
				lastipmatch = matches[1]
				continue
			}

			if matches := mac.FindStringSubmatch(line); matches != nil {
				lastmacmatch = strings.ToLower(matches[1])
				continue
			}
		}
	}

	log.Debugf("Following IPs found %s", m)

	// map is filled to now lets check if we have a MAC associated to an IP
	currentip, ok := m[strings.ToLower(macaddr)]

	if !ok {
		return "", fmt.Errorf("IP not found for MAC %s in vmnet configuration", macaddr)
	}

	log.Debugf("IP found in vmnet configuration file: %s", currentip)

	return currentip, nil

}

func (d *Driver) getIPfromDHCPLease(macaddr string) (string, error) {

	// DHCP lease table for NAT vmnet interface
	leasesFiles, _ := filepath.Glob(DhcpLeaseFiles())
	for _, dhcpfile := range leasesFiles {
		log.Debugf("Trying to find IP address in leases file: %s", dhcpfile)
		if ipaddr, err := d.getIPfromDHCPLeaseFile(dhcpfile, macaddr); err == nil {
			return ipaddr, nil
		}
	}

	return "", fmt.Errorf("IP not found for MAC %s in DHCP leases", macaddr)
}

func (d *Driver) getIPfromDHCPLeaseFile(dhcpfile, macaddr string) (string, error) {
	var dhcpfh *os.File
	var dhcpcontent []byte
	var lastipmatch string
	var currentip string
	var lastleaseendtime time.Time
	var currentleadeendtime time.Time
	var err error

	if dhcpfh, err = os.Open(dhcpfile); err != nil {
		return "", err
	}
	defer dhcpfh.Close()

	if dhcpcontent, err = io.ReadAll(dhcpfh); err != nil {
		return "", err
	}

	// Get the IP from the lease table.
	leaseip := regexp.MustCompile(`^lease (.+?) {\r?$`)
	// Get the lease end date time.
	leaseend := regexp.MustCompile(`^\s*ends \d (.+?);\r?$`)
	// Get the MAC address associated.
	leasemac := regexp.MustCompile(`^\s*hardware ethernet (.+?);\r?$`)

	for _, line := range strings.Split(string(dhcpcontent), "\n") {

		if matches := leaseip.FindStringSubmatch(line); matches != nil {
			lastipmatch = matches[1]
			continue
		}

		if matches := leaseend.FindStringSubmatch(line); matches != nil {
			lastleaseendtime, _ = time.Parse("2006/01/02 15:04:05", matches[1])
			continue
		}

		if matches := leasemac.FindStringSubmatch(line); len(matches) > 0 && matches[1] == macaddr && currentleadeendtime.Before(lastleaseendtime) {
			currentip = lastipmatch
			currentleadeendtime = lastleaseendtime
		}
	}

	if currentip == "" {
		return "", fmt.Errorf("IP not found for MAC %s in DHCP leases", macaddr)
	}

	log.Debugf("IP found in DHCP lease table: %s", currentip)

	return currentip, nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}
