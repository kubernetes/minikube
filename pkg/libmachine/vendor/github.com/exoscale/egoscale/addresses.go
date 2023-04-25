package egoscale

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
)

// Get fetches the resource
func (ipaddress *IPAddress) Get(ctx context.Context, client *Client) error {
	if ipaddress.ID == "" && ipaddress.IPAddress == nil {
		return fmt.Errorf("An IPAddress may only be searched using ID or IPAddress")
	}

	req := &ListPublicIPAddresses{
		ID:        ipaddress.ID,
		IPAddress: ipaddress.IPAddress,
		Account:   ipaddress.Account,
		DomainID:  ipaddress.DomainID,
		ProjectID: ipaddress.ProjectID,
		ZoneID:    ipaddress.ZoneID,
	}

	if ipaddress.IsElastic {
		req.IsElastic = &(ipaddress.IsElastic)
	}

	resp, err := client.RequestWithContext(ctx, req)
	if err != nil {
		return err
	}

	ips := resp.(*ListPublicIPAddressesResponse)
	count := len(ips.PublicIPAddress)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("PublicIPAddress not found. id: %s, ipaddress: %s", ipaddress.ID, ipaddress.IPAddress),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one PublicIPAddress was found")
	}

	return copier.Copy(ipaddress, ips.PublicIPAddress[0])
}

// Delete removes the resource
func (ipaddress *IPAddress) Delete(ctx context.Context, client *Client) error {
	if ipaddress.ID == "" {
		return fmt.Errorf("An IPAddress may only be deleted using ID")
	}

	return client.BooleanRequestWithContext(ctx, &DisassociateIPAddress{
		ID: ipaddress.ID,
	})
}

// ResourceType returns the type of the resource
func (*IPAddress) ResourceType() string {
	return "PublicIpAddress"
}

// name returns the CloudStack API command name
func (*AssociateIPAddress) name() string {
	return "associateIpAddress"
}

func (*AssociateIPAddress) asyncResponse() interface{} {
	return new(AssociateIPAddressResponse)
}

// name returns the CloudStack API command name
func (*DisassociateIPAddress) name() string {
	return "disassociateIpAddress"
}
func (*DisassociateIPAddress) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*UpdateIPAddress) name() string {
	return "updateIpAddress"
}
func (*UpdateIPAddress) asyncResponse() interface{} {
	return new(UpdateIPAddressResponse)
}

// name returns the CloudStack API command name
func (*ListPublicIPAddresses) name() string {
	return "listPublicIpAddresses"
}

func (*ListPublicIPAddresses) response() interface{} {
	return new(ListPublicIPAddressesResponse)
}

// ListRequest builds the ListAdresses request
func (ipaddress *IPAddress) ListRequest() (ListCommand, error) {
	req := &ListPublicIPAddresses{
		Account:            ipaddress.Account,
		AllocatedNetworkID: ipaddress.AssociatedNetworkID,
		DomainID:           ipaddress.DomainID,
		ForDisplay:         &ipaddress.ForDisplay,
		//ForVirtualNetwork:  ip.ForVirtualNetwork, change ForVirtualNetwork type for type bool
		ID:                ipaddress.ID,
		IPAddress:         ipaddress.IPAddress,
		IsElastic:         &ipaddress.IsElastic,
		IsSourceNat:       &ipaddress.IsSourceNat,
		PhysicalNetworkID: ipaddress.PhysicalNetworkID,
		ProjectID:         ipaddress.ProjectID,
		VlanID:            ipaddress.VlanID,
		VpcID:             ipaddress.VpcID,
		ZoneID:            ipaddress.ZoneID,
	}

	return req, nil
}

// SetPage sets the current page
func (ls *ListPublicIPAddresses) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListPublicIPAddresses) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListPublicIPAddresses) each(resp interface{}, callback IterateItemFunc) {
	ips, ok := resp.(*ListPublicIPAddressesResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListPublicIPAddressesResponse expected, got %t", resp))
		return
	}

	for i := range ips.PublicIPAddress {
		if !callback(&ips.PublicIPAddress[i], nil) {
			break
		}
	}
}
