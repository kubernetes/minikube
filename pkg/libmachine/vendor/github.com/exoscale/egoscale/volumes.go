package egoscale

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
)

// ResourceType returns the type of the resource
func (*Volume) ResourceType() string {
	return "Volume"
}

// Get fetches the given volume by ID
func (vol *Volume) Get(ctx context.Context, client *Client) error {
	if vol.ID == "" {
		return fmt.Errorf("A Volume may only be get using ID")
	}

	vols, err := client.ListWithContext(ctx, vol)
	if err != nil {
		return err
	}

	count := len(vols)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("Volume not found. id: %s", vol.ID),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one Volume was found. Query: id: %s", vol.ID)
	}

	return copier.Copy(vol, vols[0])
}

// ListRequest builds the ListVolumes request
func (vol *Volume) ListRequest() (ListCommand, error) {
	req := &ListVolumes{
		Account:          vol.Account,
		DomainID:         vol.DomainID,
		Name:             vol.Name,
		Type:             vol.Type,
		VirtualMachineID: vol.VirtualMachineID,
		ZoneID:           vol.ZoneID,
	}

	return req, nil
}

// name returns the CloudStack API command name
func (*ResizeVolume) name() string {
	return "resizeVolume"
}

func (*ResizeVolume) asyncResponse() interface{} {
	return new(ResizeVolumeResponse)
}

// ResizeVolumeResponse represents the new Volume
type ResizeVolumeResponse struct {
	Volume Volume `json:"volume"`
}

// name returns the CloudStack API command name
func (*ListVolumes) name() string {
	return "listVolumes"
}

func (*ListVolumes) response() interface{} {
	return new(ListVolumesResponse)
}

// SetPage sets the current page
func (ls *ListVolumes) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListVolumes) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListVolumes) each(resp interface{}, callback IterateItemFunc) {
	volumes, ok := resp.(*ListVolumesResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListVolumesResponse expected, got %t", resp))
		return
	}

	for i := range volumes.Volume {
		if !callback(&volumes.Volume[i], nil) {
			break
		}
	}
}
