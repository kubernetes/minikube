package egoscale

// Volume represents a volume linked to a VM
type Volume struct {
	ID                         string        `json:"id"`
	Account                    string        `json:"account,omitempty"`
	Attached                   string        `json:"attached,omitempty"`
	ChainInfo                  string        `json:"chaininfo,omitempty"`
	Created                    string        `json:"created,omitempty"`
	Destroyed                  bool          `json:"destroyed,omitempty"`
	DisplayVolume              bool          `json:"displayvolume,omitempty"`
	Domain                     string        `json:"domain,omitempty"`
	DomainID                   string        `json:"domainid,omitempty"`
	Name                       string        `json:"name,omitempty"`
	QuiesceVM                  bool          `json:"quiescevm,omitempty"`
	ServiceOfferingDisplayText string        `json:"serviceofferingdisplaytext,omitempty"`
	ServiceOfferingID          string        `json:"serviceofferingid,omitempty"`
	ServiceOfferingName        string        `json:"serviceofferingname,omitempty"`
	Size                       uint64        `json:"size,omitempty"`
	State                      string        `json:"state,omitempty"`
	Type                       string        `json:"type,omitempty"`
	VirtualMachineID           string        `json:"virtualmachineid,omitempty"`
	VMName                     string        `json:"vmname,omitempty"`
	VMState                    string        `json:"vmstate,omitempty"`
	ZoneID                     string        `json:"zoneid,omitempty"`
	ZoneName                   string        `json:"zonename,omitempty"`
	Tags                       []ResourceTag `json:"tags,omitempty"`
	JobID                      string        `json:"jobid,omitempty"`
	JobStatus                  JobStatusType `json:"jobstatus,omitempty"`
}

// ResizeVolume (Async) resizes a volume
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/resizeVolume.html
type ResizeVolume struct {
	ID             string `json:"id"`
	DiskOfferingID string `json:"diskofferingid,omitempty"`
	ShrinkOk       *bool  `json:"shrinkok,omitempty"`
	Size           int64  `json:"size,omitempty"` // in GiB
}

// ListVolumes represents a query listing volumes
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listVolumes.html
type ListVolumes struct {
	Account          string        `json:"account,omitempty"`
	DiskOfferingID   string        `json:"diskoffering,omitempty"`
	DisplayVolume    string        `json:"displayvolume,omitempty"` // root only
	DomainID         string        `json:"domainid,omitempty"`
	HostID           string        `json:"hostid,omitempty"`
	ID               string        `json:"id,omitempty"`
	IsRecursive      *bool         `json:"isrecursive,omitempty"`
	Keyword          string        `json:"keyword,omitempty"`
	ListAll          *bool         `json:"listall,omitempty"`
	Name             string        `json:"name,omitempty"`
	Page             int           `json:"page,omitempty"`
	PageSize         int           `json:"pagesize,omitempty"`
	PodID            string        `json:"podid,omitempty"`
	ProjectID        string        `json:"projectid,omitempty"`
	StorageID        string        `json:"storageid,omitempty"`
	Tags             []ResourceTag `json:"tags,omitempty"`
	Type             string        `json:"type,omitempty"`
	VirtualMachineID string        `json:"virtualmachineid,omitempty"`
	ZoneID           string        `json:"zoneid,omitempty"`
}

// ListVolumesResponse represents a list of volumes
type ListVolumesResponse struct {
	Count  int      `json:"count"`
	Volume []Volume `json:"volume"`
}
