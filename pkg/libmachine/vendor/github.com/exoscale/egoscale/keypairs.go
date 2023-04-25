package egoscale

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
)

// Get populates the given SSHKeyPair
func (ssh *SSHKeyPair) Get(ctx context.Context, client *Client) error {
	sshs, err := client.ListWithContext(ctx, ssh)
	if err != nil {
		return err
	}

	count := len(sshs)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("SSHKeyPair not found"),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one SSHKeyPair was found")
	}

	return copier.Copy(ssh, sshs[0])
}

// Delete removes the given SSH key, by Name
func (ssh *SSHKeyPair) Delete(ctx context.Context, client *Client) error {
	if ssh.Name == "" {
		return fmt.Errorf("An SSH Key Pair may only be deleted using Name")
	}

	return client.BooleanRequestWithContext(ctx, &DeleteSSHKeyPair{
		Name:      ssh.Name,
		Account:   ssh.Account,
		DomainID:  ssh.DomainID,
		ProjectID: ssh.ProjectID,
	})
}

// ListRequest builds the ListSSHKeyPairs request
func (ssh *SSHKeyPair) ListRequest() (ListCommand, error) {
	req := &ListSSHKeyPairs{
		Account:     ssh.Account,
		DomainID:    ssh.DomainID,
		Fingerprint: ssh.Fingerprint,
		Name:        ssh.Name,
		ProjectID:   ssh.ProjectID,
	}

	return req, nil
}

// name returns the CloudStack API command name
func (*CreateSSHKeyPair) name() string {
	return "createSSHKeyPair"
}

func (*CreateSSHKeyPair) response() interface{} {
	return new(CreateSSHKeyPairResponse)
}

// name returns the CloudStack API command name
func (*DeleteSSHKeyPair) name() string {
	return "deleteSSHKeyPair"
}

func (*DeleteSSHKeyPair) response() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*RegisterSSHKeyPair) name() string {
	return "registerSSHKeyPair"
}

func (*RegisterSSHKeyPair) response() interface{} {
	return new(RegisterSSHKeyPairResponse)
}

// name returns the CloudStack API command name
func (*ListSSHKeyPairs) name() string {
	return "listSSHKeyPairs"
}

func (*ListSSHKeyPairs) response() interface{} {
	return new(ListSSHKeyPairsResponse)
}

func (*ListSSHKeyPairs) each(resp interface{}, callback IterateItemFunc) {
	sshs, ok := resp.(*ListSSHKeyPairsResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListSSHKeyPairsResponse expected, got %t", resp))
		return
	}

	for i := range sshs.SSHKeyPair {
		if !callback(&sshs.SSHKeyPair[i], nil) {
			break
		}
	}
}

// SetPage sets the current page
func (ls *ListSSHKeyPairs) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListSSHKeyPairs) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// name returns the CloudStack API command name
func (*ResetSSHKeyForVirtualMachine) name() string {
	return "resetSSHKeyForVirtualMachine"
}

func (*ResetSSHKeyForVirtualMachine) asyncResponse() interface{} {
	return new(ResetSSHKeyForVirtualMachineResponse)
}
