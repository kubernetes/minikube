package egoscale

import (
	"encoding/json"
)

// AsyncJobResult represents an asynchronous job result
type AsyncJobResult struct {
	AccountID       string           `json:"accountid"`
	Cmd             string           `json:"cmd"`
	Created         string           `json:"created"`
	JobInstanceID   string           `json:"jobinstanceid,omitempty"`
	JobInstanceType string           `json:"jobinstancetype,omitempty"`
	JobProcStatus   int              `json:"jobprocstatus"`
	JobResult       *json.RawMessage `json:"jobresult"`
	JobResultCode   int              `json:"jobresultcode"`
	JobResultType   string           `json:"jobresulttype"`
	JobStatus       JobStatusType    `json:"jobstatus"`
	UserID          string           `json:"userid"`
	JobID           string           `json:"jobid"`
}

// QueryAsyncJobResultResponse represents the current status of an asynchronous job
type QueryAsyncJobResultResponse AsyncJobResult

// ListAsyncJobs list the asynchronous jobs
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listAsyncJobs.html
type ListAsyncJobs struct {
	Account     string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	DomainID    string `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	IsRecursive *bool  `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword     string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll     *bool  `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	StartDate   string `json:"startdate,omitempty" doc:"the start date of the async job"`
}

// ListAsyncJobsResponse represents a list of job results
type ListAsyncJobsResponse struct {
	Count     int              `json:"count"`
	AsyncJobs []AsyncJobResult `json:"asyncjobs"`
}
