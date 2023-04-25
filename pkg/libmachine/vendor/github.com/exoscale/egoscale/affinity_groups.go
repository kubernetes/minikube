package egoscale

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jinzhu/copier"
)

// AffinityGroup represents an (anti-)affinity group
//
// Affinity and Anti-Affinity groups provide a way to influence where VMs should run.
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#affinity-groups
type AffinityGroup struct {
	ID                string   `json:"id,omitempty"`
	Account           string   `json:"account,omitempty"`
	Description       string   `json:"description,omitempty"`
	Domain            string   `json:"domain,omitempty"`
	DomainID          string   `json:"domainid,omitempty"`
	Name              string   `json:"name,omitempty"`
	Project           string   `json:"project,omitempty"`
	ProjectID         string   `json:"projectid,omitempty"`
	Type              string   `json:"type,omitempty"`
	VirtualMachineIDs []string `json:"virtualmachineIds,omitempty"` // *I*ds is not a typo
}

// AffinityGroupType represent an affinity group type
type AffinityGroupType struct {
	Type string `json:"type"`
}

// Get loads the given Affinity Group
func (ag *AffinityGroup) Get(ctx context.Context, client *Client) error {
	if ag.ID == "" && ag.Name == "" {
		return fmt.Errorf("An Affinity Group may only be searched using ID or Name")
	}

	resp, err := client.RequestWithContext(ctx, &ListAffinityGroups{
		ID:   ag.ID,
		Name: ag.Name,
	})

	if err != nil {
		return err
	}

	ags := resp.(*ListAffinityGroupsResponse)
	count := len(ags.AffinityGroup)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("AffinityGroup not found id: %s, name: %s", ag.ID, ag.Name),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one Affinity Group was found. Query; id: %s, name: %s", ag.ID, ag.Name)
	}

	return copier.Copy(ag, ags.AffinityGroup[0])
}

// Delete removes the given Affinity Group
func (ag *AffinityGroup) Delete(ctx context.Context, client *Client) error {
	if ag.ID == "" && ag.Name == "" {
		return fmt.Errorf("An Affinity Group may only be deleted using ID or Name")
	}

	req := &DeleteAffinityGroup{
		Account:   ag.Account,
		DomainID:  ag.DomainID,
		ProjectID: ag.ProjectID,
	}

	if ag.ID != "" {
		req.ID = ag.ID
	} else {
		req.Name = ag.Name
	}

	return client.BooleanRequestWithContext(ctx, req)
}

// CreateAffinityGroup (Async) represents a new (anti-)affinity group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createAffinityGroup.html
type CreateAffinityGroup struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Account     string `json:"account,omitempty"`
	Description string `json:"description,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
}

// name returns the CloudStack API command name
func (*CreateAffinityGroup) name() string {
	return "createAffinityGroup"
}

func (*CreateAffinityGroup) asyncResponse() interface{} {
	return new(CreateAffinityGroupResponse)
}

// CreateAffinityGroupResponse represents the response of the creation of an (anti-)affinity group
type CreateAffinityGroupResponse struct {
	AffinityGroup AffinityGroup `json:"affinitygroup"`
}

// UpdateVMAffinityGroup (Async) represents a modification of a (anti-)affinity group
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateVMAffinityGroup.html
type UpdateVMAffinityGroup struct {
	ID                 string   `json:"id"`
	AffinityGroupIDs   []string `json:"affinitygroupids,omitempty"`   // mutually exclusive with names
	AffinityGroupNames []string `json:"affinitygroupnames,omitempty"` // mutually exclusive with ids
}

// name returns the CloudStack API command name
func (*UpdateVMAffinityGroup) name() string {
	return "updateVMAffinityGroup"
}

func (*UpdateVMAffinityGroup) asyncResponse() interface{} {
	return new(UpdateVMAffinityGroupResponse)
}

func (req *UpdateVMAffinityGroup) onBeforeSend(params *url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set
	if len(req.AffinityGroupIDs) == 0 && len(req.AffinityGroupNames) == 0 {
		params.Set("affinitygroupids", "")
	}
	return nil
}

// UpdateVMAffinityGroupResponse represents the new VM
type UpdateVMAffinityGroupResponse VirtualMachineResponse

// DeleteAffinityGroup (Async) represents an (anti-)affinity group to be deleted
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteAffinityGroup.html
type DeleteAffinityGroup struct {
	Account   string `json:"account,omitempty"` // must be specified with DomainID
	DomainID  string `json:"domainid,omitempty"`
	ID        string `json:"id,omitempty"`   // mutually exclusive with Name
	Name      string `json:"name,omitempty"` // mutually exclusive with ID
	ProjectID string `json:"projectid,omitempty"`
}

// name returns the CloudStack API command name
func (*DeleteAffinityGroup) name() string {
	return "deleteAffinityGroup"
}

func (*DeleteAffinityGroup) asyncResponse() interface{} {
	return new(booleanResponse)
}

// ListAffinityGroups represents an (anti-)affinity groups search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listAffinityGroups.html
type ListAffinityGroups struct {
	Account          string `json:"account,omitempty"`
	DomainID         string `json:"domainid,omitempty"`
	ID               string `json:"id,omitempty"`
	IsRecursive      *bool  `json:"isrecursive,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	ListAll          *bool  `json:"listall,omitempty"`
	Name             string `json:"name,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	Type             string `json:"type,omitempty"`
	VirtualMachineID string `json:"virtualmachineid,omitempty"`
}

// name returns the CloudStack API command name
func (*ListAffinityGroups) name() string {
	return "listAffinityGroups"
}

func (*ListAffinityGroups) response() interface{} {
	return new(ListAffinityGroupsResponse)
}

// ListAffinityGroupTypes represents an (anti-)affinity groups search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listAffinityGroupTypes.html
type ListAffinityGroupTypes struct {
	Keyword  string `json:"keyword,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pagesize,omitempty"`
}

// name returns the CloudStack API command name
func (*ListAffinityGroupTypes) name() string {
	return "listAffinityGroupTypes"
}

func (*ListAffinityGroupTypes) response() interface{} {
	return new(ListAffinityGroupTypesResponse)
}

// ListAffinityGroupsResponse represents a list of (anti-)affinity groups
type ListAffinityGroupsResponse struct {
	Count         int             `json:"count"`
	AffinityGroup []AffinityGroup `json:"affinitygroup"`
}

// ListAffinityGroupTypesResponse represents a list of (anti-)affinity group types
type ListAffinityGroupTypesResponse struct {
	Count             int                 `json:"count"`
	AffinityGroupType []AffinityGroupType `json:"affinitygrouptype"`
}
