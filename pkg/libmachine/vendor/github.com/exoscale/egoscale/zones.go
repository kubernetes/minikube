package egoscale

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
)

// Get fetches the given zone by ID or Name
func (zone *Zone) Get(ctx context.Context, client *Client) error {
	if zone.ID == "" && zone.Name == "" {
		return fmt.Errorf("A Zone may only be obtained using ID or Name")
	}

	zones, err := client.ListWithContext(ctx, zone)
	if err != nil {
		return err
	}

	count := len(zones)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("Zone not found. id: %q, name: %q", zone.ID, zone.Name),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one Zone was found. Query: id: %s", zone.ID)
	}

	return copier.Copy(zone, zones[0])
}

// ListRequest builds the ListZones request
func (zone *Zone) ListRequest() (ListCommand, error) {
	req := &ListZones{
		DomainID: zone.DomainID,
		ID:       zone.ID,
		Name:     zone.Name,
	}

	return req, nil
}

// name returns the CloudStack API command name
func (*ListZones) name() string {
	return "listZones"
}

func (*ListZones) response() interface{} {
	return new(ListZonesResponse)
}

// SetPage sets the current page
func (ls *ListZones) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListZones) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListZones) each(resp interface{}, callback IterateItemFunc) {
	zones, ok := resp.(*ListZonesResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListZonesResponse expected, got %t", resp))
		return
	}

	for i := range zones.Zone {
		if !callback(&zones.Zone[i], nil) {
			break
		}
	}
}
