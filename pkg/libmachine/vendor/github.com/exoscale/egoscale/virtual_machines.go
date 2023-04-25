package egoscale

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/jinzhu/copier"
)

// ResourceType returns the type of the resource
func (*VirtualMachine) ResourceType() string {
	return "UserVM"
}

// Get fills the VM
func (vm *VirtualMachine) Get(ctx context.Context, client *Client) error {
	if vm.ID == "" && vm.Name == "" && vm.DefaultNic() == nil {
		return fmt.Errorf("A VirtualMachine may only be searched using ID, Name or IPAddress")
	}

	vms, err := client.ListWithContext(ctx, vm)
	if err != nil {
		return err
	}

	count := len(vms)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("VirtualMachine not found. id: %s, name: %s", vm.ID, vm.Name),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one VirtualMachine was found. Query: id: %s, name: %s", vm.ID, vm.Name)
	}

	return copier.Copy(vm, vms[0])
}

// Delete destroys the VM
func (vm *VirtualMachine) Delete(ctx context.Context, client *Client) error {
	_, err := client.RequestWithContext(ctx, &DestroyVirtualMachine{
		ID: vm.ID,
	})

	return err
}

// ListRequest builds the ListVirtualMachines request
func (vm *VirtualMachine) ListRequest() (ListCommand, error) {
	// XXX: AffinityGroupID, SecurityGroupID, Tags

	req := &ListVirtualMachines{
		Account:    vm.Account,
		DomainID:   vm.DomainID,
		GroupID:    vm.GroupID,
		ID:         vm.ID,
		Name:       vm.Name,
		ProjectID:  vm.ProjectID,
		State:      vm.State,
		TemplateID: vm.TemplateID,
		ZoneID:     vm.ZoneID,
	}

	nic := vm.DefaultNic()
	if nic != nil {
		req.IPAddress = nic.IPAddress
	}

	return req, nil
}

// DefaultNic returns the default nic
func (vm *VirtualMachine) DefaultNic() *Nic {
	for _, nic := range vm.Nic {
		if nic.IsDefault {
			return &nic
		}
	}

	return nil
}

// IP returns the default nic IP address
func (vm *VirtualMachine) IP() *net.IP {
	nic := vm.DefaultNic()
	if nic != nil {
		ip := nic.IPAddress
		return &ip
	}

	return nil
}

// NicsByType returns the corresponding interfaces base on the given type
func (vm *VirtualMachine) NicsByType(nicType string) []Nic {
	nics := make([]Nic, 0)
	for _, nic := range vm.Nic {
		if nic.Type == nicType {
			// XXX The CloudStack API forgets to specify it
			nic.VirtualMachineID = vm.ID
			nics = append(nics, nic)
		}
	}
	return nics
}

// NicByNetworkID returns the corresponding interface based on the given NetworkID
//
// A VM cannot be connected twice to a same network.
func (vm *VirtualMachine) NicByNetworkID(networkID string) *Nic {
	for _, nic := range vm.Nic {
		if nic.NetworkID == networkID {
			nic.VirtualMachineID = vm.ID
			return &nic
		}
	}
	return nil
}

// NicByID returns the corresponding interface base on its ID
func (vm *VirtualMachine) NicByID(nicID string) *Nic {
	for _, nic := range vm.Nic {
		if nic.ID == nicID {
			nic.VirtualMachineID = vm.ID
			return &nic
		}
	}

	return nil
}

// name returns the CloudStack API command name
func (*DeployVirtualMachine) name() string {
	return "deployVirtualMachine"
}

func (req *DeployVirtualMachine) onBeforeSend(params *url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set
	if len(req.AffinityGroupIDs) > 0 && len(req.AffinityGroupNames) > 0 {
		return fmt.Errorf("Either AffinityGroupIDs or AffinityGroupNames must be set")
	}

	// Either SecurityGroupIDs or SecurityGroupNames must be set
	if len(req.SecurityGroupIDs) > 0 && len(req.SecurityGroupNames) > 0 {
		return fmt.Errorf("Either SecurityGroupIDs or SecurityGroupNames must be set")
	}

	return nil
}

func (*DeployVirtualMachine) asyncResponse() interface{} {
	return new(DeployVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*StartVirtualMachine) name() string {
	return "startVirtualMachine"
}
func (*StartVirtualMachine) asyncResponse() interface{} {
	return new(StartVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*StopVirtualMachine) name() string {
	return "stopVirtualMachine"
}

func (*StopVirtualMachine) asyncResponse() interface{} {
	return new(StopVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*RebootVirtualMachine) name() string {
	return "rebootVirtualMachine"
}

func (*RebootVirtualMachine) asyncResponse() interface{} {
	return new(RebootVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*RestoreVirtualMachine) name() string {
	return "restoreVirtualMachine"
}

func (*RestoreVirtualMachine) asyncResponse() interface{} {
	return new(RestoreVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*RecoverVirtualMachine) name() string {
	return "recoverVirtualMachine"
}

func (*RecoverVirtualMachine) response() interface{} {
	return new(RecoverVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*DestroyVirtualMachine) name() string {
	return "destroyVirtualMachine"
}

func (*DestroyVirtualMachine) asyncResponse() interface{} {
	return new(DestroyVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*UpdateVirtualMachine) name() string {
	return "updateVirtualMachine"
}

func (*UpdateVirtualMachine) response() interface{} {
	return new(UpdateVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*ExpungeVirtualMachine) name() string {
	return "expungeVirtualMachine"
}

func (*ExpungeVirtualMachine) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*ScaleVirtualMachine) name() string {
	return "scaleVirtualMachine"
}

func (*ScaleVirtualMachine) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*ChangeServiceForVirtualMachine) name() string {
	return "changeServiceForVirtualMachine"
}

func (*ChangeServiceForVirtualMachine) response() interface{} {
	return new(ChangeServiceForVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*ResetPasswordForVirtualMachine) name() string {
	return "resetPasswordForVirtualMachine"
}

func (*ResetPasswordForVirtualMachine) asyncResponse() interface{} {
	return new(ResetPasswordForVirtualMachineResponse)
}

func (*GetVirtualMachineUserData) name() string {
	return "getVirtualMachineUserData"
}

func (*GetVirtualMachineUserData) response() interface{} {
	return new(GetVirtualMachineUserDataResponse)
}

// name returns the CloudStack API command name
func (*GetVMPassword) name() string {
	return "getVMPassword"
}

func (*GetVMPassword) response() interface{} {
	return new(GetVMPasswordResponse)
}

// name returns the CloudStack API command name
func (*ListVirtualMachines) name() string {
	return "listVirtualMachines"
}

func (*ListVirtualMachines) response() interface{} {
	return new(ListVirtualMachinesResponse)
}

// SetPage sets the current page
func (ls *ListVirtualMachines) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListVirtualMachines) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

func (*ListVirtualMachines) each(resp interface{}, callback IterateItemFunc) {
	vms, ok := resp.(*ListVirtualMachinesResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListVirtualMachinesResponse expected, got %t", resp))
		return
	}

	for i := range vms.VirtualMachine {
		if !callback(&vms.VirtualMachine[i], nil) {
			break
		}
	}
}

// name returns the CloudStack API command name
func (*AddNicToVirtualMachine) name() string {
	return "addNicToVirtualMachine"
}

func (*AddNicToVirtualMachine) asyncResponse() interface{} {
	return new(AddNicToVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*RemoveNicFromVirtualMachine) name() string {
	return "removeNicFromVirtualMachine"
}

func (*RemoveNicFromVirtualMachine) asyncResponse() interface{} {
	return new(RemoveNicFromVirtualMachineResponse)
}

// name returns the CloudStack API command name
func (*UpdateDefaultNicForVirtualMachine) name() string {
	return "updateDefaultNicForVirtualMachine"
}

func (*UpdateDefaultNicForVirtualMachine) asyncResponse() interface{} {
	return new(UpdateDefaultNicForVirtualMachineResponse)
}
