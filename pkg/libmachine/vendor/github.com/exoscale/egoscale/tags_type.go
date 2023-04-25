package egoscale

// ResourceTag is a tag associated with a resource
//
// http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/4.9/management.html
type ResourceTag struct {
	Account      string `json:"account,omitempty" doc:"the account associated with the tag"`
	Customer     string `json:"customer,omitempty" doc:"customer associated with the tag"`
	Domain       string `json:"domain,omitempty" doc:"the domain associated with the tag"`
	DomainID     string `json:"domainid,omitempty" doc:"the ID of the domain associated with the tag"`
	Key          string `json:"key,omitempty" doc:"tag key name"`
	Project      string `json:"project,omitempty" doc:"the project name where tag belongs to"`
	ProjectID    string `json:"projectid,omitempty" doc:"the project id the tag belongs to"`
	ResourceID   string `json:"resourceid,omitempty" doc:"id of the resource"`
	ResourceType string `json:"resourcetype,omitempty" doc:"resource type"`
	Value        string `json:"value,omitempty" doc:"tag value"`
}

// CreateTags (Async) creates resource tag(s)
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createTags.html
type CreateTags struct {
	ResourceIDs  []string      `json:"resourceids" doc:"list of resources to create the tags for"`
	ResourceType string        `json:"resourcetype" doc:"type of the resource"`
	Tags         []ResourceTag `json:"tags" doc:"Map of tags (key/value pairs)"`
	Customer     string        `json:"customer,omitempty" doc:"identifies client specific tag. When the value is not null, the tag can't be used by cloudStack code internally"`
}

// DeleteTags (Async) deletes the resource tag(s)
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteTags.html
type DeleteTags struct {
	ResourceIDs  []string      `json:"resourceids" doc:"Delete tags for resource id(s)"`
	ResourceType string        `json:"resourcetype" doc:"Delete tag by resource type"`
	Tags         []ResourceTag `json:"tags,omitempty" doc:"Delete tags matching key/value pairs"`
}

// ListTags list resource tag(s)
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listTags.html
type ListTags struct {
	Account      string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	Customer     string `json:"customer,omitempty" doc:"list by customer name"`
	DomainID     string `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	IsRecursive  *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Key          string `json:"key,omitempty" doc:"list by key"`
	Keyword      string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll      *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page         int    `json:"page,omitempty"`
	PageSize     int    `json:"pagesize,omitempty"`
	ProjectID    string `json:"projectid,omitempty" doc:"list objects by project"`
	ResourceID   string `json:"resourceid,omitempty" doc:"list by resource id"`
	ResourceType string `json:"resourcetype,omitempty" doc:"list by resource type"`
	Value        string `json:"value,omitempty" doc:"list by value"`
}

// ListTagsResponse represents a list of resource tags
type ListTagsResponse struct {
	Count int           `json:"count"`
	Tag   []ResourceTag `json:"tag"`
}
