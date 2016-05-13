package rackspace

import (
	"fmt"

	"github.com/docker/machine/drivers/openstack"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/version"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/rackspace"
)

func unsupportedOpErr(operation string) error {
	return fmt.Errorf("Rackspace does not currently support the %s operation", operation)
}

// Client is a Rackspace specialization of the generic OpenStack driver.
type Client struct {
	openstack.GenericClient
	driver *Driver
}

// Authenticate creates a Rackspace-specific Gophercloud client.
func (c *Client) Authenticate(d *openstack.Driver) error {
	if c.Provider != nil {
		return nil
	}

	log.Debug("Authenticating to Rackspace.", map[string]string{
		"Username": d.Username,
	})

	apiKey := c.driver.APIKey
	opts := gophercloud.AuthOptions{
		Username: d.Username,
		APIKey:   apiKey,
	}

	provider, err := rackspace.NewClient(rackspace.RackspaceUSIdentity)
	if err != nil {
		return err
	}

	provider.UserAgent.Prepend(fmt.Sprintf("docker-machine/v%d", version.APIVersion))

	err = rackspace.Authenticate(provider, opts)
	if err != nil {
		return err
	}

	c.Provider = provider

	return nil
}

// StartInstance is unfortunately not supported on Rackspace at this time.
func (c *Client) StartInstance(d *openstack.Driver) error {
	return unsupportedOpErr("start")
}

// StopInstance is unfortunately not support on Rackspace at this time.
func (c *Client) StopInstance(d *openstack.Driver) error {
	return unsupportedOpErr("stop")
}

// GetInstanceIPAddresses can be short-circuited with the server's AccessIPv4Addr on Rackspace.
func (c *Client) GetInstanceIPAddresses(d *openstack.Driver) ([]openstack.IPAddress, error) {
	server, err := c.GetServerDetail(d)
	if err != nil {
		return nil, err
	}
	return []openstack.IPAddress{
		{
			Network:     "public",
			Address:     server.AccessIPv4,
			AddressType: openstack.Fixed,
		},
	}, nil
}
