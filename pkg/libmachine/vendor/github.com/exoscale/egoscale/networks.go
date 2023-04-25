package egoscale

import (
	"fmt"
	"net/url"
)

// ListRequest builds the ListNetworks request
func (network *Network) ListRequest() (ListCommand, error) {
	//TODO add tags support
	req := &ListNetworks{
		Account:           network.Account,
		ACLType:           network.ACLType,
		CanUseForDeploy:   &network.CanUseForDeploy,
		DomainID:          network.DomainID,
		ID:                network.ID,
		PhysicalNetworkID: network.PhysicalNetworkID,
		ProjectID:         network.ProjectID,
		RestartRequired:   &network.RestartRequired,
		TrafficType:       network.TrafficType,
		Type:              network.Type,
		VpcID:             network.VpcID,
		ZoneID:            network.ZoneID,
	}

	return req, nil
}

// ResourceType returns the type of the resource
func (*Network) ResourceType() string {
	return "Network"
}

// name returns the CloudStack API command name
func (*CreateNetwork) name() string {
	return "createNetwork"
}

func (*CreateNetwork) response() interface{} {
	return new(CreateNetworkResponse)
}

func (req *CreateNetwork) onBeforeSend(params *url.Values) error {
	// Those fields are required but might be empty
	if req.Name == "" {
		params.Set("name", "")
	}
	if req.DisplayText == "" {
		params.Set("displaytext", "")
	}
	return nil
}

// name returns the CloudStack API command name
func (*UpdateNetwork) name() string {
	return "updateNetwork"
}

func (*UpdateNetwork) asyncResponse() interface{} {
	return new(UpdateNetworkResponse)
}

// name returns the CloudStack API command name
func (*RestartNetwork) name() string {
	return "restartNetwork"
}

func (*RestartNetwork) asyncResponse() interface{} {
	return new(RestartNetworkResponse)
}

// name returns the CloudStack API command name
func (*DeleteNetwork) name() string {
	return "deleteNetwork"
}

func (*DeleteNetwork) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*ListNetworks) name() string {
	return "listNetworks"
}

func (*ListNetworks) response() interface{} {
	return new(ListNetworksResponse)
}

// SetPage sets the current page
func (listNetwork *ListNetworks) SetPage(page int) {
	listNetwork.Page = page
}

// SetPageSize sets the page size
func (listNetwork *ListNetworks) SetPageSize(pageSize int) {
	listNetwork.PageSize = pageSize
}

func (*ListNetworks) each(resp interface{}, callback IterateItemFunc) {
	networks, ok := resp.(*ListNetworksResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListNetworksResponse expected, got %t", resp))
		return
	}

	for i := range networks.Network {
		if !callback(&networks.Network[i], nil) {
			break
		}
	}
}
