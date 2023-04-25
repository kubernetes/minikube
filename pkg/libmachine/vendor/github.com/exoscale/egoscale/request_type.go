package egoscale

import (
	"encoding/json"
	"net/url"
)

// Command represents a CloudStack request
type Command interface {
	// CloudStack API command name
	name() string
}

// SyncCommand represents a CloudStack synchronous request
type syncCommand interface {
	Command
	// Response interface to Unmarshal the JSON into
	response() interface{}
}

// AsyncCommand represents a async CloudStack request
type AsyncCommand interface {
	Command
	// Response interface to Unmarshal the JSON into
	asyncResponse() interface{}
}

// ListCommand represents a CloudStack list request
type ListCommand interface {
	Command
	// SetPage defines the current pages
	SetPage(int)
	// SetPageSize defines the size of the page
	SetPageSize(int)
	// each reads the data from the response and feeds channels, and returns true if we are on the last page
	each(interface{}, IterateItemFunc)
}

// onBeforeHook represents an action to be done on the params before sending them
//
// This little took helps with issue of relying on JSON serialization logic only.
// `omitempty` may make sense in some cases but not all the time.
type onBeforeHook interface {
	onBeforeSend(params *url.Values) error
}

//go:generate stringer -type JobStatusType
const (
	// Pending represents a job in progress
	Pending JobStatusType = iota
	// Success represents a successfully completed job
	Success
	// Failure represents a job that has failed to complete
	Failure
)

// JobStatusType represents the status of a Job
type JobStatusType int

//go:generate stringer -type ErrorCode
const (
	// Unauthorized represents ... (TODO)
	Unauthorized ErrorCode = 401
	// MethodNotAllowed represents ... (TODO)
	MethodNotAllowed ErrorCode = 405
	// UnsupportedActionError represents ... (TODO)
	UnsupportedActionError ErrorCode = 422
	// APILimitExceeded represents ... (TODO)
	APILimitExceeded ErrorCode = 429
	// MalformedParameterError represents ... (TODO)
	MalformedParameterError ErrorCode = 430
	// ParamError represents ... (TODO)
	ParamError ErrorCode = 431

	// InternalError represents a server error
	InternalError ErrorCode = 530
	// AccountError represents ... (TODO)
	AccountError ErrorCode = 531
	// AccountResourceLimitError represents ... (TODO)
	AccountResourceLimitError ErrorCode = 532
	// InsufficientCapacityError represents ... (TODO)
	InsufficientCapacityError ErrorCode = 533
	// ResourceUnavailableError represents ... (TODO)
	ResourceUnavailableError ErrorCode = 534
	// ResourceAllocationError represents ... (TODO)
	ResourceAllocationError ErrorCode = 535
	// ResourceInUseError represents ... (TODO)
	ResourceInUseError ErrorCode = 536
	// NetworkRuleConflictError represents ... (TODO)
	NetworkRuleConflictError ErrorCode = 537
)

// ErrorCode represents the CloudStack ApiErrorCode enum
//
// See: https://github.com/apache/cloudstack/blob/master/api/src/org/apache/cloudstack/api/ApiErrorCode.java
type ErrorCode int

// JobResultResponse represents a generic response to a job task
type JobResultResponse struct {
	AccountID     string           `json:"accountid,omitempty"`
	Cmd           string           `json:"cmd"`
	Created       string           `json:"created"`
	JobID         string           `json:"jobid"`
	JobProcStatus int              `json:"jobprocstatus"`
	JobResult     *json.RawMessage `json:"jobresult"`
	JobStatus     JobStatusType    `json:"jobstatus"`
	JobResultType string           `json:"jobresulttype"`
	UserID        string           `json:"userid,omitempty"`
}

// ErrorResponse represents the standard error response from CloudStack
type ErrorResponse struct {
	ErrorCode   ErrorCode  `json:"errorcode"`
	CsErrorCode int        `json:"cserrorcode"`
	ErrorText   string     `json:"errortext"`
	UUIDList    []UUIDItem `json:"uuidList,omitempty"` // uuid*L*ist is not a typo
}

// UUIDItem represents an item of the UUIDList part of an ErrorResponse
type UUIDItem struct {
	Description      string `json:"description,omitempty"`
	SerialVersionUID int64  `json:"serialVersionUID,omitempty"`
	UUID             string `json:"uuid"`
}

// booleanResponse represents a boolean response (usually after a deletion)
type booleanResponse struct {
	Success     json.RawMessage `json:"success"`
	DisplayText string          `json:"displaytext,omitempty"`
}
