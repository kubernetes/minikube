package egoscale

// Snapshot represents a volume snapshot
type Snapshot struct {
	ID           string        `json:"id"`
	Account      string        `json:"account"`
	Created      string        `json:"created,omitempty"`
	Domain       string        `json:"domain"`
	DomainID     string        `json:"domainid"`
	IntervalType string        `json:"intervaltype,omitempty"` // hourly, daily, weekly, monthly, ..., none
	Name         string        `json:"name,omitempty"`
	PhysicalSize int64         `json:"physicalsize"`
	Project      string        `json:"project"`
	ProjectID    string        `json:"projectid"`
	Revertable   bool          `json:"revertable,omitempty"`
	Size         int64         `json:"size,omitempty"`
	SnapshotType string        `json:"snapshottype,omitempty"`
	State        string        `json:"state"` // BackedUp, Creating, BackingUp, ...
	VolumeID     string        `json:"volumeid"`
	VolumeName   string        `json:"volumename,omitempty"`
	VolumeType   string        `json:"volumetype,omitempty"`
	ZoneID       string        `json:"zoneid"`
	Tags         []ResourceTag `json:"tags"`
	JobID        string        `json:"jobid,omitempty"`
	JobStatus    JobStatusType `json:"jobstatus,omitempty"`
}

// CreateSnapshot represents a request to create a volume snapshot
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/createSnapshot.html
type CreateSnapshot struct {
	VolumeID  string `json:"volumeid" doc:"The ID of the disk volume"`
	Account   string `json:"account,omitempty" doc:"The account of the snapshot. The account parameter must be used with the domainId parameter."`
	Name      string `json:"name,omitempty" doc:"the name of the snapshot"`
	DomainID  string `json:"domainid,omitempty" doc:"The domain ID of the snapshot. If used with the account parameter, specifies a domain for the account associated with the disk volume."`
	PolicyID  string `json:"policyid,omitempty" doc:"policy id of the snapshot, if this is null, then use MANUAL_POLICY."`
	QuiesceVM *bool  `json:"quiescevm,omitempty" doc:"quiesce vm if true"`
}

// ListSnapshots lists the volume snapshots
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/listSnapshots.html
type ListSnapshots struct {
	Account      string        `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID     string        `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID           string        `json:"id,omitempty" doc:"lists snapshot by snapshot ID"`
	IDs          []string      `json:"ids,omitempty" doc:"the IDs of the snapshots, mutually exclusive with id"`
	IntervalType string        `json:"intervaltype,omitempty" doc:"valid values are HOURLY, DAILY, WEEKLY, and MONTHLY."`
	IsRecursive  *bool         `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword      string        `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll      *bool         `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Name         string        `json:"name,omitempty" doc:"lists snapshot by snapshot name"`
	Page         int           `json:"page,omitempty"`
	PageSize     int           `json:"pagesize,omitempty"`
	ProjectID    string        `json:"projectid,omitempty" doc:"list objects by project"`
	SnapshotType string        `json:"snapshottype,omitempty" doc:"valid values are MANUAL or RECURRING."`
	Tags         []ResourceTag `json:"tags,omitempty" doc:"List resources by tags (key/value pairs)"`
	VolumeID     string        `json:"volumeid,omitempty" doc:"the ID of the disk volume"`
	ZoneID       string        `json:"zoneid,omitempty" doc:"list snapshots by zone id"`
}

// DeleteSnapshot represents the deletion of a volume snapshot
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSnapshot.html
type DeleteSnapshot struct {
	ID string `json:"id" doc:"The ID of the snapshot"`
}

// RevertSnapshot revert a volume snapshot
//
// CloudStackAPI: http://cloudstack.apache.org/api/apidocs-4.10/apis/revertSnapshot.html
type RevertSnapshot struct {
	ID string `json:"id" doc:"The ID of the snapshot"`
}
