package azureutil

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/docker/machine/drivers/azure/logutil"
	"github.com/docker/machine/libmachine/log"
)

type cleanupResource interface {
	// Get retrieves if the resource and saves its reference to the instance
	// for further using, returned error is used to determine if the resource
	// exists
	Get(a AzureClient) error

	// Delete deletes the resource
	Delete(a AzureClient) error

	// HasAttachedResources checks the resource reference if it has dependent
	// resources attached to it preventing it from being deleted.
	HasAttachedResources() bool

	// ResourceType returns human-readable name of the type of the resource.
	ResourceType() string

	// LogFields returns the logging fields used during cleanup logging.
	LogFields() logutil.Fields
}

// cleanupResourceIfExists checks if the resource exists, if it does and it
// does not have any attached resources, then deletes the resource. If the
// resource does not exist or is not eligible for cleanup, returns nil. If an
// error is encountered, returns the error.
func (a AzureClient) cleanupResourceIfExists(r cleanupResource) error {
	f := r.LogFields()
	log.Info(fmt.Sprintf("Attempting to clean up %s resource...", r.ResourceType()), f)
	err := r.Get(a)
	if exists, err := checkResourceExistsFromError(err); err != nil {
		return err
	} else if !exists {
		log.Debug(fmt.Sprintf("%s resource does not exist. Skipping.", r.ResourceType()), f)
		return nil
	}

	if !r.HasAttachedResources() {
		log.Debug(fmt.Sprintf("%s does not have any attached dependent resource.", r.ResourceType()), f)
		log.Info(fmt.Sprintf("Removing %s resource...", r.ResourceType()), f)
		return r.Delete(a)
	}
	log.Info(fmt.Sprintf("%s is still in use by other resources, skipping removal.", r.ResourceType()), f)
	return nil
}

// subnetCleanup manages cleanup of Subnet resources
type subnetCleanup struct {
	rg, vnet, name string
	ref            network.Subnet
}

func (c *subnetCleanup) Get(a AzureClient) (err error) {
	c.ref, err = a.subnetsClient().Get(c.rg, c.vnet, c.name, "")
	return err
}

func (c *subnetCleanup) Delete(a AzureClient) error {
	_, err := a.subnetsClient().Delete(c.rg, c.vnet, c.name, nil)
	return err
}

func (c *subnetCleanup) ResourceType() string { return "Subnet" }

func (c *subnetCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *subnetCleanup) HasAttachedResources() bool {
	return c.ref.Properties.IPConfigurations != nil && len(*c.ref.Properties.IPConfigurations) > 0
}

// vnetCleanup manages cleanup of Virtual Network resources.
type vnetCleanup struct {
	rg, name string
	ref      network.VirtualNetwork
}

func (c *vnetCleanup) Get(a AzureClient) (err error) {
	c.ref, err = a.virtualNetworksClient().Get(c.rg, c.name, "")
	return err
}

func (c *vnetCleanup) Delete(a AzureClient) error {
	_, err := a.virtualNetworksClient().Delete(c.rg, c.name, nil)
	return err
}

func (c *vnetCleanup) ResourceType() string { return "Virtual Network" }

func (c *vnetCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *vnetCleanup) HasAttachedResources() bool {
	return c.ref.Properties.Subnets != nil && len(*c.ref.Properties.Subnets) > 0
}

// avSetCleanup manages cleanup of Availability Set resources.
type avSetCleanup struct {
	rg, name string
	ref      compute.AvailabilitySet
}

func (c *avSetCleanup) Get(a AzureClient) (err error) {
	c.ref, err = a.availabilitySetsClient().Get(c.rg, c.name)
	return err
}

func (c *avSetCleanup) Delete(a AzureClient) error {
	_, err := a.availabilitySetsClient().Delete(c.rg, c.name)
	return err
}

func (c *avSetCleanup) ResourceType() string { return "Availability Set" }

func (c *avSetCleanup) LogFields() logutil.Fields { return logutil.Fields{"name": c.name} }

func (c *avSetCleanup) HasAttachedResources() bool {
	return c.ref.Properties.VirtualMachines != nil && len(*c.ref.Properties.VirtualMachines) > 0
}
