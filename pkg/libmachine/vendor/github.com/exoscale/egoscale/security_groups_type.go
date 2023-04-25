package egoscale

// SecurityGroup represent a firewalling set of rules
type SecurityGroup struct {
	ID                  string        `json:"id"`
	Account             string        `json:"account,omitempty"`
	Description         string        `json:"description,omitempty"`
	Domain              string        `json:"domain,omitempty"`
	DomainID            string        `json:"domainid,omitempty"`
	Name                string        `json:"name"`
	Project             string        `json:"project,omitempty"`
	ProjectID           string        `json:"projectid,omitempty"`
	VirtualMachineCount int           `json:"virtualmachinecount,omitempty"` // CloudStack 4.6+
	VirtualMachineIDs   []string      `json:"virtualmachineids,omitempty"`   // CloudStack 4.6+
	IngressRule         []IngressRule `json:"ingressrule"`
	EgressRule          []EgressRule  `json:"egressrule"`
	Tags                []ResourceTag `json:"tags,omitempty"`
	JobID               string        `json:"jobid,omitempty"`
	JobStatus           JobStatusType `json:"jobstatus,omitempty"`
}

// IngressRule represents the ingress rule
type IngressRule struct {
	RuleID                string              `json:"ruleid"`
	Account               string              `json:"account,omitempty"`
	Cidr                  string              `json:"cidr,omitempty"`
	Description           string              `json:"description,omitempty"`
	IcmpType              uint8               `json:"icmptype,omitempty"`
	IcmpCode              uint8               `json:"icmpcode,omitempty"`
	StartPort             uint16              `json:"startport,omitempty"`
	EndPort               uint16              `json:"endport,omitempty"`
	Protocol              string              `json:"protocol,omitempty"`
	Tags                  []ResourceTag       `json:"tags,omitempty"`
	SecurityGroupID       string              `json:"securitygroupid,omitempty"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
	JobID                 string              `json:"jobid,omitempty"`
	JobStatus             JobStatusType       `json:"jobstatus,omitempty"`
}

// EgressRule represents the ingress rule
type EgressRule IngressRule

// UserSecurityGroup represents the traffic of another security group
type UserSecurityGroup struct {
	Group   string `json:"group,omitempty"`
	Account string `json:"account,omitempty"`
}

// SecurityGroupResponse represents a generic security group response
type SecurityGroupResponse struct {
	SecurityGroup SecurityGroup `json:"securitygroup"`
}

// CreateSecurityGroup represents a security group creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/createSecurityGroup.html
type CreateSecurityGroup struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateSecurityGroupResponse represents a new security group
type CreateSecurityGroupResponse SecurityGroupResponse

// DeleteSecurityGroup represents a security group deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSecurityGroup.html
type DeleteSecurityGroup struct {
	Account   string `json:"account,omitempty"`  // Must be specified with Domain ID
	DomainID  string `json:"domainid,omitempty"` // Must be specified with Account
	ID        string `json:"id,omitempty"`       // Mutually exclusive with name
	Name      string `json:"name,omitempty"`     // Mutually exclusive with id
	ProjectID string `json:"project,omitempty"`
}

// AuthorizeSecurityGroupIngress (Async) represents the ingress rule creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/authorizeSecurityGroupIngress.html
type AuthorizeSecurityGroupIngress struct {
	Account               string              `json:"account,omitempty"`
	CidrList              []string            `json:"cidrlist,omitempty"`
	Description           string              `json:"description,omitempty"`
	DomainID              string              `json:"domainid,omitempty"`
	IcmpType              uint8               `json:"icmptype,omitempty"`
	IcmpCode              uint8               `json:"icmpcode,omitempty"`
	StartPort             uint16              `json:"startport,omitempty"`
	EndPort               uint16              `json:"endport,omitempty"`
	ProjectID             string              `json:"projectid,omitempty"`
	Protocol              string              `json:"protocol,omitempty"`
	SecurityGroupID       string              `json:"securitygroupid,omitempty"`
	SecurityGroupName     string              `json:"securitygroupname,omitempty"`
	UserSecurityGroupList []UserSecurityGroup `json:"usersecuritygrouplist,omitempty"`
}

// AuthorizeSecurityGroupIngressResponse represents the new egress rule
// /!\ the Cloud Stack API document is not fully accurate. /!\
type AuthorizeSecurityGroupIngressResponse SecurityGroupResponse

// AuthorizeSecurityGroupEgress (Async) represents the egress rule creation
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/authorizeSecurityGroupEgress.html
type AuthorizeSecurityGroupEgress AuthorizeSecurityGroupIngress

// AuthorizeSecurityGroupEgressResponse represents the new egress rule
// /!\ the Cloud Stack API document is not fully accurate. /!\
type AuthorizeSecurityGroupEgressResponse CreateSecurityGroupResponse

// RevokeSecurityGroupIngress (Async) represents the ingress/egress rule deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/revokeSecurityGroupIngress.html
type RevokeSecurityGroupIngress struct {
	ID string `json:"id"`
}

// RevokeSecurityGroupEgress (Async) represents the ingress/egress rule deletion
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/revokeSecurityGroupEgress.html
type RevokeSecurityGroupEgress struct {
	ID string `json:"id"`
}

// ListSecurityGroups represents a search for security groups
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listSecurityGroups.html
type ListSecurityGroups struct {
	Account           string        `json:"account,omitempty"`
	DomainID          string        `json:"domainid,omitempty"`
	ID                string        `json:"id,omitempty"`
	IsRecursive       *bool         `json:"isrecursive,omitempty"`
	Keyword           string        `json:"keyword,omitempty"`
	ListAll           *bool         `json:"listall,omitempty"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	ProjectID         string        `json:"projectid,omitempty"`
	Type              string        `json:"type,omitempty"`
	SecurityGroupName string        `json:"securitygroupname,omitempty"`
	Tags              []ResourceTag `json:"tags,omitempty"`
	VirtualMachineID  string        `json:"virtualmachineid,omitempty"`
}

// ListSecurityGroupsResponse represents a list of security groups
type ListSecurityGroupsResponse struct {
	Count         int             `json:"count"`
	SecurityGroup []SecurityGroup `json:"securitygroup"`
}
