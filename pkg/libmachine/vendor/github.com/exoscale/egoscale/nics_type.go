package egoscale

import (
	"net"
)

// Nic represents a Network Interface Controller (NIC)
//
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/networking_and_traffic.html#configuring-multiple-ip-addresses-on-a-single-nic
type Nic struct {
	ID               string           `json:"id,omitempty"`
	BroadcastURI     string           `json:"broadcasturi,omitempty"`
	Gateway          net.IP           `json:"gateway,omitempty"`
	IP6Address       net.IP           `json:"ip6address,omitempty"`
	IP6Cidr          string           `json:"ip6cidr,omitempty"`
	IP6Gateway       net.IP           `json:"ip6gateway,omitempty"`
	IPAddress        net.IP           `json:"ipaddress,omitempty"`
	IsDefault        bool             `json:"isdefault,omitempty"`
	IsolationURI     string           `json:"isolationuri,omitempty"`
	MacAddress       string           `json:"macaddress,omitempty"`
	Netmask          net.IP           `json:"netmask,omitempty"`
	NetworkID        string           `json:"networkid,omitempty"`
	NetworkName      string           `json:"networkname,omitempty"`
	SecondaryIP      []NicSecondaryIP `json:"secondaryip,omitempty"`
	TrafficType      string           `json:"traffictype,omitempty"`
	Type             string           `json:"type,omitempty"`
	VirtualMachineID string           `json:"virtualmachineid,omitempty"`
}

// NicSecondaryIP represents a link between NicID and IPAddress
type NicSecondaryIP struct {
	ID               string `json:"id"`
	IPAddress        net.IP `json:"ipaddress"`
	NetworkID        string `json:"networkid,omitempty"`
	NicID            string `json:"nicid,omitempty"`
	VirtualMachineID string `json:"virtualmachineid,omitempty"`
}

// ListNics represents the NIC search
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listNics.html
type ListNics struct {
	VirtualMachineID string `json:"virtualmachineid"`
	ForDisplay       bool   `json:"fordisplay,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	NetworkID        string `json:"networkid,omitempty"`
	NicID            string `json:"nicid,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
}

// ListNicsResponse represents a list of templates
type ListNicsResponse struct {
	Count int   `json:"count"`
	Nic   []Nic `json:"nic"`
}

// AddIPToNic (Async) represents the assignation of a secondary IP
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/addIpToNic.html
type AddIPToNic struct {
	NicID     string `json:"nicid"`
	IPAddress net.IP `json:"ipaddress"`
}

// AddIPToNicResponse represents the addition of an IP to a NIC
type AddIPToNicResponse struct {
	NicSecondaryIP NicSecondaryIP `json:"nicsecondaryip"`
}

// RemoveIPFromNic (Async) represents a deletion request
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/removeIpFromNic.html
type RemoveIPFromNic struct {
	ID string `json:"id"`
}

// ActivateIP6 (Async) activates the IP6 on the given NIC
//
// Exoscale specific API: https://community.exoscale.ch/api/compute/#activateip6_GET
type ActivateIP6 struct {
	NicID string `json:"nicid"`
}

// ActivateIP6Response represents the modified NIC
type ActivateIP6Response struct {
	Nic Nic `json:"nic"`
}
