package egoscale

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"
)

// ResourceType returns the type of the resource
func (*SecurityGroup) ResourceType() string {
	return "SecurityGroup"
}

// Get loads the given Security Group
func (sg *SecurityGroup) Get(ctx context.Context, client *Client) error {
	if sg.ID == "" && sg.Name == "" {
		return fmt.Errorf("A SecurityGroup may only be searched using ID or Name")
	}

	resp, err := client.List(&SecurityGroup{Name: sg.Name, ID: sg.ID})
	if err != nil {
		return err
	}

	listSize := len(resp)

	if listSize == 0 {
		err := &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("SecurityGroup not found id: %s, name: %s", sg.ID, sg.Name),
		}
		return err
	} else if listSize > 1 {
		return fmt.Errorf("More than one SecurityGroup was found. Query: id: %s, name: %s", sg.ID, sg.Name)
	}

	securGroup, ok := resp[0].(*SecurityGroup)
	if !ok {
		return fmt.Errorf("SecurityGroup expected, got %t", resp[0])
	}

	return copier.Copy(sg, securGroup)
}

// ListRequest builds the ListSecurityGroups request
func (sg *SecurityGroup) ListRequest() (ListCommand, error) {
	//TODO add tags
	req := &ListSecurityGroups{
		Account:           sg.Account,
		DomainID:          sg.DomainID,
		ID:                sg.ID,
		ProjectID:         sg.ProjectID,
		SecurityGroupName: sg.Name,
	}

	return req, nil
}

// Delete deletes the given Security Group
func (sg *SecurityGroup) Delete(ctx context.Context, client *Client) error {
	if sg.ID == "" && sg.Name == "" {
		return fmt.Errorf("A SecurityGroup may only be deleted using ID or Name")
	}

	req := &DeleteSecurityGroup{
		Account:   sg.Account,
		DomainID:  sg.DomainID,
		ProjectID: sg.ProjectID,
	}

	if sg.ID != "" {
		req.ID = sg.ID
	} else {
		req.Name = sg.Name
	}

	return client.BooleanRequestWithContext(ctx, req)
}

// RuleByID returns IngressRule or EgressRule by a rule ID
func (sg *SecurityGroup) RuleByID(ruleID string) (*IngressRule, *EgressRule) {
	for i, in := range sg.IngressRule {
		if ruleID == in.RuleID {
			return &sg.IngressRule[i], nil
		}
	}

	for i, out := range sg.EgressRule {
		if ruleID == out.RuleID {
			return nil, &sg.EgressRule[i]
		}
	}

	return nil, nil
}

// name returns the CloudStack API command name
func (*CreateSecurityGroup) name() string {
	return "createSecurityGroup"
}

func (*CreateSecurityGroup) response() interface{} {
	return new(CreateSecurityGroupResponse)
}

// name returns the CloudStack API command name
func (*DeleteSecurityGroup) name() string {
	return "deleteSecurityGroup"
}

func (*DeleteSecurityGroup) response() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*AuthorizeSecurityGroupIngress) name() string {
	return "authorizeSecurityGroupIngress"
}

func (*AuthorizeSecurityGroupIngress) asyncResponse() interface{} {
	return new(AuthorizeSecurityGroupIngressResponse)
}

func (req *AuthorizeSecurityGroupIngress) onBeforeSend(params *url.Values) error {
	// ICMP code and type may be zero but can also be omitted...
	if strings.HasPrefix(strings.ToLower(req.Protocol), "icmp") {
		params.Set("icmpcode", strconv.FormatInt(int64(req.IcmpCode), 10))
		params.Set("icmptype", strconv.FormatInt(int64(req.IcmpType), 10))
	}
	// StartPort may be zero but can also be omitted...
	if req.EndPort != 0 && req.StartPort == 0 {
		params.Set("startport", "0")
	}
	return nil
}

// name returns the CloudStack API command name
func (*AuthorizeSecurityGroupEgress) name() string {
	return "authorizeSecurityGroupEgress"
}

func (*AuthorizeSecurityGroupEgress) asyncResponse() interface{} {
	return new(AuthorizeSecurityGroupEgressResponse)
}

func (req *AuthorizeSecurityGroupEgress) onBeforeSend(params *url.Values) error {
	return (*AuthorizeSecurityGroupIngress)(req).onBeforeSend(params)
}

// name returns the CloudStack API command name
func (*RevokeSecurityGroupIngress) name() string {
	return "revokeSecurityGroupIngress"
}

func (*RevokeSecurityGroupIngress) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*RevokeSecurityGroupEgress) name() string {
	return "revokeSecurityGroupEgress"
}

func (*RevokeSecurityGroupEgress) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*ListSecurityGroups) name() string {
	return "listSecurityGroups"
}

func (*ListSecurityGroups) response() interface{} {
	return new(ListSecurityGroupsResponse)
}

// SetPage sets the current page
func (lsg *ListSecurityGroups) SetPage(page int) {
	lsg.Page = page
}

// SetPageSize sets the page size
func (lsg *ListSecurityGroups) SetPageSize(pageSize int) {
	lsg.PageSize = pageSize
}

func (*ListSecurityGroups) each(resp interface{}, callback IterateItemFunc) {
	sgs, ok := resp.(*ListSecurityGroupsResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListSecurityGroupsResponse expected, got %t", resp))
		return
	}

	for i := range sgs.SecurityGroup {
		if !callback(&sgs.SecurityGroup[i], nil) {
			break
		}
	}
}
