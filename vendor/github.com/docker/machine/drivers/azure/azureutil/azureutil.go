package azureutil

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/docker/machine/drivers/azure/logutil"
	"github.com/docker/machine/libmachine/log"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	blobstorage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	storageAccountPrefix     = "vhds"   // do not contaminate to user's existing storage accounts
	fmtOSDiskContainer       = "vhd-%s" // place vhds of VMs in separate containers for ease of cleanup
	fmtOSDiskBlobName        = "%s-os-disk.vhd"
	fmtOSDiskResourceName    = "%s-os-disk"
	defaultStorageAPIVersion = blobstorage.DefaultAPIVersion
)

var (
	// Private IPv4 address space per RFC 1918.
	defaultVnetAddressPrefixes = []string{
		"192.168.0.0/16",
		"10.0.0.0/6",
		"172.16.0.0/12"}

	// Polling interval for VM power state check.
	powerStatePollingInterval = time.Second * 5
	waitStartTimeout          = time.Minute * 10
	waitPowerOffTimeout       = time.Minute * 5
)

type AzureClient struct {
	env            azure.Environment
	subscriptionID string
	auth           autorest.Authorizer
}

func New(env azure.Environment, subsID string, auth autorest.Authorizer) *AzureClient {
	return &AzureClient{env, subsID, auth}
}

// RegisterResourceProviders registers current subscription to the specified
// resource provider namespaces if they are not already registered. Namespaces
// are case-insensitive.
func (a AzureClient) RegisterResourceProviders(namespaces ...string) error {
	l, err := a.providersClient().List(nil)
	if err != nil {
		return err
	}
	if l.Value == nil {
		return errors.New("Resource Providers list is returned as nil.")
	}

	m := make(map[string]bool)
	for _, p := range *l.Value {
		m[strings.ToLower(to.String(p.Namespace))] = to.String(p.RegistrationState) == "Registered"
	}

	for _, ns := range namespaces {
		registered, ok := m[strings.ToLower(ns)]
		if !ok {
			return fmt.Errorf("Unknown resource provider %q", ns)
		}
		if registered {
			log.Debugf("Already registered for %q", ns)
		} else {
			log.Info("Registering subscription to resource provider.", logutil.Fields{
				"ns":   ns,
				"subs": a.subscriptionID,
			})
			if _, err := a.providersClient().Register(ns); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateResourceGroup creates a Resource Group if not exists
func (a AzureClient) CreateResourceGroup(name, location string) error {
	if ok, err := a.resourceGroupExists(name); err != nil {
		return err
	} else if ok {
		log.Infof("Resource group %q already exists.", name)
		return nil
	}

	log.Info("Creating resource group.", logutil.Fields{
		"name":     name,
		"location": location})
	_, err := a.resourceGroupsClient().CreateOrUpdate(name,
		resources.ResourceGroup{
			Location: to.StringPtr(location),
		})
	return err
}

func (a AzureClient) resourceGroupExists(name string) (bool, error) {
	log.Info("Querying existing resource group.", logutil.Fields{"name": name})
	_, err := a.resourceGroupsClient().Get(name)
	return checkResourceExistsFromError(err)
}

func (a AzureClient) CreateNetworkSecurityGroup(ctx *DeploymentContext, resourceGroup, name, location string, rules *[]network.SecurityRule) error {
	log.Info("Configuring network security group.", logutil.Fields{
		"name":     name,
		"location": location})
	_, err := a.securityGroupsClient().CreateOrUpdate(resourceGroup, name,
		network.SecurityGroup{
			Location: to.StringPtr(location),
			Properties: &network.SecurityGroupPropertiesFormat{
				SecurityRules: rules,
			},
		}, nil)
	if err != nil {
		return err
	}
	nsg, err := a.securityGroupsClient().Get(resourceGroup, name, "")
	ctx.NetworkSecurityGroupID = to.String(nsg.ID)
	return err
}

func (a AzureClient) DeleteNetworkSecurityGroupIfExists(resourceGroup, name string) error {
	return deleteResourceIfExists("Network Security Group", name,
		func() error {
			_, err := a.securityGroupsClient().Get(resourceGroup, name, "")
			return err
		},
		func() (autorest.Response, error) { return a.securityGroupsClient().Delete(resourceGroup, name, nil) })
}

func (a AzureClient) CreatePublicIPAddress(ctx *DeploymentContext, resourceGroup, name, location string, isStatic bool) error {
	log.Info("Creating public IP address.", logutil.Fields{
		"name":   name,
		"static": isStatic})

	var ipType network.IPAllocationMethod
	if isStatic {
		ipType = network.Static
	} else {
		ipType = network.Dynamic
	}

	_, err := a.publicIPAddressClient().CreateOrUpdate(resourceGroup, name,
		network.PublicIPAddress{
			Location: to.StringPtr(location),
			Properties: &network.PublicIPAddressPropertiesFormat{
				PublicIPAllocationMethod: ipType,
			},
		}, nil)
	if err != nil {
		return err
	}
	ip, err := a.publicIPAddressClient().Get(resourceGroup, name, "")
	ctx.PublicIPAddressID = to.String(ip.ID)
	return err
}

func (a AzureClient) DeletePublicIPAddressIfExists(resourceGroup, name string) error {
	return deleteResourceIfExists("Public IP", name,
		func() error {
			_, err := a.publicIPAddressClient().Get(resourceGroup, name, "")
			return err
		},
		func() (autorest.Response, error) { return a.publicIPAddressClient().Delete(resourceGroup, name, nil) })
}

func (a AzureClient) CreateVirtualNetworkIfNotExists(resourceGroup, name, location string) error {
	f := logutil.Fields{
		"name":     name,
		"location": location}

	log.Info("Querying if virtual network already exists.", f)

	if exists, err := a.virtualNetworkExists(resourceGroup, name); err != nil {
		return err
	} else if exists {
		log.Info("Virtual network already exists.", f)
		return nil
	}

	log.Debug("Creating virtual network.", f)
	_, err := a.virtualNetworksClient().CreateOrUpdate(resourceGroup, name,
		network.VirtualNetwork{
			Location: to.StringPtr(location),
			Properties: &network.VirtualNetworkPropertiesFormat{
				AddressSpace: &network.AddressSpace{
					AddressPrefixes: to.StringSlicePtr(defaultVnetAddressPrefixes),
				},
			},
		}, nil)
	return err
}

func (a AzureClient) virtualNetworkExists(resourceGroup, name string) (bool, error) {
	_, err := a.virtualNetworksClient().Get(resourceGroup, name, "")
	return checkResourceExistsFromError(err)
}

// CleanupVirtualNetworkIfExists removes a subnet if there are no subnets
// attached to it. Note that this method is not safe for multiple concurrent
// writers, in case of races, deployment of a machine could fail or resource
// might not be cleaned up.
func (a AzureClient) CleanupVirtualNetworkIfExists(resourceGroup, name string) error {
	return a.cleanupResourceIfExists(&vnetCleanup{rg: resourceGroup, name: name})
}

func (a AzureClient) GetSubnet(resourceGroup, virtualNetwork, name string) (network.Subnet, error) {
	return a.subnetsClient().Get(resourceGroup, virtualNetwork, name, "")
}

func (a AzureClient) CreateSubnet(ctx *DeploymentContext, resourceGroup, virtualNetwork, name, subnetPrefix string) error {
	log.Info("Configuring subnet.", logutil.Fields{
		"name": name,
		"vnet": virtualNetwork,
		"cidr": subnetPrefix})
	_, err := a.subnetsClient().CreateOrUpdate(resourceGroup, virtualNetwork, name,
		network.Subnet{
			Properties: &network.SubnetPropertiesFormat{
				AddressPrefix: to.StringPtr(subnetPrefix),
			},
		}, nil)
	if err != nil {
		return err
	}
	subnet, err := a.subnetsClient().Get(resourceGroup, virtualNetwork, name, "")
	ctx.SubnetID = to.String(subnet.ID)
	return err
}

// CleanupSubnetIfExists removes a subnet if there are no IP configurations
// (through NICs) are attached to it. Note that this method is not safe for
// multiple concurrent writers, in case of races, deployment of a machine could
// fail or resource might not be cleaned up.
func (a AzureClient) CleanupSubnetIfExists(resourceGroup, virtualNetwork, name string) error {
	return a.cleanupResourceIfExists(&subnetCleanup{
		rg: resourceGroup, vnet: virtualNetwork, name: name,
	})
}

func (a AzureClient) CreateNetworkInterface(ctx *DeploymentContext, resourceGroup, name, location, publicIPAddressID, subnetID, nsgID, privateIPAddress string) error {
	// NOTE(ahmetalpbalkan) This method is expected to fail if the user
	// specified Azure location is different than location of the virtual
	// network as Azure does not support cross-region virtual networks. In this
	// situation, user will get an explanatory API error from Azure.
	log.Info("Creating network interface.", logutil.Fields{"name": name})

	var publicIP *network.PublicIPAddress
	if publicIPAddressID != "" {
		publicIP = &network.PublicIPAddress{ID: to.StringPtr(publicIPAddressID)}
	}

	var privateIPAllocMethod = network.Dynamic
	if privateIPAddress != "" {
		privateIPAllocMethod = network.Static
	}
	_, err := a.networkInterfacesClient().CreateOrUpdate(resourceGroup, name, network.Interface{
		Location: to.StringPtr(location),
		Properties: &network.InterfacePropertiesFormat{
			NetworkSecurityGroup: &network.SecurityGroup{
				ID: to.StringPtr(nsgID),
			},
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr("ip"),
					Properties: &network.InterfaceIPConfigurationPropertiesFormat{
						PrivateIPAddress:          to.StringPtr(privateIPAddress),
						PrivateIPAllocationMethod: privateIPAllocMethod,
						PublicIPAddress:           publicIP,
						Subnet: &network.Subnet{
							ID: to.StringPtr(subnetID),
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}
	nic, err := a.networkInterfacesClient().Get(resourceGroup, name, "")
	ctx.NetworkInterfaceID = to.String(nic.ID)
	return err
}

func (a AzureClient) DeleteNetworkInterfaceIfExists(resourceGroup, name string) error {
	return deleteResourceIfExists("Network Interface", name,
		func() error {
			_, err := a.networkInterfacesClient().Get(resourceGroup, name, "")
			return err
		},
		func() (autorest.Response, error) { return a.networkInterfacesClient().Delete(resourceGroup, name, nil) })
}

func (a AzureClient) CreateStorageAccount(ctx *DeploymentContext, resourceGroup, location string, storageType storage.AccountType) error {
	s, err := a.findOrCreateStorageAccount(resourceGroup, location, storageType)
	ctx.StorageAccount = s
	return err
}

func (a AzureClient) findOrCreateStorageAccount(resourceGroup, location string, storageType storage.AccountType) (*storage.AccountProperties, error) {
	prefix := storageAccountPrefix
	if s, err := a.findStorageAccount(resourceGroup, location, prefix, storageType); err != nil {
		return nil, err
	} else if s != nil {
		return s, nil
	}

	log.Debug("No eligible storage account found.", logutil.Fields{
		"location": location,
		"type":     storageType})
	return a.createStorageAccount(resourceGroup, location, storageType)
}

func (a AzureClient) findStorageAccount(resourceGroup, location, prefix string, storageType storage.AccountType) (*storage.AccountProperties, error) {
	f := logutil.Fields{
		"type":     storageType,
		"prefix":   prefix,
		"location": location}
	log.Debug("Querying existing storage accounts.", f)
	l, err := a.storageAccountsClient().ListByResourceGroup(resourceGroup)
	if err != nil {
		return nil, err
	}

	if l.Value != nil {
		for _, v := range *l.Value {
			log.Debug("Iterating...", logutil.Fields{
				"name":     to.String(v.Name),
				"type":     storageType,
				"location": to.String(v.Location),
			})
			if to.String(v.Location) == location && v.Properties.AccountType == storageType && strings.HasPrefix(to.String(v.Name), prefix) {
				log.Debug("Found eligible storage account.", logutil.Fields{"name": to.String(v.Name)})
				return v.Properties, nil
			}
		}
	}
	log.Debug("No account matching the pattern is found.", f)
	return nil, err
}

func (a AzureClient) createStorageAccount(resourceGroup, location string, storageType storage.AccountType) (*storage.AccountProperties, error) {
	name := randomAzureStorageAccountName() // if it's not random enough, then you're unlucky
	f := logutil.Fields{
		"name":     name,
		"location": location}

	log.Info("Creating storage account.", f)
	_, err := a.storageAccountsClient().Create(resourceGroup, name,
		storage.AccountCreateParameters{
			Location: to.StringPtr(location),
			Properties: &storage.AccountPropertiesCreateParameters{
				AccountType: storageType,
			},
		}, nil)
	if err != nil {
		return nil, err
	}

	s, err := a.storageAccountsClient().GetProperties(resourceGroup, name)
	if err != nil {
		return nil, err
	}
	return s.Properties, nil
}

func (a AzureClient) VirtualMachineExists(resourceGroup, name string) (bool, error) {
	_, err := a.virtualMachinesClient().Get(resourceGroup, name, "")
	return checkResourceExistsFromError(err)
}

func (a AzureClient) DeleteVirtualMachineIfExists(resourceGroup, name string) error {
	var vmRef compute.VirtualMachine
	err := deleteResourceIfExists("Virtual Machine", name,
		func() error {
			vm, err := a.virtualMachinesClient().Get(resourceGroup, name, "")
			vmRef = vm
			return err
		},
		func() (autorest.Response, error) { return a.virtualMachinesClient().Delete(resourceGroup, name, nil) })
	if err != nil {
		return err
	}

	// Remove disk
	if vmRef.Properties != nil {
		vhdURL := to.String(vmRef.Properties.StorageProfile.OsDisk.Vhd.URI)
		return a.removeOSDiskBlob(resourceGroup, name, vhdURL)
	}
	return nil
}

func (a AzureClient) removeOSDiskBlob(resourceGroup, vmName, vhdURL string) error {
	// NOTE(ahmetalpbalkan) Currently Azure APIs do not offer a Delete Virtual
	// Machine functionality which deletes the attached disks along with the VM
	// as well. Therefore we find out the storage account from OS disk URL and
	// fetch storage account keys to delete the container containing the disk.
	log.Debug("Attempting to remove OS disk.", logutil.Fields{"vm": vmName})
	log.Debugf("OS Disk vhd URL: %q", vhdURL)

	vhdContainer := osDiskStorageContainerName(vmName)

	storageAccount, blobServiceBaseURL := extractStorageAccountFromVHDURL(vhdURL)
	if storageAccount == "" {
		log.Warn("Could not extract the storage account name from URL. Please clean up the disk yourself.")
		return nil
	}
	log.Debug("Fetching storage account keys.", logutil.Fields{
		"account":     storageAccount,
		"storageBase": blobServiceBaseURL,
	})
	keys, err := a.storageAccountsClient().ListKeys(resourceGroup, storageAccount)
	if err != nil {
		return err
	}

	storageAccountKey := to.String(keys.Key1)
	bs, err := blobstorage.NewClient(storageAccount, storageAccountKey, blobServiceBaseURL, defaultStorageAPIVersion, true)
	if err != nil {
		return fmt.Errorf("Error constructing blob storage client :%v", err)
	}

	f := logutil.Fields{
		"account":   storageAccount,
		"container": vhdContainer}
	log.Debug("Removing container of disk blobs.", f)
	ok, err := bs.GetBlobService().DeleteContainerIfExists(vhdContainer) // HTTP round-trip will not be inspected
	if err != nil {
		log.Debugf("Container remove happened: %v", ok)
	}
	return err
}

func (a AzureClient) CreateVirtualMachine(resourceGroup, name, location, size, availabilitySetID, networkInterfaceID,
	username, sshPublicKey, imageName string, storageAccount *storage.AccountProperties) error {
	log.Info("Creating virtual machine.", logutil.Fields{
		"name":     name,
		"location": location,
		"size":     size,
		"username": username,
		"osImage":  imageName,
	})

	img, err := parseImageName(imageName)
	if err != nil {
		return err
	}

	var (
		osDiskBlobURL = osDiskStorageBlobURL(storageAccount, name)
		sshKeyPath    = fmt.Sprintf("/home/%s/.ssh/authorized_keys", username)
	)
	log.Debugf("OS disk blob will be placed at: %s", osDiskBlobURL)
	log.Debugf("SSH key will be placed at: %s", sshKeyPath)

	_, err = a.virtualMachinesClient().CreateOrUpdate(resourceGroup, name,
		compute.VirtualMachine{
			Location: to.StringPtr(location),
			Properties: &compute.VirtualMachineProperties{
				AvailabilitySet: &compute.SubResource{
					ID: to.StringPtr(availabilitySetID),
				},
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypes(size),
				},
				NetworkProfile: &compute.NetworkProfile{
					NetworkInterfaces: &[]compute.NetworkInterfaceReference{
						{
							ID: to.StringPtr(networkInterfaceID),
						},
					},
				},
				OsProfile: &compute.OSProfile{
					ComputerName:  to.StringPtr(name),
					AdminUsername: to.StringPtr(username),
					LinuxConfiguration: &compute.LinuxConfiguration{
						DisablePasswordAuthentication: to.BoolPtr(true),
						SSH: &compute.SSHConfiguration{
							PublicKeys: &[]compute.SSHPublicKey{
								{
									Path:    to.StringPtr(sshKeyPath),
									KeyData: to.StringPtr(sshPublicKey),
								},
							},
						},
					},
				},
				StorageProfile: &compute.StorageProfile{
					ImageReference: &compute.ImageReference{
						Publisher: to.StringPtr(img.publisher),
						Offer:     to.StringPtr(img.offer),
						Sku:       to.StringPtr(img.sku),
						Version:   to.StringPtr(img.version),
					},
					OsDisk: &compute.OSDisk{
						Name:         to.StringPtr(fmt.Sprintf(fmtOSDiskResourceName, name)),
						Caching:      compute.ReadWrite,
						CreateOption: compute.FromImage,
						Vhd: &compute.VirtualHardDisk{
							URI: to.StringPtr(osDiskBlobURL),
						},
					},
				},
			},
		}, nil)
	return err
}

func (a AzureClient) GetVirtualMachinePowerState(resourceGroup, name string) (VMPowerState, error) {
	log.Debug("Querying instance view for power state.")
	vm, err := a.virtualMachinesClient().Get(resourceGroup, name, "instanceView")
	if err != nil {
		log.Errorf("Error querying instance view: %v", err)
		return Unknown, err
	}
	return powerStateFromInstanceView(vm.Properties.InstanceView), nil
}

func (a AzureClient) GetAvailabilitySet(resourceGroup, name string) (compute.AvailabilitySet, error) {
	return a.availabilitySetsClient().Get(resourceGroup, name)
}

func (a AzureClient) CreateAvailabilitySetIfNotExists(ctx *DeploymentContext, resourceGroup, name, location string) error {
	f := logutil.Fields{"name": name}
	log.Info("Configuring availability set.", f)
	as, err := a.availabilitySetsClient().CreateOrUpdate(resourceGroup, name,
		compute.AvailabilitySet{
			Location: to.StringPtr(location),
		})
	ctx.AvailabilitySetID = to.String(as.ID)
	return err
}

// CleanupAvailabilitySetIfExists removes an availability set if there are no
// virtual machines attached to it. Note that this method is not safe for
// multiple concurrent writers, in case of races, deployment of a machine could
// fail or resource might not be cleaned up.
func (a AzureClient) CleanupAvailabilitySetIfExists(resourceGroup, name string) error {
	return a.cleanupResourceIfExists(&avSetCleanup{rg: resourceGroup, name: name})
}

// GetPublicIPAddress attempts to get public IP address from the Public IP
// resource. If IP address is not allocated yet, returns empty string.
func (a AzureClient) GetPublicIPAddress(resourceGroup, name string) (string, error) {
	f := logutil.Fields{"name": name}
	log.Debug("Querying public IP address.", f)
	ip, err := a.publicIPAddressClient().Get(resourceGroup, name, "")
	if err != nil {
		return "", err
	}
	if ip.Properties == nil {
		log.Debug("publicIP.Properties is nil. Could not determine IP address", f)
		return "", nil
	}
	return to.String(ip.Properties.IPAddress), nil
}

// GetPrivateIPAddress attempts to retrieve private IP address of the specified
// network interface name.  If IP address is not allocated yet, returns empty
// string.
func (a AzureClient) GetPrivateIPAddress(resourceGroup, name string) (string, error) {
	f := logutil.Fields{"name": name}
	log.Debug("Querying network interface.", f)
	nic, err := a.networkInterfacesClient().Get(resourceGroup, name, "")
	if err != nil {
		return "", err
	}
	if nic.Properties == nil || nic.Properties.IPConfigurations == nil ||
		len(*nic.Properties.IPConfigurations) == 0 {
		log.Debug("No IPConfigurations found on NIC", f)
		return "", nil
	}
	return to.String((*nic.Properties.IPConfigurations)[0].Properties.PrivateIPAddress), nil
}

// StartVirtualMachine starts the virtual machine and waits until it reaches
// the goal state (running) or times out.
func (a AzureClient) StartVirtualMachine(resourceGroup, name string) error {
	log.Info("Starting virtual machine.", logutil.Fields{"vm": name})
	if _, err := a.virtualMachinesClient().Start(resourceGroup, name, nil); err != nil {
		return err
	}
	return a.waitVMPowerState(resourceGroup, name, Running, waitStartTimeout)
}

// StopVirtualMachine power offs the virtual machine and waits until it reaches
// the goal state (stopped) or times out.
func (a AzureClient) StopVirtualMachine(resourceGroup, name string) error {
	log.Info("Stopping virtual machine.", logutil.Fields{"vm": name})
	if _, err := a.virtualMachinesClient().PowerOff(resourceGroup, name, nil); err != nil {
		return err
	}
	return a.waitVMPowerState(resourceGroup, name, Stopped, waitPowerOffTimeout)
}

// RestartVirtualMachine restarts the virtual machine and waits until it reaches
// the goal state (stopped) or times out.
func (a AzureClient) RestartVirtualMachine(resourceGroup, name string) error {
	log.Info("Restarting virtual machine.", logutil.Fields{"vm": name})
	if _, err := a.virtualMachinesClient().Restart(resourceGroup, name, nil); err != nil {
		return err
	}
	return a.waitVMPowerState(resourceGroup, name, Running, waitStartTimeout)
}

// deleteResourceIfExists is an utility method to determine if a resource exists
// from the error returned from its Get response. If so, deletes it. name is
// used only for logging purposes.
func deleteResourceIfExists(resourceType, name string, getFunc func() error, deleteFunc func() (autorest.Response, error)) error {
	f := logutil.Fields{"name": name}
	log.Debug(fmt.Sprintf("Querying if %s exists.", resourceType), f)
	if exists, err := checkResourceExistsFromError(getFunc()); err != nil {
		return err
	} else if !exists {
		log.Info(fmt.Sprintf("%s does not exist. Skipping.", resourceType), f)
		return nil
	}
	log.Info(fmt.Sprintf("Removing %s resource.", resourceType), f)
	_, err := deleteFunc()
	return err
}

// waitVMPowerState polls the Virtual Machine instance view until it reaches the
// specified goal power state or times out. If checking for virtual machine
// state fails or waiting times out, an error is returned.
func (a AzureClient) waitVMPowerState(resourceGroup, name string, goalState VMPowerState, timeout time.Duration) error {
	// NOTE(ahmetalpbalkan): Azure APIs for Start and Stop are actually async
	// operations on which our SDK blocks and does polling until the operation
	// is complete.
	//
	// By the time the issued power cycle operation is complete, the VM will be
	// already in the goal PowerState. Hence, this method will return in the
	// first check, however there is no harm in being defensive.
	log.Debug("Waiting until VM reaches goal power state.", logutil.Fields{
		"vm":        name,
		"goalState": goalState,
		"timeout":   timeout,
	})

	chErr := make(chan error)
	go func(ch chan error) {
		for {
			select {
			case <-ch:
				// channel closed
				return
			default:
				state, err := a.GetVirtualMachinePowerState(resourceGroup, name)
				if err != nil {
					ch <- err
					return
				}
				if state != goalState {
					log.Debug(fmt.Sprintf("Waiting %v...", powerStatePollingInterval),
						logutil.Fields{
							"goalState": goalState,
							"state":     state,
						})
					time.Sleep(powerStatePollingInterval)
				} else {
					log.Debug("Reached goal power state.",
						logutil.Fields{"state": state})
					ch <- nil
					return
				}
			}
		}
	}(chErr)

	select {
	case <-time.After(timeout):
		close(chErr)
		return fmt.Errorf("Waiting for goal state %q timed out after %v", goalState, timeout)
	case err := <-chErr:
		return err
	}
}

// checkExistsFromError inspects an error and returns a true if err is nil,
// false if error is an autorest.Error with StatusCode=404 and will return the
// error back if error is another status code or another type of error.
func checkResourceExistsFromError(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	v, ok := err.(autorest.DetailedError)
	if ok && v.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, v
}

// osDiskStorageBlobURL gives the full url of the VHD blob where the OS disk for
// the given VM should be stored.
func osDiskStorageBlobURL(account *storage.AccountProperties, vmName string) string {
	containerURL := osDiskStorageContainerURL(account, vmName) // has trailing slash
	blobName := fmt.Sprintf(fmtOSDiskBlobName, vmName)
	return containerURL + blobName
}

// osDiskStorageContainerName returns the container name the OS disk for the VM
// should be saved.
func osDiskStorageContainerName(vm string) string { return fmt.Sprintf(fmtOSDiskContainer, vm) }

// osDiskStorageContainerURL crafts a URL with a trailing slash pointing
// to the full Azure Blob Container URL for given VM name.
func osDiskStorageContainerURL(account *storage.AccountProperties, vmName string) string {
	return fmt.Sprintf("%s%s/", to.String(account.PrimaryEndpoints.Blob), osDiskStorageContainerName(vmName))
}

// extractStorageAccountFromVHDURL parses a blob URL and extracts the Azure
// Storage account name from the URL, namely first subdomain of the hostname and
// the Azure Storage service base URL (e.g. core.windows.net). If it could not
// be parsed, returns empty string.
func extractStorageAccountFromVHDURL(vhdURL string) (string, string) {
	u, err := url.Parse(vhdURL)
	if err != nil {
		log.Warn(fmt.Sprintf("URL parse error: %v", err), logutil.Fields{"url": vhdURL})
		return "", ""
	}
	parts := strings.SplitN(u.Host, ".", 2)
	if len(parts) != 2 {
		log.Warnf("Could not split account name and storage base URL: %s", vhdURL)
		return "", ""
	}
	return parts[0], strings.TrimPrefix(parts[1], "blob.") // "blob." prefix will added by azure storage sdk
}
