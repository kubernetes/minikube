package egoscale

import (
	"net"
)

// VirtualMachine represents a virtual machine
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html
type VirtualMachine struct {
	Account               string            `json:"account,omitempty" doc:"the account associated with the virtual machine"`
	AffinityGroup         []AffinityGroup   `json:"affinitygroup,omitempty" doc:"list of affinity groups associated with the virtual machine"`
	ClusterID             string            `json:"clusterid,omitempty" doc:"the ID of the vm's cluster"`
	ClusterName           string            `json:"clustername,omitempty" doc:"the name of the vm's cluster"`
	CPUNumber             int               `json:"cpunumber,omitempty" doc:"the number of cpu this virtual machine is running with"`
	CPUSpeed              int               `json:"cpuspeed,omitempty" doc:"the speed of each cpu"`
	CPUUsed               string            `json:"cpuused,omitempty" doc:"the amount of the vm's CPU currently used"`
	Created               string            `json:"created,omitempty" doc:"the date when this virtual machine was created"`
	Details               map[string]string `json:"details,omitempty" doc:"Vm details in key/value pairs."`
	DiskIoRead            int64             `json:"diskioread,omitempty" doc:"the read (io) of disk on the vm"`
	DiskIoWrite           int64             `json:"diskiowrite,omitempty" doc:"the write (io) of disk on the vm"`
	DiskKbsRead           int64             `json:"diskkbsread,omitempty" doc:"the read (bytes) of disk on the vm"`
	DiskKbsWrite          int64             `json:"diskkbswrite,omitempty" doc:"the write (bytes) of disk on the vm"`
	DiskOfferingID        string            `json:"diskofferingid,omitempty" doc:"the ID of the disk offering of the virtual machine"`
	DiskOfferingName      string            `json:"diskofferingname,omitempty" doc:"the name of the disk offering of the virtual machine"`
	DisplayName           string            `json:"displayname,omitempty" doc:"user generated name. The name of the virtual machine is returned if no displayname exists."`
	DisplayVM             bool              `json:"displayvm,omitempty" doc:"an optional field whether to the display the vm to the end user or not."`
	Domain                string            `json:"domain,omitempty" doc:"the name of the domain in which the virtual machine exists"`
	DomainID              string            `json:"domainid,omitempty" doc:"the ID of the domain in which the virtual machine exists"`
	ForVirtualNetwork     bool              `json:"forvirtualnetwork,omitempty" doc:"the virtual network for the service offering"`
	Group                 string            `json:"group,omitempty" doc:"the group name of the virtual machine"`
	GroupID               string            `json:"groupid,omitempty" doc:"the group ID of the virtual machine"`
	HAEnable              bool              `json:"haenable,omitempty" doc:"true if high-availability is enabled, false otherwise"`
	HostID                string            `json:"hostid,omitempty" doc:"the ID of the host for the virtual machine"`
	HostName              string            `json:"hostname,omitempty" doc:"the name of the host for the virtual machine"`
	Hypervisor            string            `json:"hypervisor,omitempty" doc:"the hypervisor on which the template runs"`
	ID                    string            `json:"id,omitempty" doc:"the ID of the virtual machine"`
	InstanceName          string            `json:"instancename,omitempty" doc:"instance name of the user vm; this parameter is returned to the ROOT admin only"`
	IsDynamicallyScalable bool              `json:"isdynamicallyscalable,omitempty" doc:"true if vm contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory."`
	IsoDisplayText        string            `json:"isodisplaytext,omitempty" doc:"an alternate display text of the ISO attached to the virtual machine"`
	IsoID                 string            `json:"isoid,omitempty" doc:"the ID of the ISO attached to the virtual machine"`
	IsoName               string            `json:"isoname,omitempty" doc:"the name of the ISO attached to the virtual machine"`
	KeyPair               string            `json:"keypair,omitempty" doc:"ssh key-pair"`
	Memory                int               `json:"memory,omitempty" doc:"the memory allocated for the virtual machine"`
	Name                  string            `json:"name,omitempty" doc:"the name of the virtual machine"`
	NetworkKbsRead        int64             `json:"networkkbsread,omitempty" doc:"the incoming network traffic on the vm"`
	NetworkKbsWrite       int64             `json:"networkkbswrite,omitempty" doc:"the outgoing network traffic on the host"`
	Nic                   []Nic             `json:"nic,omitempty" doc:"the list of nics associated with vm"`
	OsCategoryID          string            `json:"oscategoryid,omitempty" doc:"Os category ID of the virtual machine"`
	Password              string            `json:"password,omitempty" doc:"the password (if exists) of the virtual machine"`
	PasswordEnabled       bool              `json:"passwordenabled,omitempty" doc:"true if the password rest feature is enabled, false otherwise"`
	PCIDevices            []string          `json:"pcidevices,omitempty" doc:"list of PCI devices"`
	PodID                 string            `json:"podid,omitempty" doc:"the ID of the vm's pod"`
	PodName               string            `json:"podname,omitempty" doc:"the name of the vm's pod"`
	Project               string            `json:"project,omitempty" doc:"the project name of the vm"`
	ProjectID             string            `json:"projectid,omitempty" doc:"the project id of the vm"`
	PublicIP              string            `json:"publicip,omitempty" doc:"public IP address id associated with vm via Static nat rule"`
	PublicIPID            string            `json:"publicipid,omitempty" doc:"public IP address id associated with vm via Static nat rule"`
	RootDeviceID          int64             `json:"rootdeviceid,omitempty" doc:"device ID of the root volume"`
	RootDeviceType        string            `json:"rootdevicetype,omitempty" doc:"device type of the root volume"`
	SecurityGroup         []SecurityGroup   `json:"securitygroup,omitempty" doc:"list of security groups associated with the virtual machine"`
	ServiceOfferingID     string            `json:"serviceofferingid,omitempty" doc:"the ID of the service offering of the virtual machine"`
	ServiceOfferingName   string            `json:"serviceofferingname,omitempty" doc:"the name of the service offering of the virtual machine"`
	ServiceState          string            `json:"servicestate,omitempty" doc:"State of the Service from LB rule"`
	State                 string            `json:"state,omitempty" doc:"the state of the virtual machine"`
	Tags                  []ResourceTag     `json:"tags,omitempty" doc:"the list of resource tags associated with vm"`
	TemplateDisplayText   string            `json:"templatedisplaytext,omitempty" doc:" an alternate display text of the template for the virtual machine"`
	TemplateID            string            `json:"templateid,omitempty" doc:"the ID of the template for the virtual machine. A -1 is returned if the virtual machine was created from an ISO file."`
	TemplateName          string            `json:"templatename,omitempty" doc:"the name of the template for the virtual machine"`
	ZoneID                string            `json:"zoneid,omitempty" doc:"the ID of the availablility zone for the virtual machine"`
	ZoneName              string            `json:"zonename,omitempty" doc:"the name of the availability zone for the virtual machine"`
}

// IPToNetwork represents a mapping between ip and networks
type IPToNetwork struct {
	IP        string `json:"ip,omitempty"`
	Ipv6      string `json:"ipv6,omitempty"`
	NetworkID string `json:"networkid,omitempty"`
}

// Password represents an encrypted password
//
// TODO: method to decrypt it, https://cwiki.apache.org/confluence/pages/viewpage.action?pageId=34014652
type Password struct {
	EncryptedPassword string `json:"encryptedpassword"`
}

// VirtualMachineResponse represents a generic Virtual Machine response
type VirtualMachineResponse struct {
	VirtualMachine VirtualMachine `json:"virtualmachine"`
}

// DeployVirtualMachine (Async) represents the machine creation
//
// Regarding the UserData field, the client is responsible to base64 (and probably gzip) it. Doing it within this library would make the integration with other tools, e.g. Terraform harder.
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/deployVirtualMachine.html
type DeployVirtualMachine struct {
	Account            string            `json:"account,omitempty" doc:"an optional account for the virtual machine. Must be used with domainId."`
	AffinityGroupIDs   []string          `json:"affinitygroupids,omitempty" doc:"comma separated list of affinity groups id that are going to be applied to the virtual machine. Mutually exclusive with affinitygroupnames parameter"`
	AffinityGroupNames []string          `json:"affinitygroupnames,omitempty" doc:"comma separated list of affinity groups names that are going to be applied to the virtual machine.Mutually exclusive with affinitygroupids parameter"`
	CustomID           string            `json:"customid,omitempty" doc:"an optional field, in case you want to set a custom id to the resource. Allowed to Root Admins only"`
	DeploymentPlanner  string            `json:"deploymentplanner,omitempty" doc:"Deployment planner to use for vm allocation. Available to ROOT admin only"`
	Details            map[string]string `json:"details,omitempty" doc:"used to specify the custom parameters."`
	DiskOfferingID     string            `json:"diskofferingid,omitempty" doc:"the ID of the disk offering for the virtual machine. If the template is of ISO format, the diskOfferingId is for the root disk volume. Otherwise this parameter is used to indicate the offering for the data disk volume. If the templateId parameter passed is from a Template object, the diskOfferingId refers to a DATA Disk Volume created. If the templateId parameter passed is from an ISO object, the diskOfferingId refers to a ROOT Disk Volume created."`
	DisplayName        string            `json:"displayname,omitempty" doc:"an optional user generated name for the virtual machine"`
	DisplayVM          *bool             `json:"displayvm,omitempty" doc:"an optional field, whether to the display the vm to the end user or not."`
	DomainID           string            `json:"domainid,omitempty" doc:"an optional domainId for the virtual machine. If the account parameter is used, domainId must also be used."`
	Group              string            `json:"group,omitempty" doc:"an optional group for the virtual machine"`
	HostID             string            `json:"hostid,omitempty" doc:"destination Host ID to deploy the VM to - parameter available for root admin only"`
	Hypervisor         string            `json:"hypervisor,omitempty" doc:"the hypervisor on which to deploy the virtual machine"`
	IP4                *bool             `json:"ip4,omitempty" doc:"True to set an IPv4 to the default interface"`
	IP6                *bool             `json:"ip6,omitempty" doc:"True to set an IPv6 to the default interface"`
	IP6Address         net.IP            `json:"ip6address,omitempty" doc:"the ipv6 address for default vm's network"`
	IPAddress          net.IP            `json:"ipaddress,omitempty" doc:"the ip address for default vm's network"`
	IPToNetworkList    []IPToNetwork     `json:"iptonetworklist,omitempty" doc:"ip to network mapping. Can't be specified with networkIds parameter. Example: iptonetworklist[0].ip=10.10.10.11&iptonetworklist[0].ipv6=fc00:1234:5678::abcd&iptonetworklist[0].networkid=uuid - requests to use ip 10.10.10.11 in network id=uuid"`
	Keyboard           string            `json:"keyboard,omitempty" doc:"an optional keyboard device type for the virtual machine. valid value can be one of de,de-ch,es,fi,fr,fr-be,fr-ch,is,it,jp,nl-be,no,pt,uk,us"`
	KeyPair            string            `json:"keypair,omitempty" doc:"name of the ssh key pair used to login to the virtual machine"`
	Name               string            `json:"name,omitempty" doc:"host name for the virtual machine"`
	NetworkIDs         []string          `json:"networkids,omitempty" doc:"list of network ids used by virtual machine. Can't be specified with ipToNetworkList parameter"`
	ProjectID          string            `json:"projectid,omitempty" doc:"Deploy vm for the project"`
	RootDiskSize       int64             `json:"rootdisksize,omitempty" doc:"Optional field to resize root disk on deploy. Value is in GB. Only applies to template-based deployments. Analogous to details[0].rootdisksize, which takes precedence over this parameter if both are provided"`
	SecurityGroupIDs   []string          `json:"securitygroupids,omitempty" doc:"comma separated list of security groups id that going to be applied to the virtual machine. Should be passed only when vm is created from a zone with Basic Network support. Mutually exclusive with securitygroupnames parameter"`
	SecurityGroupNames []string          `json:"securitygroupnames,omitempty" doc:"comma separated list of security groups names that going to be applied to the virtual machine. Should be passed only when vm is created from a zone with Basic Network support. Mutually exclusive with securitygroupids parameter"`
	ServiceOfferingID  string            `json:"serviceofferingid" doc:"the ID of the service offering for the virtual machine"`
	Size               int64             `json:"size,omitempty" doc:"the arbitrary size for the DATADISK volume. Mutually exclusive with diskOfferingId"`
	StartVM            *bool             `json:"startvm,omitempty" doc:"true if start vm after creating. Default value is true"`
	TemplateID         string            `json:"templateid" doc:"the ID of the template for the virtual machine"`
	UserData           string            `json:"userdata,omitempty" doc:"an optional binary data that can be sent to the virtual machine upon a successful deployment. This binary data must be base64 encoded before adding it to the request. Using HTTP GET (via querystring), you can send up to 2KB of data after base64 encoding. Using HTTP POST(via POST body), you can send up to 32K of data after base64 encoding."`
	ZoneID             string            `json:"zoneid" doc:"availability zone for the virtual machine"`
}

// DeployVirtualMachineResponse represents a deployed VM instance
type DeployVirtualMachineResponse VirtualMachineResponse

// StartVirtualMachine (Async) represents the creation of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/startVirtualMachine.html
type StartVirtualMachine struct {
	ID                string `json:"id" doc:"The ID of the virtual machine"`
	DeploymentPlanner string `json:"deploymentplanner,omitempty" doc:"Deployment planner to use for vm allocation. Available to ROOT admin only"`
	HostID            string `json:"hostid,omitempty" doc:"destination Host ID to deploy the VM to - parameter available for root admin only"`
}

// StartVirtualMachineResponse represents a started VM instance
type StartVirtualMachineResponse VirtualMachineResponse

// StopVirtualMachine (Async) represents the stopping of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/stopVirtualMachine.html
type StopVirtualMachine struct {
	ID     string `json:"id" doc:"The ID of the virtual machine"`
	Forced *bool  `json:"forced,omitempty" doc:"Force stop the VM (vm is marked as Stopped even when command fails to be send to the backend).  The caller knows the VM is stopped."`
}

// StopVirtualMachineResponse represents a stopped VM instance
type StopVirtualMachineResponse VirtualMachineResponse

// RebootVirtualMachine (Async) represents the rebooting of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/rebootVirtualMachine.html
type RebootVirtualMachine struct {
	ID string `json:"id" doc:"The ID of the virtual machine"`
}

// RebootVirtualMachineResponse represents a rebooted VM instance
type RebootVirtualMachineResponse VirtualMachineResponse

// RestoreVirtualMachine (Async) represents the restoration of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/restoreVirtualMachine.html
type RestoreVirtualMachine struct {
	VirtualMachineID string `json:"virtualmachineid" doc:"Virtual Machine ID"`
	TemplateID       string `json:"templateid,omitempty" doc:"an optional template Id to restore vm from the new template. This can be an ISO id in case of restore vm deployed using ISO"`
	RootDiskSize     int64  `json:"rootdisksize,omitempty" doc:"Optional field to resize root disk on restore. Value is in GB. Only applies to template-based deployments."`
}

// RestoreVirtualMachineResponse represents a restored VM instance
type RestoreVirtualMachineResponse VirtualMachineResponse

// RecoverVirtualMachine represents the restoration of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/recoverVirtualMachine.html
type RecoverVirtualMachine RebootVirtualMachine

// RecoverVirtualMachineResponse represents a recovered VM instance
type RecoverVirtualMachineResponse VirtualMachineResponse

// DestroyVirtualMachine (Async) represents the destruction of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/destroyVirtualMachine.html
type DestroyVirtualMachine struct {
	ID      string `json:"id" doc:"The ID of the virtual machine"`
	Expunge *bool  `json:"expunge,omitempty" doc:"If true is passed, the vm is expunged immediately. False by default."`
}

// DestroyVirtualMachineResponse represents a destroyed VM instance
type DestroyVirtualMachineResponse VirtualMachineResponse

// UpdateVirtualMachine represents the update of the virtual machine
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/updateVirtualMachine.html
type UpdateVirtualMachine struct {
	ID                    string            `json:"id" doc:"The ID of the virtual machine"`
	CustomID              string            `json:"customid,omitempty" doc:"an optional field, in case you want to set a custom id to the resource. Allowed to Root Admins only"`
	Details               map[string]string `json:"details,omitempty" doc:"Details in key/value pairs."`
	DisplayName           string            `json:"displayname,omitempty" doc:"user generated name"`
	DisplayVM             *bool             `json:"displayvm,omitempty" doc:"an optional field, whether to the display the vm to the end user or not."`
	Group                 string            `json:"group,omitempty" doc:"group of the virtual machine"`
	HAEnable              *bool             `json:"haenable,omitempty" doc:"true if high-availability is enabled for the virtual machine, false otherwise"`
	IsDynamicallyScalable *bool             `json:"isdynamicallyscalable,omitempty" doc:"true if VM contains XS/VMWare tools inorder to support dynamic scaling of VM cpu/memory"`
	Name                  string            `json:"name,omitempty" doc:"new host name of the vm. The VM has to be stopped/started for this update to take affect"`
	SecurityGroupIDs      []string          `json:"securitygroupids,omitempty" doc:"list of security group ids to be applied on the virtual machine."`
	UserData              string            `json:"userdata,omitempty" doc:"an optional binary data that can be sent to the virtual machine upon a successful deployment. This binary data must be base64 encoded before adding it to the request. Using HTTP GET (via querystring), you can send up to 2KB of data after base64 encoding. Using HTTP POST(via POST body), you can send up to 32K of data after base64 encoding."`
}

// UpdateVirtualMachineResponse represents an updated VM instance
type UpdateVirtualMachineResponse VirtualMachineResponse

// ExpungeVirtualMachine represents the annihilation of a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/expungeVirtualMachine.html
type ExpungeVirtualMachine RebootVirtualMachine

// ScaleVirtualMachine (Async) represents the scaling of a VM
//
// ChangeServiceForVirtualMachine does the same thing but returns the
// new Virtual Machine which is more consistent with the rest of the API.
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/scaleVirtualMachine.html
type ScaleVirtualMachine struct {
	ID                string            `json:"id" doc:"The ID of the virtual machine"`
	ServiceOfferingID string            `json:"serviceofferingid" doc:"the ID of the service offering for the virtual machine"`
	Details           map[string]string `json:"details,omitempty" doc:"name value pairs of custom parameters for cpu,memory and cpunumber. example details[i].name=value"`
}

// ChangeServiceForVirtualMachine represents the scaling of a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/changeServiceForVirtualMachine.html
type ChangeServiceForVirtualMachine ScaleVirtualMachine

// ChangeServiceForVirtualMachineResponse represents an changed VM instance
type ChangeServiceForVirtualMachineResponse VirtualMachineResponse

// ResetPasswordForVirtualMachine resets the password for virtual machine. The virtual machine must be in a "Stopped" state...
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/resetPasswordForVirtualMachine.html
type ResetPasswordForVirtualMachine RebootVirtualMachine

// ResetPasswordForVirtualMachineResponse represents the updated vm
type ResetPasswordForVirtualMachineResponse VirtualMachineResponse

// GetVMPassword asks for an encrypted password
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/getVMPassword.html
type GetVMPassword RebootVirtualMachine

// GetVMPasswordResponse represents the encrypted password
type GetVMPasswordResponse struct {
	// Base64 encrypted password for the VM
	Password Password `json:"password"`
}

// ListVirtualMachines represents a search for a VM
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listVirtualMachine.html
type ListVirtualMachines struct {
	Account           string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	AffinityGroupID   string        `json:"affinitygroupid,omitempty" doc:"list vms by affinity group"`
	Details           []string      `json:"details,omitempty" doc:"comma separated list of host details requested, value can be a list of [all, group, nics, stats, secgrp, tmpl, servoff, diskoff, iso, volume, min, affgrp]. If no parameter is passed in, the details will be defaulted to all"`
	DisplayVM         *bool         `json:"displayvm,omitempty" doc:"list resources by display flag; only ROOT admin is eligible to pass this parameter"`
	DomainID          string        `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ForVirtualNetwork *bool         `json:"forvirtualnetwork,omitempty" doc:"list by network type; true if need to list vms using Virtual Network, false otherwise"`
	GroupID           string        `json:"groupid,omitempty" doc:"the group ID"`
	HostID            string        `json:"hostid,omitempty" doc:"the host ID"`
	Hypervisor        string        `json:"hypervisor,omitempty" doc:"the target hypervisor for the template"`
	ID                string        `json:"id,omitempty" doc:"the ID of the virtual machine"`
	IDs               []string      `json:"ids,omitempty" doc:"the IDs of the virtual machines, mutually exclusive with id"`
	IPAddress         net.IP        `json:"ipaddress,omitempty" doc:"an IP address to filter the result"`
	IsoID             string        `json:"isoid,omitempty" doc:"list vms by iso"`
	IsRecursive       *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword           string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll           *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name              string        `json:"name,omitempty" doc:"name of the virtual machine"`
	NetworkID         string        `json:"networkid,omitempty" doc:"list by network id"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	PodID             string        `json:"podid,omitempty" doc:"the pod ID"`
	ProjectID         string        `json:"projectid,omitempty" doc:"list objects by project"`
	ServiceOfferindID string        `json:"serviceofferingid,omitempty" doc:"list by the service offering"`
	State             string        `json:"state,omitempty" doc:"state of the virtual machine"`
	StorageID         string        `json:"storageid,omitempty" doc:"the storage ID where vm's volumes belong to"`
	Tags              []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	TemplateID        string        `json:"templateid,omitempty" doc:"list vms by template"`
	VpcID             string        `json:"vpcid,omitempty" doc:"list vms by vpc"`
	ZoneID            string        `json:"zoneid,omitempty" doc:"the availability zone ID"`
}

// ListVirtualMachinesResponse represents a list of virtual machines
type ListVirtualMachinesResponse struct {
	Count          int              `json:"count"`
	VirtualMachine []VirtualMachine `json:"virtualmachine"`
}

// AddNicToVirtualMachine (Async) adds a NIC to a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/addNicToVirtualMachine.html
type AddNicToVirtualMachine struct {
	NetworkID        string `json:"networkid" doc:"Network ID"`
	VirtualMachineID string `json:"virtualmachineid" doc:"Virtual Machine ID"`
	IPAddress        net.IP `json:"ipaddress,omitempty" doc:"IP Address for the new network"`
}

// AddNicToVirtualMachineResponse represents the modified VM
type AddNicToVirtualMachineResponse VirtualMachineResponse

// RemoveNicFromVirtualMachine (Async) removes a NIC from a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/removeNicFromVirtualMachine.html
type RemoveNicFromVirtualMachine struct {
	NicID            string `json:"nicid" doc:"NIC ID"`
	VirtualMachineID string `json:"virtualmachineid" doc:"Virtual Machine ID"`
}

// RemoveNicFromVirtualMachineResponse represents the modified VM
type RemoveNicFromVirtualMachineResponse VirtualMachineResponse

// UpdateDefaultNicForVirtualMachine (Async) adds a NIC to a VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateDefaultNicForVirtualMachine.html
type UpdateDefaultNicForVirtualMachine RemoveNicFromVirtualMachine

// UpdateDefaultNicForVirtualMachineResponse represents the modified VM
type UpdateDefaultNicForVirtualMachineResponse VirtualMachineResponse

// GetVirtualMachineUserData returns the user-data of the given VM
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/getVirtualMachineUserData.html
type GetVirtualMachineUserData RebootVirtualMachine

// GetVirtualMachineUserDataResponse represents the base64 encoded user-data
type GetVirtualMachineUserDataResponse struct {
	UserData         string `json:"userdata,omitempty" doc:"Base 64 encoded VM user data"`
	VirtualMachineID string `json:"virtualmachineid,omitempty" doc:"the ID of the virtual machine"`
}
