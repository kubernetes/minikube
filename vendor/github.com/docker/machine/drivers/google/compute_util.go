package google

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/log"
	raw "google.golang.org/api/compute/v1"

	"errors"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
)

// ComputeUtil is used to wrap the raw GCE API code and store common parameters.
type ComputeUtil struct {
	zone              string
	instanceName      string
	userName          string
	project           string
	diskTypeURL       string
	address           string
	preemptible       bool
	useInternalIP     bool
	useInternalIPOnly bool
	service           *raw.Service
	zoneURL           string
	globalURL         string
	SwarmMaster       bool
	SwarmHost         string
}

const (
	apiURL             = "https://www.googleapis.com/compute/v1/projects/"
	firewallRule       = "docker-machines"
	port               = "2376"
	firewallTargetTag  = "docker-machine"
	dockerStartCommand = "sudo service docker start"
	dockerStopCommand  = "sudo service docker stop"
)

// NewComputeUtil creates and initializes a ComputeUtil.
func newComputeUtil(driver *Driver) (*ComputeUtil, error) {
	client, err := google.DefaultClient(oauth2.NoContext, raw.ComputeScope)
	if err != nil {
		return nil, err
	}

	service, err := raw.New(client)
	if err != nil {
		return nil, err
	}

	return &ComputeUtil{
		zone:              driver.Zone,
		instanceName:      driver.MachineName,
		userName:          driver.SSHUser,
		project:           driver.Project,
		diskTypeURL:       driver.DiskType,
		address:           driver.Address,
		preemptible:       driver.Preemptible,
		useInternalIP:     driver.UseInternalIP,
		useInternalIPOnly: driver.UseInternalIPOnly,
		service:           service,
		zoneURL:           apiURL + driver.Project + "/zones/" + driver.Zone,
		globalURL:         apiURL + driver.Project + "/global",
		SwarmMaster:       driver.SwarmMaster,
		SwarmHost:         driver.SwarmHost,
	}, nil
}

func (c *ComputeUtil) diskName() string {
	return c.instanceName + "-disk"
}

func (c *ComputeUtil) diskType() string {
	return apiURL + c.project + "/zones/" + c.zone + "/diskTypes/" + c.diskTypeURL
}

// disk returns the persistent disk attached to the vm.
func (c *ComputeUtil) disk() (*raw.Disk, error) {
	return c.service.Disks.Get(c.project, c.zone, c.diskName()).Do()
}

// deleteDisk deletes the persistent disk.
func (c *ComputeUtil) deleteDisk() error {
	disk, _ := c.disk()
	if disk == nil {
		return nil
	}

	log.Infof("Deleting disk.")
	op, err := c.service.Disks.Delete(c.project, c.zone, c.diskName()).Do()
	if err != nil {
		return err
	}

	log.Infof("Waiting for disk to delete.")
	return c.waitForRegionalOp(op.Name)
}

// staticAddress returns the external static IP address.
func (c *ComputeUtil) staticAddress() (string, error) {
	// is the address a name?
	isName, err := regexp.MatchString("[a-z]([-a-z0-9]*[a-z0-9])?", c.address)
	if err != nil {
		return "", err
	}

	if !isName {
		return c.address, nil
	}

	// resolve the address by name
	externalAddress, err := c.service.Addresses.Get(c.project, c.region(), c.address).Do()
	if err != nil {
		return "", err
	}

	return externalAddress.Address, nil
}

func (c *ComputeUtil) region() string {
	return c.zone[:len(c.zone)-2]
}

func (c *ComputeUtil) firewallRule() (*raw.Firewall, error) {
	return c.service.Firewalls.Get(c.project, firewallRule).Do()
}

func missingOpenedPorts(rule *raw.Firewall, ports []string) []string {
	missing := []string{}
	opened := map[string]bool{}

	for _, allowed := range rule.Allowed {
		for _, allowedPort := range allowed.Ports {
			opened[allowedPort] = true
		}
	}

	for _, port := range ports {
		if !opened[port] {
			missing = append(missing, port)
		}
	}

	return missing
}

func (c *ComputeUtil) portsUsed() ([]string, error) {
	ports := []string{port}

	if c.SwarmMaster {
		u, err := url.Parse(c.SwarmHost)
		if err != nil {
			return nil, fmt.Errorf("error authorizing port for swarm: %s", err)
		}

		swarmPort := strings.Split(u.Host, ":")[1]
		ports = append(ports, swarmPort)
	}

	return ports, nil
}

// openFirewallPorts configures the firewall to open docker and swarm ports.
func (c *ComputeUtil) openFirewallPorts() error {
	log.Infof("Opening firewall ports")

	create := false
	rule, _ := c.firewallRule()
	if rule == nil {
		create = true
		rule = &raw.Firewall{
			Name:         firewallRule,
			Allowed:      []*raw.FirewallAllowed{},
			SourceRanges: []string{"0.0.0.0/0"},
			TargetTags:   []string{firewallTargetTag},
		}
	}

	portsUsed, err := c.portsUsed()
	if err != nil {
		return err
	}

	missingPorts := missingOpenedPorts(rule, portsUsed)
	if len(missingPorts) == 0 {
		return nil
	}

	rule.Allowed = append(rule.Allowed, &raw.FirewallAllowed{
		IPProtocol: "tcp",
		Ports:      missingPorts,
	})

	var op *raw.Operation
	if create {
		op, err = c.service.Firewalls.Insert(c.project, rule).Do()
	} else {
		op, err = c.service.Firewalls.Update(c.project, firewallRule, rule).Do()
	}

	if err != nil {
		return err
	}

	return c.waitForGlobalOp(op.Name)
}

// instance retrieves the instance.
func (c *ComputeUtil) instance() (*raw.Instance, error) {
	return c.service.Instances.Get(c.project, c.zone, c.instanceName).Do()
}

// createInstance creates a GCE VM instance.
func (c *ComputeUtil) createInstance(d *Driver) error {
	log.Infof("Creating instance")

	instance := &raw.Instance{
		Name:        c.instanceName,
		Description: "docker host vm",
		MachineType: c.zoneURL + "/machineTypes/" + d.MachineType,
		Disks: []*raw.AttachedDisk{
			{
				Boot:       true,
				AutoDelete: false,
				Type:       "PERSISTENT",
				Mode:       "READ_WRITE",
			},
		},
		NetworkInterfaces: []*raw.NetworkInterface{
			{
				Network: c.globalURL + "/networks/default",
			},
		},
		Tags: &raw.Tags{
			Items: parseTags(d),
		},
		ServiceAccounts: []*raw.ServiceAccount{
			{
				Email:  "default",
				Scopes: strings.Split(d.Scopes, ","),
			},
		},
		Scheduling: &raw.Scheduling{
			Preemptible: c.preemptible,
		},
	}

	if !c.useInternalIPOnly {
		cfg := &raw.AccessConfig{
			Type: "ONE_TO_ONE_NAT",
		}
		instance.NetworkInterfaces[0].AccessConfigs = append(instance.NetworkInterfaces[0].AccessConfigs, cfg)
	}

	if c.address != "" {
		staticAddress, err := c.staticAddress()
		if err != nil {
			return err
		}

		instance.NetworkInterfaces[0].AccessConfigs[0].NatIP = staticAddress
	}

	disk, err := c.disk()
	if disk == nil || err != nil {
		instance.Disks[0].InitializeParams = &raw.AttachedDiskInitializeParams{
			DiskName:    c.diskName(),
			SourceImage: d.MachineImage,
			// The maximum supported disk size is 1000GB, the cast should be fine.
			DiskSizeGb: int64(d.DiskSize),
			DiskType:   c.diskType(),
		}
	} else {
		instance.Disks[0].Source = c.zoneURL + "/disks/" + c.instanceName + "-disk"
	}
	op, err := c.service.Instances.Insert(c.project, c.zone, instance).Do()

	if err != nil {
		return err
	}

	log.Infof("Waiting for Instance")
	if err = c.waitForRegionalOp(op.Name); err != nil {
		return err
	}

	instance, err = c.instance()
	if err != nil {
		return err
	}

	return c.uploadSSHKey(instance, d.GetSSHKeyPath())
}

// configureInstance configures an existing instance for use with Docker Machine.
func (c *ComputeUtil) configureInstance(d *Driver) error {
	log.Infof("Configuring instance")

	instance, err := c.instance()
	if err != nil {
		return err
	}

	if err := c.addFirewallTag(instance); err != nil {
		return err
	}

	return c.uploadSSHKey(instance, d.GetSSHKeyPath())
}

// addFirewallTag adds a tag to the instance to match the firewall rule.
func (c *ComputeUtil) addFirewallTag(instance *raw.Instance) error {
	log.Infof("Adding tag for the firewall rule")

	tags := instance.Tags
	for _, tag := range tags.Items {
		if tag == firewallTargetTag {
			return nil
		}
	}

	tags.Items = append(tags.Items, firewallTargetTag)

	op, err := c.service.Instances.SetTags(c.project, c.zone, instance.Name, tags).Do()
	if err != nil {
		return err
	}

	return c.waitForRegionalOp(op.Name)
}

// uploadSSHKey updates the instance metadata with the given ssh key.
func (c *ComputeUtil) uploadSSHKey(instance *raw.Instance, sshKeyPath string) error {
	log.Infof("Uploading SSH Key")

	sshKey, err := ioutil.ReadFile(sshKeyPath + ".pub")
	if err != nil {
		return err
	}

	metaDataValue := fmt.Sprintf("%s:%s %s\n", c.userName, strings.TrimSpace(string(sshKey)), c.userName)

	op, err := c.service.Instances.SetMetadata(c.project, c.zone, c.instanceName, &raw.Metadata{
		Fingerprint: instance.Metadata.Fingerprint,
		Items: []*raw.MetadataItems{
			{
				Key:   "sshKeys",
				Value: &metaDataValue,
			},
		},
	}).Do()

	return c.waitForRegionalOp(op.Name)
}

// parseTags computes the tags for the instance.
func parseTags(d *Driver) []string {
	tags := []string{firewallTargetTag}

	if d.Tags != "" {
		tags = append(tags, strings.Split(d.Tags, ",")...)
	}

	return tags
}

// deleteInstance deletes the instance, leaving the persistent disk.
func (c *ComputeUtil) deleteInstance() error {
	log.Infof("Deleting instance.")
	op, err := c.service.Instances.Delete(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return err
	}

	log.Infof("Waiting for instance to delete.")
	return c.waitForRegionalOp(op.Name)
}

// stopInstance stops the instance.
func (c *ComputeUtil) stopInstance() error {
	op, err := c.service.Instances.Stop(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return err
	}

	log.Infof("Waiting for instance to stop.")
	return c.waitForRegionalOp(op.Name)
}

// startInstance starts the instance.
func (c *ComputeUtil) startInstance() error {
	op, err := c.service.Instances.Start(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return err
	}

	log.Infof("Waiting for instance to start.")
	return c.waitForRegionalOp(op.Name)
}

// waitForOp waits for the operation to finish.
func (c *ComputeUtil) waitForOp(opGetter func() (*raw.Operation, error)) error {
	for {
		op, err := opGetter()
		if err != nil {
			return err
		}

		log.Debugf("Operation %q status: %s", op.Name, op.Status)
		if op.Status == "DONE" {
			if op.Error != nil {
				return fmt.Errorf("Operation error: %v", *op.Error.Errors[0])
			}
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

// waitForRegionalOp waits for the regional operation to finish.
func (c *ComputeUtil) waitForRegionalOp(name string) error {
	return c.waitForOp(func() (*raw.Operation, error) {
		return c.service.ZoneOperations.Get(c.project, c.zone, name).Do()
	})
}

// waitForGlobalOp waits for the global operation to finish.
func (c *ComputeUtil) waitForGlobalOp(name string) error {
	return c.waitForOp(func() (*raw.Operation, error) {
		return c.service.GlobalOperations.Get(c.project, name).Do()
	})
}

// ip retrieves and returns the external IP address of the instance.
func (c *ComputeUtil) ip() (string, error) {
	instance, err := c.service.Instances.Get(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return "", unwrapGoogleError(err)
	}

	nic := instance.NetworkInterfaces[0]
	if c.useInternalIP {
		return nic.NetworkIP, nil
	}
	return nic.AccessConfigs[0].NatIP, nil
}

func unwrapGoogleError(err error) error {
	if googleErr, ok := err.(*googleapi.Error); ok {
		return errors.New(googleErr.Message)
	}

	return err
}
