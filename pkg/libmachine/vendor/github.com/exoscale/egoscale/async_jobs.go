package egoscale

import "encoding/json"

// QueryAsyncJobResult represents a query to fetch the status of async job
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/queryAsyncJobResult.html
type QueryAsyncJobResult struct {
	JobID string `json:"jobid" doc:"the ID of the asychronous job"`
}

// name returns the CloudStack API command name
func (*QueryAsyncJobResult) name() string {
	return "queryAsyncJobResult"
}

func (*QueryAsyncJobResult) response() interface{} {
	return new(QueryAsyncJobResultResponse)
}

// name returns the CloudStack API command name
func (*ListAsyncJobs) name() string {
	return "listAsyncJobs"
}

func (*ListAsyncJobs) response() interface{} {
	return new(ListAsyncJobsResponse)
}

//Response return response of AsyncJobResult from a given type
func (a *AsyncJobResult) Response(i interface{}) error {
	if a.JobStatus == Failure {
		return a.Error()
	}
	if a.JobStatus == Success {
		if err := json.Unmarshal(*(a.JobResult), i); err != nil {
			return err
		}
	}
	return nil
}

func (a *AsyncJobResult) Error() error {
	r := new(ErrorResponse)
	if e := json.Unmarshal(*a.JobResult, r); e != nil {
		return e
	}
	return r
}
