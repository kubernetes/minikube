/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarefusion

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"

	"errors"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	cryptossh "golang.org/x/crypto/ssh"
)

const (
	B2DUser        = "docker"
	B2DPass        = "tcuser"
	isoFilename    = "boot2docker.iso"
	isoConfigDrive = "configdrive.iso"
)

// Driver for VMware Fusion
type Driver struct {
	*drivers.BaseDriver
	Memory         int
	DiskSize       int
	CPU            int
	ISO            string
	Boot2DockerURL string

	SSHPassword    string
	ConfigDriveISO string
	ConfigDriveURL string
	NoShare        bool
}

const (
	defaultSSHUser  = B2DUser
	defaultSSHPass  = B2DPass
	defaultDiskSize = 20000
	defaultCPU      = 1
	defaultMemory   = 1024
)

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "FUSION_BOOT2DOCKER_URL",
			Name:   "vmwarefusion-boot2docker-url",
			Usage:  "Fusion URL for boot2docker image",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "FUSION_CONFIGDRIVE_URL",
			Name:   "vmwarefusion-configdrive-url",
			Usage:  "Fusion URL for cloud-init configdrive",
			Value:  "",
		},
		mcnflag.IntFlag{
			EnvVar: "FUSION_CPU_COUNT",
			Name:   "vmwarefusion-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  defaultCPU,
		},
		mcnflag.IntFlag{
			EnvVar: "FUSION_MEMORY_SIZE",
			Name:   "vmwarefusion-memory-size",
			Usage:  "Fusion size of memory for host VM (in MB)",
			Value:  defaultMemory,
		},
		mcnflag.IntFlag{
			EnvVar: "FUSION_DISK_SIZE",
			Name:   "vmwarefusion-disk-size",
			Usage:  "Fusion size of disk for host VM (in MB)",
			Value:  defaultDiskSize,
		},
		mcnflag.StringFlag{
			EnvVar: "FUSION_SSH_USER",
			Name:   "vmwarefusion-ssh-user",
			Usage:  "SSH user",
			Value:  defaultSSHUser,
		},
		mcnflag.StringFlag{
			EnvVar: "FUSION_SSH_PASSWORD",
			Name:   "vmwarefusion-ssh-password",
			Usage:  "SSH password",
			Value:  defaultSSHPass,
		},
		mcnflag.BoolFlag{
			EnvVar: "FUSION_NO_SHARE",
			Name:   "vmwarefusion-no-share",
			Usage:  "Disable the mount of your home directory",
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		CPU:         defaultCPU,
		Memory:      defaultMemory,
		DiskSize:    defaultDiskSize,
		SSHPassword: defaultSSHPass,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
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
	return "vmwarefusion"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	if drivers.EngineInstallURLFlagSet(flags) {
		return errors.New("--engine-install-url cannot be used with the vmwarefusion driver, use --vmwarefusion-boot2docker-url instead")
	}
	d.Memory = flags.Int("vmwarefusion-memory-size")
	d.CPU = flags.Int("vmwarefusion-cpu-count")
	d.DiskSize = flags.Int("vmwarefusion-disk-size")
	d.Boot2DockerURL = flags.String("vmwarefusion-boot2docker-url")
	d.ConfigDriveURL = flags.String("vmwarefusion-configdrive-url")
	d.ISO = d.ResolveStorePath(isoFilename)
	d.ConfigDriveISO = d.ResolveStorePath(isoConfigDrive)
	d.SetSwarmConfigFromFlags(flags)
	d.SSHUser = flags.String("vmwarefusion-ssh-user")
	d.SSHPassword = flags.String("vmwarefusion-ssh-password")
	d.SSHPort = 22
	d.NoShare = flags.Bool("vmwarefusion-no-share")

	// We support a maximum of 16 cpu to be consistent with Virtual Hardware 10
	// specs.
	if d.CPU < 1 {
		d.CPU = int(runtime.NumCPU())
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

func (d *Driver) GetIP() (string, error) {
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	// determine MAC address for VM
	macaddr, err := d.getMacAddressFromVmx()
	if err != nil {
		return "", err
	}

	// attempt to find the address in the vmnet configuration
	if ip, err := d.getIPfromVmnetConfiguration(macaddr); err == nil {
		return ip, err
	}

	// address not found in vmnet so look for a DHCP lease
	ip, err := d.getIPfromDHCPLease(macaddr)
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
	if err := b2dutils.UpdateISOCache(d.Boot2DockerURL); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Create() error {
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
	vmxt.Execute(vmxfile, d)

	// Generate vmdk file
	diskImg := d.ResolveStorePath(fmt.Sprintf("%s.vmdk", d.MachineName))
	if _, err := os.Stat(diskImg); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := vdiskmanager(diskImg, d.DiskSize); err != nil {
			return err
		}
	}

	log.Infof("Starting %s...", d.MachineName)
	vmrun("start", d.vmxPath(), "nogui")

	var ip string

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
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, 22), time.Duration(2*time.Second))
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

	// we got an IP, let's copy ssh keys over
	d.IPAddress = ip

	// Do not execute the rest of boot2docker specific configuration
	// The upload of the public ssh key uses a ssh connection,
	// this works without installed vmware client tools
	if d.ConfigDriveURL != "" {
		var keyfh *os.File
		var keycontent []byte

		log.Infof("Copy public SSH key to %s [%s]", d.MachineName, d.IPAddress)

		// create .ssh folder in users home
		if err := executeSSHCommand(fmt.Sprintf("mkdir -p /home/%s/.ssh", d.SSHUser), d); err != nil {
			return err
		}

		// read generated public ssh key
		if keyfh, err = os.Open(d.publicSSHKeyPath()); err != nil {
			return err
		}
		defer keyfh.Close()

		if keycontent, err = ioutil.ReadAll(keyfh); err != nil {
			return err
		}

		// add public ssh key to authorized_keys
		if err := executeSSHCommand(fmt.Sprintf("echo '%s' > /home/%s/.ssh/authorized_keys", string(keycontent), d.SSHUser), d); err != nil {
			return err
		}

		// make it secure
		if err := executeSSHCommand(fmt.Sprintf("chmod 600 /home/%s/.ssh/authorized_keys", d.SSHUser), d); err != nil {
			return err
		}

		log.Debugf("Leaving create sequence early, configdrive found")
		return nil
	}

	// Generate a tar keys bundle
	if err := d.generateKeyBundle(); err != nil {
		return err
	}

	// Test if /var/lib/boot2docker exists
	vmrun("-gu", B2DUser, "-gp", B2DPass, "directoryExistsInGuest", d.vmxPath(), "/var/lib/boot2docker")

	// Copy SSH keys bundle
	vmrun("-gu", B2DUser, "-gp", B2DPass, "CopyFileFromHostToGuest", d.vmxPath(), d.ResolveStorePath("userdata.tar"), "/home/docker/userdata.tar")

	// Expand tar file.
	vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", "sudo /bin/mv /home/docker/userdata.tar /var/lib/boot2docker/userdata.tar && sudo tar xf /var/lib/boot2docker/userdata.tar -C /home/docker/ > /var/log/userdata.log 2>&1 && sudo chown -R docker:staff /home/docker")

	// Enable Shared Folders
	vmrun("-gu", B2DUser, "-gp", B2DPass, "enableSharedFolders", d.vmxPath())

	var shareName, shareDir string // TODO configurable at some point
	switch runtime.GOOS {
	case "darwin":
		shareName = "Users"
		shareDir = "/Users"
		// TODO "linux" and "windows"
	}

	if shareDir != "" && !d.NoShare {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			// add shared folder, create mountpoint and mount it.
			vmrun("-gu", B2DUser, "-gp", B2DPass, "addSharedFolder", d.vmxPath(), shareName, shareDir)
			command := "[ ! -d " + shareDir + " ]&& sudo mkdir " + shareDir + "; sudo mount --bind /mnt/hgfs/" + shareDir + " " + shareDir + " || [ -f /usr/local/bin/vmhgfs-fuse ]&& sudo /usr/local/bin/vmhgfs-fuse -o allow_other .host:/" + shareName + " " + shareDir + " || sudo mount -t vmhgfs -o uid=$(id -u),gid=$(id -g) .host:/" + shareName + " " + shareDir
			vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", command)
		}
	}
	return nil
}

func (d *Driver) Start() error {
	vmrun("start", d.vmxPath(), "nogui")

	// Do not execute the rest of boot2docker specific configuration, exit here
	if d.ConfigDriveURL != "" {
		log.Debugf("Leaving start sequence early, configdrive found")
		return nil
	}

	log.Debugf("Mounting Shared Folders...")
	var shareName, shareDir string // TODO configurable at some point
	switch runtime.GOOS {
	case "darwin":
		shareName = "Users"
		shareDir = "/Users"
		// TODO "linux" and "windows"
	}

	if shareDir != "" {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			// create mountpoint and mount shared folder
			command := "[ ! -d " + shareDir + " ]&& sudo mkdir " + shareDir + "; sudo mount --bind /mnt/hgfs/" + shareDir + " " + shareDir + " || [ -f /usr/local/bin/vmhgfs-fuse ]&& sudo /usr/local/bin/vmhgfs-fuse -o allow_other .host:/" + shareName + " " + shareDir + " || sudo mount -t vmhgfs -o uid=$(id -u),gid=$(id -g) .host:/" + shareName + " " + shareDir
			vmrun("-gu", B2DUser, "-gp", B2DPass, "runScriptInGuest", d.vmxPath(), "/bin/sh", command)
		}
	}

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
	if err := d.Start(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Kill() error {
	_, _, err := vmrun("stop", d.vmxPath(), "hard nogui")
	return err
}

func (d *Driver) Remove() error {
	s, _ := d.GetState()
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return fmt.Errorf("Error stopping VM before deletion")
		}
	}
	log.Infof("Deleting %s...", d.MachineName)
	vmrun("deleteVM", d.vmxPath(), "nogui")
	return nil
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("VMware Fusion does not currently support the upgrade operation")
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

	if vmxcontent, err = ioutil.ReadAll(vmxfh); err != nil {
		return "", err
	}

	// Look for generatedAddress as we're passing a VMX with addressType = "generated".
	var macaddr string
	vmxparse := regexp.MustCompile(`^ethernet0.generatedAddress\s*=\s*"(.*?)"\s*$`)
	for _, line := range strings.Split(string(vmxcontent), "\n") {
		if matches := vmxparse.FindStringSubmatch(line); matches == nil {
			continue
		} else {
			macaddr = strings.ToLower(matches[1])
		}
	}

	if macaddr == "" {
		return "", fmt.Errorf("couldn't find MAC address in VMX file %s", d.vmxPath())
	}

	log.Debugf("MAC address in VMX: %s", macaddr)

	return macaddr, nil
}

func (d *Driver) getIPfromVmnetConfiguration(macaddr string) (string, error) {

	// DHCP lease table for NAT vmnet interface
	confFiles, _ := filepath.Glob("/Library/Preferences/VMware Fusion/vmnet*/dhcpd.conf")
	for _, conffile := range confFiles {
		log.Debugf("Trying to find IP address in configuration file: %s", conffile)
		if ipaddr, err := d.getIPfromVmnetConfigurationFile(conffile, macaddr); err == nil {
			return ipaddr, err
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

	if confcontent, err = ioutil.ReadAll(conffh); err != nil {
		return "", err
	}

	// find all occurences of 'host .* { .. }' and extract
	// out of the inner block the MAC and IP addresses

	// key = MAC, value = IP
	m := make(map[string]string)

	// Begin of a host block, that contains the IP, MAC
	hostbegin := regexp.MustCompile(`^host (.+?) {`)
	// End of a host block
	hostend := regexp.MustCompile(`^}`)

	// Get the IP address.
	ip := regexp.MustCompile(`^\s*fixed-address (.+?);$`)
	// Get the MAC address associated.
	mac := regexp.MustCompile(`^\s*hardware ethernet (.+?);$`)

	// we use a block depth so that just in case inner blocks exists
	// we are not being fooled by them
	blockdepth := 0
	for _, line := range strings.Split(string(confcontent), "\n") {

		if matches := hostbegin.FindStringSubmatch(line); matches != nil {
			blockdepth = blockdepth + 1
			continue
		}

		// we are only in interested in endings if we in a block. Otherwise we will count
		// ending of non host blocks as well
		if matches := hostend.FindStringSubmatch(line); blockdepth > 0 && matches != nil {
			blockdepth = blockdepth - 1

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
	leasesFiles, _ := filepath.Glob("/var/db/vmware/*.leases")
	for _, dhcpfile := range leasesFiles {
		log.Debugf("Trying to find IP address in leases file: %s", dhcpfile)
		if ipaddr, err := d.getIPfromDHCPLeaseFile(dhcpfile, macaddr); err == nil {
			return ipaddr, err
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

	if dhcpcontent, err = ioutil.ReadAll(dhcpfh); err != nil {
		return "", err
	}

	// Get the IP from the lease table.
	leaseip := regexp.MustCompile(`^lease (.+?) {$`)
	// Get the lease end date time.
	leaseend := regexp.MustCompile(`^\s*ends \d (.+?);$`)
	// Get the MAC address associated.
	leasemac := regexp.MustCompile(`^\s*hardware ethernet (.+?);$`)

	for _, line := range strings.Split(string(dhcpcontent), "\n") {

		if matches := leaseip.FindStringSubmatch(line); matches != nil {
			lastipmatch = matches[1]
			continue
		}

		if matches := leaseend.FindStringSubmatch(line); matches != nil {
			lastleaseendtime, _ = time.Parse("2006/01/02 15:04:05", matches[1])
			continue
		}

		if matches := leasemac.FindStringSubmatch(line); matches != nil && matches[1] == macaddr && currentleadeendtime.Before(lastleaseendtime) {
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

// Make a boot2docker userdata.tar key bundle
func (d *Driver) generateKeyBundle() error {
	log.Debugf("Creating Tar key bundle...")

	magicString := "boot2docker, this is vmware speaking"

	tf, err := os.Create(d.ResolveStorePath("userdata.tar"))
	if err != nil {
		return err
	}
	defer tf.Close()
	var fileWriter = tf

	tw := tar.NewWriter(fileWriter)
	defer tw.Close()

	// magicString first so we can figure out who originally wrote the tar.
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
	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}

	return nil

}

// execute command over SSH with user / password authentication
func executeSSHCommand(command string, d *Driver) error {
	log.Debugf("Execute executeSSHCommand: %s", command)

	config := &cryptossh.ClientConfig{
		User: d.SSHUser,
		Auth: []cryptossh.AuthMethod{
			cryptossh.Password(d.SSHPassword),
		},
	}

	client, err := cryptossh.Dial("tcp", fmt.Sprintf("%s:%d", d.IPAddress, d.SSHPort), config)
	if err != nil {
		log.Debugf("Failed to dial:", err)
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		log.Debugf("Failed to create session: " + err.Error())
		return err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		log.Debugf("Failed to run: " + err.Error())
		return err
	}
	log.Debugf("Stdout from executeSSHCommand: %s", b.String())

	return nil
}
