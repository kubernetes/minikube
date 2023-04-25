package egoscale

// SSHKeyPair represents an SSH key pair
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#creating-the-ssh-keypair
type SSHKeyPair struct {
	Account     string `json:"account,omitempty"` // must be used with a Domain ID
	DomainID    string `json:"domainid,omitempty"`
	ProjectID   string `json:"projectid,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Name        string `json:"name,omitempty"`
	PrivateKey  string `json:"privatekey,omitempty"`
}

// CreateSSHKeyPair represents a new keypair to be created
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createSSHKeyPair.html
type CreateSSHKeyPair struct {
	Name      string `json:"name" doc:"Name of the keypair"`
	Account   string `json:"account,omitempty" doc:"an optional account for the ssh key. Must be used with domainId."`
	DomainID  string `json:"domainid,omitempty" doc:"an optional domainId for the ssh key. If the account parameter is used, domainId must also be used."`
	ProjectID string `json:"projectid,omitempty" doc:"an optional project for the ssh key"`
}

// CreateSSHKeyPairResponse represents the creation of an SSH Key Pair
type CreateSSHKeyPairResponse struct {
	KeyPair SSHKeyPair `json:"keypair"`
}

// DeleteSSHKeyPair represents a new keypair to be created
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSSHKeyPair.html
type DeleteSSHKeyPair struct {
	Name      string `json:"name" doc:"Name of the keypair"`
	Account   string `json:"account,omitempty" doc:"the account associated with the keypair. Must be used with the domainId parameter."`
	DomainID  string `json:"domainid,omitempty" doc:"the domain ID associated with the keypair"`
	ProjectID string `json:"projectid,omitempty" doc:"the project associated with keypair"`
}

// RegisterSSHKeyPair represents a new registration of a public key in a keypair
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/registerSSHKeyPair.html
type RegisterSSHKeyPair struct {
	Name      string `json:"name" doc:"Name of the keypair"`
	PublicKey string `json:"publickey" doc:"Public key material of the keypair"`
	Account   string `json:"account,omitempty" doc:"an optional account for the ssh key. Must be used with domainId."`
	DomainID  string `json:"domainid,omitempty" doc:"an optional domainId for the ssh key. If the account parameter is used, domainId must also be used."`
	ProjectID string `json:"projectid,omitempty" doc:"an optional project for the ssh key"`
}

// RegisterSSHKeyPairResponse represents the creation of an SSH Key Pair
type RegisterSSHKeyPairResponse struct {
	KeyPair SSHKeyPair `json:"keypair"`
}

// ListSSHKeyPairs represents a query for a list of SSH KeyPairs
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listSSHKeyPairs.html
type ListSSHKeyPairs struct {
	Account     string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID    string `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	Fingerprint string `json:"fingerprint,omitempty" doc:"A public key fingerprint to look for"`
	IsRecursive *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword     string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll     *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name        string `json:"name,omitempty" doc:"A key pair name to look for"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	ProjectID   string `json:"projectid,omitempty" doc:"list objects by project"`
}

// ListSSHKeyPairsResponse represents a list of SSH key pairs
type ListSSHKeyPairsResponse struct {
	Count      int          `json:"count"`
	SSHKeyPair []SSHKeyPair `json:"sshkeypair"`
}

// ResetSSHKeyForVirtualMachine (Async) represents a change for the key pairs
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/resetSSHKeyForVirtualMachine.html
type ResetSSHKeyForVirtualMachine struct {
	ID        string `json:"id" doc:"The ID of the virtual machine"`
	KeyPair   string `json:"keypair" doc:"name of the ssh key pair used to login to the virtual machine"`
	Account   string `json:"account,omitempty" doc:"an optional account for the ssh key. Must be used with domainId."`
	DomainID  string `json:"domainid,omitempty" doc:"an optional domainId for the virtual machine. If the account parameter is used, domainId must also be used."`
	ProjectID string `json:"projectid,omitempty" doc:"an optional project for the ssh key"`
}

// ResetSSHKeyForVirtualMachineResponse represents the modified VirtualMachine
type ResetSSHKeyForVirtualMachineResponse VirtualMachineResponse
