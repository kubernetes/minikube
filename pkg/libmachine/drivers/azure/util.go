package azure

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"k8s.io/minikube/pkg/libmachine/drivers/azure/azureutil"
	"k8s.io/minikube/pkg/libmachine/drivers/azure/logutil"
	"k8s.io/minikube/pkg/libmachine/drivers/driverutil"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/ssh"
	"k8s.io/minikube/pkg/libmachine/libmachine/state"
)

var (
	environments = map[string]azure.Environment{
		azure.PublicCloud.Name:       azure.PublicCloud,
		azure.USGovernmentCloud.Name: azure.USGovernmentCloud,
		azure.ChinaCloud.Name:        azure.ChinaCloud,
		azure.GermanCloud.Name:       azure.GermanCloud,
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
		valid := make([]string, 0, len(environments))
		for k := range environments {
			valid = append(valid, k)
		}

		return nil, fmt.Errorf("Invalid Azure environment: %q, supported values: %s", d.Environment, strings.Join(valid, ", "))
	}

	var (
		token *azure.ServicePrincipalToken
		err   error
	)
	if d.ClientID != "" && d.ClientSecret != "" { // use Service Principal auth
		log.Debug("Using Azure service principal authentication.")
		token, err = azureutil.AuthenticateServicePrincipal(env, d.SubscriptionID, d.ClientID, d.ClientSecret)
		if err != nil {
			return nil, fmt.Errorf("Failed to authenticate using service principal credentials: %+v", err)
		}
	} else { // use browser-based device auth
		log.Debug("Using Azure device flow authentication.")
		token, err = azureutil.AuthenticateDeviceFlow(env, d.SubscriptionID)
		if err != nil {
			return nil, fmt.Errorf("Error creating Azure client: %v", err)
		}
	}
	return azureutil.New(env, d.SubscriptionID, token), nil
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
	mkRule := func(priority int, name, description, srcPort, dstPort string, proto network.SecurityRuleProtocol) network.SecurityRule {
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
				Protocol:                 proto,
				Priority:                 to.Int32Ptr(int32(priority)),
			},
		}
	}

	log.Debugf("Docker port is configured as %d", d.DockerPort)

	// Base ports to be opened for any machine
	rl := []network.SecurityRule{
		mkRule(100, "SSHAllowAny", "Allow ssh from public Internet", "*", fmt.Sprintf("%d", d.BaseDriver.SSHPort), network.TCP),
		mkRule(300, "DockerAllowAny", "Allow docker engine access (TLS-protected)", "*", fmt.Sprintf("%d", d.DockerPort), network.TCP),
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
		rl = append(rl, mkRule(500, "DockerSwarmAllowAny", "Allow swarm manager access (TLS-protected)", "*", swarmPort, network.TCP))
	} else {
		log.Debug("Swarm host is not configured.")
	}

	// extra port numbers requested by user
	basePri := 1000
	for i, p := range extraPorts {
		port, protocol := driverutil.SplitPortProto(p)
		proto, err := parseSecurityRuleProtocol(protocol)
		if err != nil {
			return nil, fmt.Errorf("cannot parse security rule protocol: %v", err)
		}
		log.Debugf("User-requested port to be opened on NSG: %v/%s", port, proto)
		name := fmt.Sprintf("Port%s-%sAllowAny", port, proto)
		name = strings.Replace(name, "*", "Asterisk", -1)
		r := mkRule(basePri+i, name, "User requested port to be accessible from Internet via docker-machine", "*", port, proto)
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
		ip, err = c.GetPublicIPAddress(d.ResourceGroup,
			d.naming().IP(),
			d.DNSLabel != "")
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

// parseVirtualNetwork parses Virtual Network input format "[resourcegroup:]name"
// into Resource Group (uses provided one if omitted) and Virtual Network Name
func parseVirtualNetwork(name string, defaultRG string) (string, string) {
	l := strings.SplitN(name, ":", 2)
	if len(l) == 2 {
		return l[0], l[1]
	}
	return defaultRG, name
}

// parseSecurityRuleProtocol parses a protocol string into a network.SecurityRuleProtocol
// and returns error if the protocol is not supported
func parseSecurityRuleProtocol(proto string) (network.SecurityRuleProtocol, error) {
	switch strings.ToLower(proto) {
	case "tcp":
		return network.TCP, nil
	case "udp":
		return network.UDP, nil
	case "*":
		return network.Asterisk, nil
	default:
		return "", fmt.Errorf("invalid protocol %s", proto)
	}
}
