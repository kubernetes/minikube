package exoscale

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
	"github.com/pyr/egoscale/src/egoscale"
)

type Driver struct {
	*drivers.BaseDriver
	URL              string
	APIKey           string `json:"ApiKey"`
	APISecretKey     string `json:"ApiSecretKey"`
	InstanceProfile  string
	DiskSize         int
	Image            string
	SecurityGroup    string
	AvailabilityZone string
	KeyPair          string
	PublicKey        string
	UserDataFile     string
	ID               string `json:"Id"`
}

const (
	defaultInstanceProfile  = "small"
	defaultDiskSize         = 50
	defaultImage            = "ubuntu-15.10"
	defaultAvailabilityZone = "ch-gva-2"
	defaultSSHUser          = "ubuntu"
)

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_ENDPOINT",
			Name:   "exoscale-url",
			Usage:  "exoscale API endpoint",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_API_KEY",
			Name:   "exoscale-api-key",
			Usage:  "exoscale API key",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_API_SECRET",
			Name:   "exoscale-api-secret-key",
			Usage:  "exoscale API secret key",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_INSTANCE_PROFILE",
			Name:   "exoscale-instance-profile",
			Value:  defaultInstanceProfile,
			Usage:  "exoscale instance profile (small, medium, large, ...)",
		},
		mcnflag.IntFlag{
			EnvVar: "EXOSCALE_DISK_SIZE",
			Name:   "exoscale-disk-size",
			Value:  defaultDiskSize,
			Usage:  "exoscale disk size (10, 50, 100, 200, 400)",
		},
		mcnflag.StringFlag{
			EnvVar: "EXSOCALE_IMAGE",
			Name:   "exoscale-image",
			Value:  defaultImage,
			Usage:  "exoscale image template",
		},
		mcnflag.StringSliceFlag{
			EnvVar: "EXOSCALE_SECURITY_GROUP",
			Name:   "exoscale-security-group",
			Value:  []string{},
			Usage:  "exoscale security group",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_AVAILABILITY_ZONE",
			Name:   "exoscale-availability-zone",
			Value:  defaultAvailabilityZone,
			Usage:  "exoscale availibility zone",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_SSH_USER",
			Name:   "exoscale-ssh-user",
			Value:  defaultSSHUser,
			Usage:  "Set the name of the ssh user",
		},
		mcnflag.StringFlag{
			EnvVar: "EXOSCALE_USERDATA",
			Name:   "exoscale-userdata",
			Usage:  "path to file with cloud-init user-data",
		},
	}
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		InstanceProfile:  defaultInstanceProfile,
		DiskSize:         defaultDiskSize,
		Image:            defaultImage,
		AvailabilityZone: defaultAvailabilityZone,
		BaseDriver: &drivers.BaseDriver{
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
		d.SSHUser = defaultSSHUser
	}

	return d.SSHUser
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "exoscale"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.URL = flags.String("exoscale-url")
	d.APIKey = flags.String("exoscale-api-key")
	d.APISecretKey = flags.String("exoscale-api-secret-key")
	d.InstanceProfile = flags.String("exoscale-instance-profile")
	d.DiskSize = flags.Int("exoscale-disk-size")
	d.Image = flags.String("exoscale-image")
	securityGroups := flags.StringSlice("exoscale-security-group")
	if len(securityGroups) == 0 {
		securityGroups = []string{"docker-machine"}
	}
	d.SecurityGroup = strings.Join(securityGroups, ",")
	d.AvailabilityZone = flags.String("exoscale-availability-zone")
	d.SSHUser = flags.String("exoscale-ssh-user")
	d.UserDataFile = flags.String("exoscale-userdata")
	d.SetSwarmConfigFromFlags(flags)

	if d.URL == "" {
		d.URL = "https://api.exoscale.ch/compute"
	}
	if d.APIKey == "" || d.APISecretKey == "" {
		return fmt.Errorf("Please specify an API key (--exoscale-api-key) and an API secret key (--exoscale-api-secret-key).")
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if d.UserDataFile != "" {
		if _, err := os.Stat(d.UserDataFile); os.IsNotExist(err) {
			return fmt.Errorf("user-data file %s could not be found", d.UserDataFile)
		}
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	client := egoscale.NewClient(d.URL, d.APIKey, d.APISecretKey)
	vm, err := client.GetVirtualMachine(d.ID)
	if err != nil {
		return state.Error, err
	}
	switch vm.State {
	case "Starting":
		return state.Starting, nil
	case "Running":
		return state.Running, nil
	case "Stopping":
		return state.Running, nil
	case "Stopped":
		return state.Stopped, nil
	case "Destroyed":
		return state.Stopped, nil
	case "Expunging":
		return state.Stopped, nil
	case "Migrating":
		return state.Paused, nil
	case "Error":
		return state.Error, nil
	case "Unknown":
		return state.Error, nil
	case "Shutdowned":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) createDefaultSecurityGroup(client *egoscale.Client, group string) (string, error) {
	rules := []egoscale.SecurityGroupRule{
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "TCP",
			Port:            22,
		},
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "TCP",
			Port:            2376,
		},
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "TCP",
			Port:            3376,
		},
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "ICMP",
			IcmpType:        8,
			IcmpCode:        0,
		},
	}
	sgresp, err := client.CreateSecurityGroupWithRules(
		group,
		rules,
		make([]egoscale.SecurityGroupRule, 0, 0))
	if err != nil {
		return "", err
	}
	sg := sgresp.Id
	return sg, nil
}

func (d *Driver) Create() error {
	userdata, err := d.getCloudInit()
	if err != nil {
		return err
	}

	log.Infof("Querying exoscale for the requested parameters...")
	client := egoscale.NewClient(d.URL, d.APIKey, d.APISecretKey)
	topology, err := client.GetTopology()
	if err != nil {
		return err
	}

	// Availability zone UUID
	zone, ok := topology.Zones[d.AvailabilityZone]
	if !ok {
		return fmt.Errorf("Availability zone %v doesn't exist",
			d.AvailabilityZone)
	}
	log.Debugf("Availability zone %v = %s", d.AvailabilityZone, zone)

	// Image UUID
	var tpl string
	images, ok := topology.Images[strings.ToLower(d.Image)]
	if ok {
		tpl, ok = images[d.DiskSize]
	}
	if !ok {
		return fmt.Errorf("Unable to find image %v with size %d",
			d.Image, d.DiskSize)
	}
	log.Debugf("Image %v(%d) = %s", d.Image, d.DiskSize, tpl)

	// Profile UUID
	profile, ok := topology.Profiles[strings.ToLower(d.InstanceProfile)]
	if !ok {
		return fmt.Errorf("Unable to find the %s profile",
			d.InstanceProfile)
	}
	log.Debugf("Profile %v = %s", d.InstanceProfile, profile)

	// Security groups
	securityGroups := strings.Split(d.SecurityGroup, ",")
	sgs := make([]string, len(securityGroups))
	for idx, group := range securityGroups {
		sg, ok := topology.SecurityGroups[group]
		if !ok {
			log.Infof("Security group %v does not exist, create it",
				group)
			sg, err = d.createDefaultSecurityGroup(client, group)
			if err != nil {
				return err
			}
		}
		log.Debugf("Security group %v = %s", group, sg)
		sgs[idx] = sg
	}

	log.Infof("Generate an SSH keypair...")
	keypairName := fmt.Sprintf("docker-machine-%s", d.MachineName)
	kpresp, err := client.CreateKeypair(keypairName)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(d.GetSSHKeyPath(), []byte(kpresp.Privatekey), 0600)
	if err != nil {
		return err
	}
	d.KeyPair = keypairName

	log.Infof("Spawn exoscale host...")
	log.Debugf("Using the following cloud-init file:")
	log.Debugf("%s", userdata)

	machineProfile := egoscale.MachineProfile{
		Template:        tpl,
		ServiceOffering: profile,
		SecurityGroups:  sgs,
		Userdata:        userdata,
		Zone:            zone,
		Keypair:         d.KeyPair,
		Name:            d.MachineName,
	}

	cvmresp, err := client.CreateVirtualMachine(machineProfile)
	if err != nil {
		return err
	}

	vm, err := d.waitForVM(client, cvmresp)
	if err != nil {
		return err
	}
	d.IPAddress = vm.Nic[0].Ipaddress
	d.ID = vm.Id

	return nil
}

func (d *Driver) Start() error {
	client := egoscale.NewClient(d.URL, d.APIKey, d.APISecretKey)

	svmresp, err := client.StartVirtualMachine(d.ID)
	if err != nil {
		return err
	}

	return d.waitForJob(client, svmresp)
}

func (d *Driver) Stop() error {
	client := egoscale.NewClient(d.URL, d.APIKey, d.APISecretKey)

	svmresp, err := client.StopVirtualMachine(d.ID)
	if err != nil {
		return err
	}

	return d.waitForJob(client, svmresp)
}

func (d *Driver) Restart() error {
	client := egoscale.NewClient(d.URL, d.APIKey, d.APISecretKey)

	svmresp, err := client.RebootVirtualMachine(d.ID)
	if err != nil {
		return err
	}

	return d.waitForJob(client, svmresp)
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Remove() error {
	client := egoscale.NewClient(d.URL, d.APIKey, d.APISecretKey)

	// Destroy the SSH key
	if _, err := client.DeleteKeypair(d.KeyPair); err != nil {
		return err
	}

	// Destroy the virtual machine
	dvmresp, err := client.DestroyVirtualMachine(d.ID)
	if err != nil {
		return err
	}
	if err = d.waitForJob(client, dvmresp); err != nil {
		return err
	}
	return nil
}

func (d *Driver) jobIsDone(client *egoscale.Client, jobid string) (bool, error) {
	resp, err := client.PollAsyncJob(jobid)
	if err != nil {
		return true, err
	}
	switch resp.Jobstatus {
	case 0: // Job is still in progress
	case 1: // Job has successfully completed
		return true, nil
	case 2: // Job has failed to complete
		return true, fmt.Errorf("Operation failed to complete")
	default: // Some other code
	}
	return false, nil
}

func (d *Driver) waitForJob(client *egoscale.Client, jobid string) error {
	log.Infof("Waiting for job to complete...")
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		return d.jobIsDone(client, jobid)
	}, 60, 2*time.Second)
}

func (d *Driver) waitForVM(client *egoscale.Client, jobid string) (*egoscale.DeployVirtualMachineResponse, error) {
	if err := d.waitForJob(client, jobid); err != nil {
		return nil, err
	}
	resp, err := client.PollAsyncJob(jobid)
	if err != nil {
		return nil, err
	}
	vm, err := client.AsyncToVirtualMachine(*resp)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// Build a cloud-init user data string that will install and run
// docker.
func (d *Driver) getCloudInit() (string, error) {
	if d.UserDataFile != "" {
		buf, err := ioutil.ReadFile(d.UserDataFile)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}

	return `#cloud-config
manage_etc_hosts: true
`, nil
}
