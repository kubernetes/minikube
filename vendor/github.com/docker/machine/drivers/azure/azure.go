package azure

import (
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/docker/machine/drivers/azure/azureutil"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"

	"github.com/Azure/azure-sdk-for-go/arm/storage"
)

const (
	defaultAzureEnvironment     = "AzurePublicCloud"
	defaultAzureResourceGroup   = "docker-machine"
	defaultAzureSize            = "Standard_A2"
	defaultAzureLocation        = "westus"
	defaultSSHUser              = "docker-user" // 'root' not allowed on Azure
	defaultDockerPort           = 2376
	defaultAzureImage           = "canonical:UbuntuServer:15.10:latest"
	defaultAzureVNet            = "docker-machine-vnet"
	defaultAzureSubnet          = "docker-machine"
	defaultAzureSubnetPrefix    = "192.168.0.0/16"
	defaultStorageType          = storage.StandardLRS
	defaultAzureAvailabilitySet = "docker-machine"
)

const (
	flAzureEnvironment     = "azure-environment"
	flAzureSubscriptionID  = "azure-subscription-id"
	flAzureResourceGroup   = "azure-resource-group"
	flAzureSSHUser         = "azure-ssh-user"
	flAzureDockerPort      = "azure-docker-port"
	flAzureLocation        = "azure-location"
	flAzureSize            = "azure-size"
	flAzureImage           = "azure-image"
	flAzureVNet            = "azure-vnet"
	flAzureSubnet          = "azure-subnet"
	flAzureSubnetPrefix    = "azure-subnet-prefix"
	flAzureAvailabilitySet = "azure-availability-set"
	flAzurePorts           = "azure-open-port"
	flAzurePrivateIPAddr   = "azure-private-ip-address"
	flAzureUsePrivateIP    = "azure-use-private-ip"
	flAzureStaticPublicIP  = "azure-static-public-ip"
	flAzureNoPublicIP      = "azure-no-public-ip"
)

const (
	driverName = "azure"
	sshPort    = 22
)

// Driver represents Azure Docker Machine Driver.
type Driver struct {
	*drivers.BaseDriver

	Environment    string
	SubscriptionID string
	ResourceGroup  string

	DockerPort      int
	Location        string
	Size            string
	Image           string
	VirtualNetwork  string
	SubnetName      string
	SubnetPrefix    string
	AvailabilitySet string

	OpenPorts      []string
	PrivateIPAddr  string
	UsePrivateIP   bool
	NoPublicIP     bool
	StaticPublicIP bool

	// Ephemeral fields
	ctx        *azureutil.DeploymentContext
	resolvedIP string // cache
}

// NewDriver returns a new driver instance.
func NewDriver(hostName, storePath string) drivers.Driver {
	// NOTE(ahmetalpbalkan): any driver initialization I do here gets lost
	// afterwards, especially for non-Create RPC calls. Therefore I am mostly
	// making rest of the driver stateless by just relying on the following
	// piece of info.
	d := &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
	return d
}

// GetCreateFlags returns list of create flags driver accepts.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   flAzureEnvironment,
			Usage:  "Azure environment (e.g. AzurePublicCloud, AzureChinaCloud)",
			EnvVar: "AZURE_ENVIRONMENT",
			Value:  defaultAzureEnvironment,
		},
		mcnflag.StringFlag{
			Name:   flAzureSubscriptionID,
			Usage:  "Azure Subscription ID",
			EnvVar: "AZURE_SUBSCRIPTION_ID",
		},
		mcnflag.StringFlag{
			Name:   flAzureResourceGroup,
			Usage:  "Azure Resource Group name (will be created if missing)",
			EnvVar: "AZURE_RESOURCE_GROUP",
			Value:  defaultAzureResourceGroup,
		},
		mcnflag.StringFlag{
			Name:   flAzureSSHUser,
			Usage:  "Username for SSH login",
			EnvVar: "AZURE_SSH_USER",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			Name:   flAzureDockerPort,
			Usage:  "Port number for Docker engine",
			EnvVar: "AZURE_DOCKER_PORT",
			Value:  defaultDockerPort,
		},
		mcnflag.StringFlag{
			Name:   flAzureLocation,
			Usage:  "Azure region to create the virtual machine",
			EnvVar: "AZURE_LOCATION",
			Value:  defaultAzureLocation,
		},
		mcnflag.StringFlag{
			Name:   flAzureSize,
			Usage:  "Size for Azure Virtual Machine",
			EnvVar: "AZURE_SIZE",
			Value:  defaultAzureSize,
		},
		mcnflag.StringFlag{
			Name:   flAzureImage,
			Usage:  "Azure virtual machine OS image",
			EnvVar: "AZURE_IMAGE",
			Value:  defaultAzureImage,
		},
		mcnflag.StringFlag{
			Name:   flAzureVNet,
			Usage:  "Azure Virtual Network name to connect the virtual machine",
			EnvVar: "AZURE_VNET",
			Value:  defaultAzureVNet,
		},
		mcnflag.StringFlag{
			Name:   flAzureSubnet,
			Usage:  "Azure Subnet Name to be used within the Virtual Network",
			EnvVar: "AZURE_SUBNET",
			Value:  defaultAzureSubnet,
		},
		mcnflag.StringFlag{
			Name:   flAzureSubnetPrefix,
			Usage:  "Private CIDR block to be used for the new subnet, should comply RFC 1918",
			EnvVar: "AZURE_SUBNET_PREFIX",
			Value:  defaultAzureSubnetPrefix,
		},
		mcnflag.StringFlag{
			Name:   flAzureAvailabilitySet,
			Usage:  "Azure Availability Set to place the virtual machine into",
			EnvVar: "AZURE_AVAILABILITY_SET",
			Value:  defaultAzureAvailabilitySet,
		},
		mcnflag.StringFlag{
			Name:  flAzurePrivateIPAddr,
			Usage: "Specify a static private IP address for the machine",
		},
		mcnflag.BoolFlag{
			Name:  flAzureUsePrivateIP,
			Usage: "Use private IP address of the machine to connect",
		},
		mcnflag.BoolFlag{
			Name:  flAzureNoPublicIP,
			Usage: "Do not create a public IP address for the machine",
		},
		mcnflag.BoolFlag{
			Name:  flAzureStaticPublicIP,
			Usage: "Assign a static public IP address to the machine",
		},
		mcnflag.StringSliceFlag{
			Name:  flAzurePorts,
			Usage: "Make the specified port number accessible from the Internet",
		},
	}
}

// SetConfigFromFlags initializes driver values from the command line values
// and checks if the arguments have values.
func (d *Driver) SetConfigFromFlags(fl drivers.DriverOptions) error {
	// Initialize driver context for machine
	d.ctx = &azureutil.DeploymentContext{}

	// Required string flags
	flags := []struct {
		target *string
		flag   string
	}{
		{&d.BaseDriver.SSHUser, flAzureSSHUser},
		{&d.SubscriptionID, flAzureSubscriptionID},
		{&d.ResourceGroup, flAzureResourceGroup},
		{&d.Location, flAzureLocation},
		{&d.Size, flAzureSize},
		{&d.Image, flAzureImage},
		{&d.VirtualNetwork, flAzureVNet},
		{&d.SubnetName, flAzureSubnet},
		{&d.SubnetPrefix, flAzureSubnetPrefix},
		{&d.AvailabilitySet, flAzureAvailabilitySet},
	}
	for _, f := range flags {
		*f.target = fl.String(f.flag)
		if *f.target == "" {
			return requiredOptionError(f.flag)
		}
	}

	// Optional flags or Flags of other types
	d.Environment = fl.String(flAzureEnvironment)
	d.OpenPorts = fl.StringSlice(flAzurePorts)
	d.PrivateIPAddr = fl.String(flAzurePrivateIPAddr)
	d.UsePrivateIP = fl.Bool(flAzureUsePrivateIP)
	d.NoPublicIP = fl.Bool(flAzureNoPublicIP)
	d.StaticPublicIP = fl.Bool(flAzureStaticPublicIP)
	d.DockerPort = fl.Int(flAzureDockerPort)

	// Set flags on the BaseDriver
	d.BaseDriver.SSHPort = sshPort
	d.SetSwarmConfigFromFlags(fl)

	log.Debug("Set configuration from flags.")
	return nil
}

// DriverName returns the name of the driver.
func (d *Driver) DriverName() string { return driverName }

// PreCreateCheck validates if driver values are valid to create the machine.
func (d *Driver) PreCreateCheck() (err error) {
	c, err := d.newAzureClient()
	if err != nil {
		return err
	}

	// Register used resource providers with current Azure subscription.
	if err := c.RegisterResourceProviders(
		"Microsoft.Compute",
		"Microsoft.Network",
		"Microsoft.Storage"); err != nil {
		return err
	}

	// Validate if firewall rules can be read correctly
	d.ctx.FirewallRules, err = d.getSecurityRules(d.OpenPorts)
	if err != nil {
		return err
	}

	// Check if virtual machine exists. An existing virtual machine cannot be updated.
	log.Debug("Checking if Virtual Machine already exists.")
	if exists, err := c.VirtualMachineExists(d.ResourceGroup, d.naming().VM()); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("Virtual Machine with name %s already exists in resource group %q", d.naming().VM(), d.ResourceGroup)
	}

	// NOTE(ahmetalpbalkan) we could have done more checks here but Azure often
	// returns meaningful error messages and it would be repeating the backend
	// logic on the client side. Some examples:
	//   - Deployment of a machine to an existing Virtual Network fails if
	//     virtual network is in a different region.
	//   - Changing IP Address space of a subnet would fail if there are machines
	//     running in the Virtual Network.
	log.Info("Completed machine pre-create checks.")
	return nil
}

// Create creates the virtual machine.
func (d *Driver) Create() error {
	// NOTE(ahmetalpbalkan): We can probably parallelize the sh*t out of this.
	// However that would lead to a concurrency logic and while creation of a
	// resource fails, other ones would be kicked off, which could lead to a
	// resource leak. This is slower but safer.
	c, err := d.newAzureClient()
	if err != nil {
		return err
	}

	if err := c.CreateResourceGroup(d.ResourceGroup, d.Location); err != nil {
		return err
	}
	if err := c.CreateAvailabilitySetIfNotExists(d.ctx, d.ResourceGroup, d.AvailabilitySet, d.Location); err != nil {
		return err
	}
	if err := c.CreateNetworkSecurityGroup(d.ctx, d.ResourceGroup, d.naming().NSG(), d.Location, d.ctx.FirewallRules); err != nil {
		return err
	}
	if err := c.CreateVirtualNetworkIfNotExists(d.ResourceGroup, d.VirtualNetwork, d.Location); err != nil {
		return err
	}
	if err := c.CreateSubnet(d.ctx, d.ResourceGroup, d.VirtualNetwork, d.SubnetName, d.SubnetPrefix); err != nil {
		return err
	}
	if d.NoPublicIP {
		log.Info("Not creating a public IP address.")
	} else {
		if err := c.CreatePublicIPAddress(d.ctx, d.ResourceGroup, d.naming().IP(), d.Location, d.StaticPublicIP); err != nil {
			return err
		}
	}
	if err := c.CreateNetworkInterface(d.ctx, d.ResourceGroup, d.naming().NIC(), d.Location,
		d.ctx.PublicIPAddressID, d.ctx.SubnetID, d.ctx.NetworkSecurityGroupID, d.PrivateIPAddr); err != nil {
		return err
	}
	if err := c.CreateStorageAccount(d.ctx, d.ResourceGroup, d.Location, defaultStorageType); err != nil {
		return err
	}
	if err := d.generateSSHKey(d.ctx); err != nil {
		return err
	}
	if err := c.CreateVirtualMachine(d.ResourceGroup, d.naming().VM(), d.Location, d.Size, d.ctx.AvailabilitySetID,
		d.ctx.NetworkInterfaceID, d.BaseDriver.SSHUser, d.ctx.SSHPublicKey, d.Image, d.ctx.StorageAccount); err != nil {
		return err
	}
	return nil
}

// Remove deletes the virtual machine and resources associated to it.
func (d *Driver) Remove() error {
	if err := d.checkLegacyDriver(false); err != nil {
		return err
	}

	// NOTE(ahmetalpbalkan):
	//   - remove attemps are best effort and if a resource is already gone, we
	//     continue removing other resources instead of failing.
	//   - we can probably do a lot of parallelization here but a sequential
	//     logic works fine too. If we were to detach the NIC from the VM and
	//     then delete the VM, this could enable some parallelization.

	log.Info("NOTICE: Please check Azure portal/CLI to make sure you have no leftover resources to avoid unexpected charges.")
	c, err := d.newAzureClient()
	if err != nil {
		return err
	}
	if err := c.DeleteVirtualMachineIfExists(d.ResourceGroup, d.naming().VM()); err != nil {
		return err
	}
	if err := c.DeleteNetworkInterfaceIfExists(d.ResourceGroup, d.naming().NIC()); err != nil {
		return err
	}
	if err := c.DeletePublicIPAddressIfExists(d.ResourceGroup, d.naming().IP()); err != nil {
		return err
	}
	if err := c.DeleteNetworkSecurityGroupIfExists(d.ResourceGroup, d.naming().NSG()); err != nil {
		return err
	}
	if err := c.CleanupAvailabilitySetIfExists(d.ResourceGroup, d.AvailabilitySet); err != nil {
		return err
	}
	if err := c.CleanupSubnetIfExists(d.ResourceGroup, d.VirtualNetwork, d.SubnetName); err != nil {
		return err
	}
	if err := c.CleanupVirtualNetworkIfExists(d.ResourceGroup, d.VirtualNetwork); err != nil {
		return err
	}
	return nil
}

// GetIP returns public IP address or hostname of the machine instance.
func (d *Driver) GetIP() (string, error) {
	if err := d.checkLegacyDriver(true); err != nil {
		return "", err
	}

	if d.resolvedIP == "" {
		ip, err := d.ipAddress()
		if err != nil {
			return "", err
		}
		d.resolvedIP = ip
	}
	log.Debugf("Machine IP address resolved to: %s", d.resolvedIP)
	return d.resolvedIP, nil
}

// GetSSHHostname returns an IP address or hostname for the machine instance.
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetURL returns a socket address to connect to Docker engine of the machine
// instance.
func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	// NOTE (ahmetalpbalkan) I noticed that this is not used until machine is
	// actually created and provisioned. By then GetIP() should be returning
	// a non-empty IP address as the VM is already allocated and connected to.
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	u := (&url.URL{
		Scheme: "tcp",
		Host:   net.JoinHostPort(ip, fmt.Sprintf("%d", d.DockerPort)),
	}).String()
	log.Debugf("Machine URL is resolved to: %s", u)
	return u, nil
}

// GetState returns the state of the virtual machine role instance.
func (d *Driver) GetState() (state.State, error) {
	if err := d.checkLegacyDriver(true); err != nil {
		return state.None, err
	}

	c, err := d.newAzureClient()
	if err != nil {
		return state.None, err
	}
	powerState, err := c.GetVirtualMachinePowerState(
		d.ResourceGroup, d.naming().VM())
	if err != nil {
		return state.None, err
	}

	machineState := machineStateForVMPowerState(powerState)
	log.Debugf("Determined Azure PowerState=%q, docker-machine state=%q",
		powerState, machineState)
	return machineState, nil
}

// Start issues a power on for the virtual machine instance.
func (d *Driver) Start() error {
	if err := d.checkLegacyDriver(true); err != nil {
		return err
	}

	c, err := d.newAzureClient()
	if err != nil {
		return err
	}
	return c.StartVirtualMachine(d.ResourceGroup, d.naming().VM())
}

// Stop issues a power off for the virtual machine instance.
func (d *Driver) Stop() error {
	if err := d.checkLegacyDriver(true); err != nil {
		return err
	}

	c, err := d.newAzureClient()
	if err != nil {
		return err
	}
	log.Info("NOTICE: Stopping an Azure Virtual Machine is just going to power it off, not deallocate.")
	log.Info("NOTICE: You should remove the machine if you would like to avoid unexpected costs.")
	return c.StopVirtualMachine(d.ResourceGroup, d.naming().VM())
}

// Restart reboots the virtual machine instance.
func (d *Driver) Restart() error {
	if err := d.checkLegacyDriver(true); err != nil {
		return err
	}

	// NOTE(ahmetalpbalkan) Azure will always keep the VM in Running state
	// during the restart operation. Hence we rely on returned async operation
	// polling to make sure the reboot is waited upon.
	c, err := d.newAzureClient()
	if err != nil {
		return err
	}
	return c.RestartVirtualMachine(d.ResourceGroup, d.naming().VM())
}

// Kill stops the virtual machine role instance.
func (d *Driver) Kill() error {
	// NOTE(ahmetalpbalkan) In Azure, there is no kill option for virtual
	// machines, Stop() is the closest option.
	log.Debug("Azure does not implement kill. Calling Stop instead.")
	return d.Stop()
}

// checkLegacyDriver errors out if it encounters an Azure VM created with the
// legacy (<=0.6.0) docker-machine Azure driver.
func (d *Driver) checkLegacyDriver(short bool) error {
	if d.ResourceGroup == "" {
		if short {
			return errors.New("New azure driver cannot manage old VMs, downgrade to v0.6.0")
		}
		return errors.New("New azure driver uses the new Azure Resource Manager APIs and therefore cannot manage this existing machine created with old azure driver. Please downgrade to docker-machine 0.6.0 to continue using these machines or to remove them.")
	}
	return nil
}
