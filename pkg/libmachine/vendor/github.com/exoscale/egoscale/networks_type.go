package egoscale

import "net"

// Network represents a network
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/networking_and_traffic.html
type Network struct {
	Account                     string        `json:"account,omitempty" doc:"the owner of the network"`
	ACLID                       string        `json:"aclid,omitempty" doc:"ACL Id associated with the VPC network"`
	ACLType                     string        `json:"acltype,omitempty" doc:"acl type - access type to the network"`
	BroadcastDomainType         string        `json:"broadcastdomaintype,omitempty" doc:"Broadcast domain type of the network"`
	BroadcastURI                string        `json:"broadcasturi,omitempty" doc:"broadcast uri of the network. This parameter is visible to ROOT admins only"`
	CanUseForDeploy             bool          `json:"canusefordeploy,omitempty" doc:"list networks available for vm deployment"`
	Cidr                        string        `json:"cidr,omitempty" doc:"Cloudstack managed address space, all CloudStack managed VMs get IP address from CIDR"`
	DisplayNetwork              bool          `json:"displaynetwork,omitempty" doc:"an optional field, whether to the display the network to the end user or not."`
	DisplayText                 string        `json:"displaytext,omitempty" doc:"the displaytext of the network"`
	DNS1                        net.IP        `json:"dns1,omitempty" doc:"the first DNS for the network"`
	DNS2                        net.IP        `json:"dns2,omitempty" doc:"the second DNS for the network"`
	Domain                      string        `json:"domain,omitempty" doc:"the domain name of the network owner"`
	DomainID                    string        `json:"domainid,omitempty" doc:"the domain id of the network owner"`
	Gateway                     net.IP        `json:"gateway,omitempty" doc:"the network's gateway"`
	ID                          string        `json:"id,omitempty" doc:"the id of the network"`
	IP6Cidr                     string        `json:"ip6cidr,omitempty" doc:"the cidr of IPv6 network"`
	IP6Gateway                  net.IP        `json:"ip6gateway,omitempty" doc:"the gateway of IPv6 network"`
	IsDefault                   bool          `json:"isdefault,omitempty" doc:"true if network is default, false otherwise"`
	IsPersistent                bool          `json:"ispersistent,omitempty" doc:"list networks that are persistent"`
	IsSystem                    bool          `json:"issystem,omitempty" doc:"true if network is system, false otherwise"`
	Name                        string        `json:"name,omitempty" doc:"the name of the network"`
	Netmask                     net.IP        `json:"netmask,omitempty" doc:"the network's netmask"`
	NetworkCidr                 string        `json:"networkcidr,omitempty" doc:"the network CIDR of the guest network configured with IP reservation. It is the summation of CIDR and RESERVED_IP_RANGE"`
	NetworkDomain               string        `json:"networkdomain,omitempty" doc:"the network domain"`
	NetworkOfferingAvailability string        `json:"networkofferingavailability,omitempty" doc:"availability of the network offering the network is created from"`
	NetworkOfferingConserveMode bool          `json:"networkofferingconservemode,omitempty" doc:"true if network offering is ip conserve mode enabled"`
	NetworkOfferingDisplayText  string        `json:"networkofferingdisplaytext,omitempty" doc:"display text of the network offering the network is created from"`
	NetworkOfferingID           string        `json:"networkofferingid,omitempty" doc:"network offering id the network is created from"`
	NetworkOfferingName         string        `json:"networkofferingname,omitempty" doc:"name of the network offering the network is created from"`
	PhysicalNetworkID           string        `json:"physicalnetworkid,omitempty" doc:"the physical network id"`
	Project                     string        `json:"project,omitempty" doc:"the project name of the address"`
	ProjectID                   string        `json:"projectid,omitempty" doc:"the project id of the ipaddress"`
	Related                     string        `json:"related,omitempty" doc:"related to what other network configuration"`
	ReservedIPRange             string        `json:"reservediprange,omitempty" doc:"the network's IP range not to be used by CloudStack guest VMs and can be used for non CloudStack purposes"`
	RestartRequired             bool          `json:"restartrequired,omitempty" doc:"true network requires restart"`
	Service                     []Service     `json:"service,omitempty" doc:"the list of services"`
	SpecifyIPRanges             bool          `json:"specifyipranges,omitempty" doc:"true if network supports specifying ip ranges, false otherwise"`
	State                       string        `json:"state,omitempty" doc:"state of the network"`
	StrechedL2Subnet            bool          `json:"strechedl2subnet,omitempty" doc:"true if network can span multiple zones"`
	SubdomainAccess             bool          `json:"subdomainaccess,omitempty" doc:"true if users from subdomains can access the domain level network"`
	Tags                        []ResourceTag `json:"tags,omitempty" doc:"the list of resource tags associated with network"`
	TrafficType                 string        `json:"traffictype,omitempty" doc:"the traffic type of the network"`
	Type                        string        `json:"type,omitempty" doc:"the type of the network"`
	Vlan                        string        `json:"vlan,omitemtpy" doc:"The vlan of the network. This parameter is visible to ROOT admins only"`
	VpcID                       string        `json:"vpcid,omitempty" doc:"VPC the network belongs to"`
	ZoneID                      string        `json:"zoneid,omitempty" doc:"zone id of the network"`
	ZoneName                    string        `json:"zonename,omitempty" doc:"the name of the zone the network belongs to"`
	ZonesNetworkSpans           []Zone        `json:"zonesnetworkspans,omitempty" doc:"If a network is enabled for 'streched l2 subnet' then represents zones on which network currently spans"`
}

// Service is a feature of a network
type Service struct {
	Capability []ServiceCapability `json:"capability,omitempty"`
	Name       string              `json:"name"`
	Provider   []ServiceProvider   `json:"provider,omitempty"`
}

// ServiceCapability represents optional capability of a service
type ServiceCapability struct {
	CanChooseServiceCapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

// ServiceProvider represents the provider of the service
type ServiceProvider struct {
	CanEnableIndividualService   bool     `json:"canenableindividualservice"`
	DestinationPhysicalNetworkID string   `json:"destinationphysicalnetworkid"`
	ID                           string   `json:"id"`
	Name                         string   `json:"name"`
	PhysicalNetworkID            string   `json:"physicalnetworkid"`
	ServiceList                  []string `json:"servicelist,omitempty"`
}

// NetworkResponse represents a network
type NetworkResponse struct {
	Network Network `json:"network"`
}

// CreateNetwork creates a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createNetwork.html
type CreateNetwork struct {
	Account           string `json:"account,omitempty" doc:"account who will own the network"`
	ACLID             string `json:"aclid,omitempty" doc:"Network ACL Id associated for the network"`
	ACLType           string `json:"acltype,omitempty" doc:"Access control type; supported values are account and domain. In 3.0 all shared networks should have aclType=Domain, and all Isolated networks - Account. Account means that only the account owner can use the network, domain - all accouns in the domain can use the network"`
	DisplayNetwork    *bool  `json:"displaynetwork,omitempty" doc:"an optional field, whether to the display the network to the end user or not."`
	DisplayText       string `json:"displaytext" doc:"the display text of the network"`
	DomainID          string `json:"domainid,omitempty" doc:"domain ID of the account owning a network"`
	EndIP             net.IP `json:"endip,omitempty" doc:"the ending IP address in the network IP range. If not specified, will be defaulted to startIP"`
	EndIpv6           net.IP `json:"endipv6,omitempty" doc:"the ending IPv6 address in the IPv6 network range"`
	Gateway           net.IP `json:"gateway,omitempty" doc:"the gateway of the network. Required for Shared networks and Isolated networks when it belongs to VPC"`
	IP6Cidr           string `json:"ip6cidr,omitempty" doc:"the CIDR of IPv6 network, must be at least /64"`
	IP6Gateway        net.IP `json:"ip6gateway,omitempty" doc:"the gateway of the IPv6 network. Required for Shared networks and Isolated networks when it belongs to VPC"`
	IsolatedPVlan     string `json:"isolatedpvlan,omitempty" doc:"the isolated private vlan for this network"`
	Name              string `json:"name" doc:"the name of the network"`
	Netmask           net.IP `json:"netmask,omitempty" doc:"the netmask of the network. Required for Shared networks and Isolated networks when it belongs to VPC"`
	NetworkDomain     string `json:"networkdomain,omitempty" doc:"network domain"`
	NetworkOfferingID string `json:"networkofferingid" doc:"the network offering id"`
	PhysicalNetworkID string `json:"physicalnetworkid,omitempty" doc:"the Physical Network ID the network belongs to"`
	ProjectID         string `json:"projectid,omitempty" doc:"an optional project for the ssh key"`
	StartIP           net.IP `json:"startip,omitempty" doc:"the beginning IP address in the network IP range"`
	StartIpv6         net.IP `json:"startipv6,omitempty" doc:"the beginning IPv6 address in the IPv6 network range"`
	SubdomainAccess   *bool  `json:"subdomainaccess,omitempty" doc:"Defines whether to allow subdomains to use networks dedicated to their parent domain(s). Should be used with aclType=Domain, defaulted to allow.subdomain.network.access global config if not specified"`
	Vlan              string `json:"vlan,omitempty" doc:"the ID or VID of the network"`
	VpcID             string `json:"vpcid,omitempty" doc:"the VPC network belongs to"`
	ZoneID            string `json:"zoneid" doc:"the Zone ID for the network"`
}

// CreateNetworkResponse represents a freshly created network
type CreateNetworkResponse NetworkResponse

// UpdateNetwork (Async) updates a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateNetwork.html
type UpdateNetwork struct {
	ID                string `json:"id" doc:"the ID of the network"`
	ChangeCidr        *bool  `json:"changecidr,omitempty" doc:"Force update even if cidr type is different"`
	CustomID          string `json:"customid,omitempty" doc:"an optional field, in case you want to set a custom id to the resource. Allowed to Root Admins only"`
	DisplayNetwork    *bool  `json:"displaynetwork,omitempty" doc:"an optional field, whether to the display the network to the end user or not."`
	DisplayText       string `json:"displaytext,omitempty" doc:"the new display text for the network"`
	GuestVMCidr       string `json:"guestvmcidr,omitempty" doc:"CIDR for Guest VMs,Cloudstack allocates IPs to Guest VMs only from this CIDR"`
	Name              string `json:"name,omitempty" doc:"the new name for the network"`
	NetworkDomain     string `json:"networkdomain,omitempty" doc:"network domain"`
	NetworkOfferingID string `json:"networkofferingid,omitempty" doc:"network offering ID"`
}

// UpdateNetworkResponse represents a freshly created network
type UpdateNetworkResponse NetworkResponse

// RestartNetwork (Async) updates a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/restartNetwork.html
type RestartNetwork struct {
	ID      string `json:"id" doc:"The id of the network to restart."`
	Cleanup *bool  `json:"cleanup,omitempty" doc:"If cleanup old network elements"`
}

// RestartNetworkResponse represents a freshly created network
type RestartNetworkResponse NetworkResponse

// DeleteNetwork deletes a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteNetwork.html
type DeleteNetwork struct {
	ID     string `json:"id" doc:"the ID of the network"`
	Forced *bool  `json:"forced,omitempty" doc:"Force delete a network. Network will be marked as 'Destroy' even when commands to shutdown and cleanup to the backend fails."`
}

// ListNetworks represents a query to a network
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listNetworks.html
type ListNetworks struct {
	Account           string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	ACLType           string        `json:"acltype,omitempty" doc:"list networks by ACL (access control list) type. Supported values are Account and Domain"`
	CanUseForDeploy   *bool         `json:"canusefordeploy,omitempty" doc:"list networks available for vm deployment"`
	DisplayNetwork    *bool         `json:"displaynetwork,omitempty" doc:"list resources by display flag; only ROOT admin is eligible to pass this parameter"`
	DomainID          string        `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ForVpc            *bool         `json:"forvpc,omitempty" doc:"the network belongs to vpc"`
	ID                string        `json:"id,omitempty" doc:"list networks by id"`
	IsRecursive       *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	IsSystem          *bool         `json:"issystem,omitempty" doc:"true if network is system, false otherwise"`
	Keyword           string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll           *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page              int           `json:"page,omitempty"`
	PageSize          int           `json:"pagesize,omitempty"`
	PhysicalNetworkID string        `json:"physicalnetworkid,omitempty" doc:"list networks by physical network id"`
	ProjectID         string        `json:"projectid,omitempty" doc:"list objects by project"`
	RestartRequired   *bool         `json:"restartrequired,omitempty" doc:"list networks by restartRequired"`
	SpecifyIPRanges   *bool         `json:"specifyipranges,omitempty" doc:"true if need to list only networks which support specifying ip ranges"`
	SupportedServices []Service     `json:"supportedservices,omitempty" doc:"list networks supporting certain services"`
	Tags              []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	TrafficType       string        `json:"traffictype,omitempty" doc:"type of the traffic"`
	Type              string        `json:"type,omitempty" doc:"the type of the network. Supported values are: Isolated and Shared"`
	VpcID             string        `json:"vpcid,omitempty" doc:"List networks by VPC"`
	ZoneID            string        `json:"zoneid,omitempty" doc:"the Zone ID of the network"`
}

// ListNetworksResponse represents the list of networks
type ListNetworksResponse struct {
	Count   int       `json:"count"`
	Network []Network `json:"network"`
}
