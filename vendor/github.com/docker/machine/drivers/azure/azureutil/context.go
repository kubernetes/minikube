package azureutil

import (
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
)

// DeploymentContext contains references to various sources created and then
// used in creating other resources.
type DeploymentContext struct {
	VirtualNetworkExists   bool
	StorageAccount         *storage.AccountProperties
	PublicIPAddressID      string
	NetworkSecurityGroupID string
	SubnetID               string
	NetworkInterfaceID     string
	SSHPublicKey           string
	AvailabilitySetID      string
	FirewallRules          *[]network.SecurityRule
}
