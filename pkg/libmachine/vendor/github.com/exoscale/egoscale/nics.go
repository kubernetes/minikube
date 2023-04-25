package egoscale

import (
	"fmt"
)

// ListRequest build a ListNics request from the given Nic
func (nic *Nic) ListRequest() (ListCommand, error) {
	if nic.VirtualMachineID == "" {
		return nil, fmt.Errorf("ListNics command requires the VirtualMachineID field to be set")
	}

	req := &ListNics{
		VirtualMachineID: nic.VirtualMachineID,
		NicID:            nic.ID,
		NetworkID:        nic.NetworkID,
	}

	return req, nil
}

// name returns the CloudStack API command name
func (*ListNics) name() string {
	return "listNics"
}

func (*ListNics) response() interface{} {
	return new(ListNicsResponse)
}

// SetPage sets the current page
func (ls *ListNics) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListNics) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListNics) each(resp interface{}, callback IterateItemFunc) {
	nics := resp.(*ListNicsResponse)
	for i := range nics.Nic {
		if !callback(&(nics.Nic[i]), nil) {
			break
		}
	}
}

// name returns the CloudStack API command name: addIpToNic
func (*AddIPToNic) name() string {
	return "addIpToNic"
}
func (*AddIPToNic) asyncResponse() interface{} {
	return new(AddIPToNicResponse)
}

// name returns the CloudStack API command name: removeIpFromNic
func (*RemoveIPFromNic) name() string {
	return "removeIpFromNic"
}

func (*RemoveIPFromNic) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name: activateIp6
func (*ActivateIP6) name() string {
	return "activateIp6"
}

func (*ActivateIP6) asyncResponse() interface{} {
	return new(ActivateIP6Response)
}
