package azure

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/docker/machine/drivers/azure/azureutil"
	"github.com/docker/machine/drivers/azure/logutil"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

var (
	environments = map[string]azure.Environment{
		azure.PublicCloud.Name:       azure.PublicCloud,
		azure.USGovernmentCloud.Name: azure.USGovernmentCloud,
		azure.ChinaCloud.Name:        azure.ChinaCloud,
	}
)

// requiredOptionError forms an error from the error indicating the option has
// to be provided with a value for this driver.
type requiredOptionError string

func (r requiredOptionError) Error() string {
	return fmt.Sprintf("%s driver requires the %q option.", driverName, string(r))
}

// newAzureClient creates an AzureClient helper from the Driver context and
// initiates authentication if required.
func (d *Driver) newAzureClient() (*azureutil.AzureClient, error) {
	env, ok := environments[d.Environment]
	if !ok {
		return nil, fmt.Errorf("Invalid Azure environment: %q", d.Environment)
	}

	servicePrincipalToken, err := azureutil.Authenticate(env, d.SubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("Error creating Azure client: %v", err)
	}
	return azureutil.New(env, d.SubscriptionID, servicePrincipalToken), nil
}

// generateSSHKey creates a ssh key pair locally and saves the public key file
// contents in OpenSSH format to the DeploymentContext.
func (d *Driver) generateSSHKey(ctx *azureutil.DeploymentContext) error {
	privPath := d.GetSSHKeyPath()
	pubPath := privPath + ".pub"

	log.Debug("Creating SSH key...", logutil.Fields{
		"pub":  pubPath,
		"priv": privPath,
	})

	if err := ssh.GenerateSSHKey(privPath); err != nil {
		return err
	}
	log.Debug("SSH key pair generated.")

	publicKey, err := ioutil.ReadFile(pubPath)
	ctx.SSHPublicKey = string(publicKey)
	return err
}

// getSecurityRules creates network security group rules based on driver
// configuration such as SSH port, docker port and swarm port.
func (d *Driver) getSecurityRules(extraPorts []string) (*[]network.SecurityRule, error) {
	mkRule := func(priority int, name, description, srcPort, dstPort string) network.SecurityRule {
		return network.SecurityRule{
			Name: to.StringPtr(name),
			Properties: &network.SecurityRulePropertiesFormat{
				Description:              to.StringPtr(description),
				SourceAddressPrefix:      to.StringPtr("*"),
				DestinationAddressPrefix: to.StringPtr("*"),
				SourcePortRange:          to.StringPtr(srcPort),
				DestinationPortRange:     to.StringPtr(dstPort),
				Access:                   network.Allow,
				Direction:                network.Inbound,
				Protocol:                 network.TCP,
				Priority:                 to.Int32Ptr(int32(priority)),
			},
		}
	}

	log.Debugf("Docker port is configured as %d", d.DockerPort)

	// Base ports to be opened for any machine
	rl := []network.SecurityRule{
		mkRule(100, "SSHAllowAny", "Allow ssh from public Internet", "*", fmt.Sprintf("%d", d.BaseDriver.SSHPort)),
		mkRule(300, "DockerAllowAny", "Allow docker engine access (TLS-protected)", "*", fmt.Sprintf("%d", d.DockerPort)),
	}

	// Open swarm port if configured
	if d.BaseDriver.SwarmMaster {
		swarmHost := d.BaseDriver.SwarmHost
		log.Debugf("Swarm host is configured as %q", swarmHost)
		u, err := url.Parse(swarmHost)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse URL %q: %v", swarmHost, err)
		}
		_, swarmPort, err := net.SplitHostPort(u.Host)
		if err != nil {
			return nil, fmt.Errorf("Could not parse swarm port in %q: %v", u.Host, err)
		}
		rl = append(rl, mkRule(500, "DockerSwarmAllowAny", "Allow swarm manager access (TLS-protected)", "*", swarmPort))
	} else {
		log.Debug("Swarm host is not configured.")
	}

	// extra port numbers requested by user
	basePri := 1000
	for i, port := range extraPorts {
		log.Debugf("User-requested port number to be opened on NSG: %v", port)
		r := mkRule(basePri+i, fmt.Sprintf("Port%sAllowAny", port), "User requested port to be accessible from Internet via docker-machine", "*", port)
		rl = append(rl, r)
	}
	log.Debugf("Total NSG rules: %d", len(rl))

	return &rl, nil
}

func (d *Driver) naming() azureutil.ResourceNaming {
	return azureutil.ResourceNaming(d.BaseDriver.MachineName)
}

// ipAddress returns machine’s private or public IP address according to the
// configuration. If no IP address is found it returns empty string.
func (d *Driver) ipAddress() (ip string, err error) {
	c, err := d.newAzureClient()
	if err != nil {
		return "", err
	}

	var ipType string
	if d.UsePrivateIP || d.NoPublicIP {
		ipType = "Private"
		ip, err = c.GetPrivateIPAddress(d.ResourceGroup, d.naming().NIC())
	} else {
		ipType = "Public"
		ip, err = c.GetPublicIPAddress(d.ResourceGroup, d.naming().IP())
	}

	log.Debugf("Retrieving %s IP address...", ipType)
	if err != nil {
		return "", fmt.Errorf("Error querying %s IP: %v", ipType, err)
	}
	if ip == "" {
		log.Debugf("%s IP address is not yet allocated.", ipType)
	}
	return ip, nil
}

func machineStateForVMPowerState(ps azureutil.VMPowerState) state.State {
	m := map[azureutil.VMPowerState]state.State{
		azureutil.Running:      state.Running,
		azureutil.Starting:     state.Starting,
		azureutil.Stopping:     state.Stopping,
		azureutil.Stopped:      state.Stopped,
		azureutil.Deallocating: state.Stopping,
		azureutil.Deallocated:  state.Stopped,
		azureutil.Unknown:      state.None,
	}

	if v, ok := m[ps]; ok {
		return v
	}
	log.Warnf("Azure PowerState %q does not map to a docker-machine state.", ps)
	return state.None
}
