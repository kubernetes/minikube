// Copyright 2015 opentsdb-goclient authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//
// Package client defines the client and the corresponding
// rest api implementaion of OpenTSDB.
//
// client.go contains the global interface and implementation struct
// definition of the OpenTSDB Client, as well as the common private
// and public methods used by all the rest-api implementation files,
// whose names are just like put.go, query.go, and so on.
//
package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bluebreezecf/opentsdb-goclient/config"
)

const (
	DefaultDialTimeout = 5 * time.Second
	KeepAliveTimeout   = 30 * time.Second
	GetMethod          = "GET"
	PostMethod         = "POST"
	PutMethod          = "PUT"
	DeleteMethod       = "DELETE"

	PutPath            = "/api/put"
	PutRespWithSummary = "summary"
	PutRespWithDetails = "details"

	QueryPath = "/api/query"
	// The three keys in the rateOption parameter of the QueryParam
	QueryRateOptionCounter    = "counter"    // The corresponding value type is bool
	QueryRateOptionCounterMax = "counterMax" // The corresponding value type is int,int64
	QueryRateOptionResetValue = "resetValue" // The corresponding value type is int,int64

	AggregatorPath  = "/api/aggregators"
	ConfigPath      = "/api/config"
	SerializersPath = "/api/serializers"
	StatsPath       = "/api/stats"
	SuggestPath     = "/api/suggest"
	// Only the one of the three query type can be used in SuggestParam, UIDMetaData:
	TypeMetrics = "metrics"
	TypeTagk    = "tagk"
	TypeTagv    = "tagv"

	VersionPath        = "/api/version"
	DropcachesPath     = "/api/dropcaches"
	AnnotationPath     = "/api/annotation"
	AnQueryStartTime   = "start_time"
	AnQueryTSUid       = "tsuid"
	BulkAnnotationPath = "/api/annotation/bulk"
	UIDMetaDataPath    = "/api/uid/uidmeta"
	UIDAssignPath      = "/api/uid/assign"
	TSMetaDataPath     = "/api/uid/tsmeta"

	// The above three constants are used in /put
	DefaultMaxPutPointsNum = 75
	DefaultDetectDeltaNum  = 3
	// Unit is bytes, and assumes that config items of 'tsd.http.request.enable_chunked = true'
	// and 'tsd.http.request.max_chunk = 40960' are all in the opentsdb.conf:
	DefaultMaxContentLength = 40960
)

var (
	DefaultTransport = &http.Transport{
		MaxIdleConnsPerHost: 10,
		Dial: (&net.Dialer{
			Timeout:   DefaultDialTimeout,
			KeepAlive: KeepAliveTimeout,
		}).Dial,
	}
)

// Client defines the sdk methods, by which other go applications can
// commnicate with the OpenTSDB via the pre-defined rest-apis.
// Each method defined in the interface of Client is in the correspondance
// a rest-api definition in (http://opentsdb.net/docs/build/html/api_http/index.html#api-endpoints).
type Client interface {

	// Ping detects whether the target OpenTSDB is reachable or not.
	// If error occurs during the detection, an error instance will be returned, or nil otherwise.
	Ping() error

	// Put is the implementation of 'POST /api/put' endpoint.
	// This endpoint allows for storing data in OpenTSDB over HTTP as an alternative to the Telnet interface.
	//
	// datas is a slice of DataPoint holding at least one instance.
	// queryParam can only be github.com/bluebreezecf/opentsdb-goclient/client.PutRespWithSummary,
	// github.com/bluebreezecf/opentsdb-goclient/client.PutRespWithDetails or the empty string "";
	// It means get put summary response info by using PutRespWithSummary, and
	// with PutRespWithDetails means get put detailed response.
	//
	// When put operation is successful, a pointer of PutResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, when the given parameters
	// are invalid, it failed to parese the response, or OpenTSDB is un-connectable right now.
	Put(datas []DataPoint, queryParam string) (*PutResponse, error)

	// Query is the implementation of 'GET /api/query' endpoint.
	// It is probably the most useful endpoint in the API, /api/query enables extracting data from the storage
	// system in various formats determined by the serializer selected.
	//
	// param is a instance of QueryParam holding current query parameters.
	//
	// When query operation is successful, a pointer of QueryResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, when the given parameter
	// is invalid, it failed to parese the response, or OpenTSDB is un-connectable right now.
	Query(param QueryParam) (*QueryResponse, error)

	// Aggregators is the implementation of 'GET /api/aggregators' endpoint.
	// It simply lists the names of implemented aggregation functions used in timeseries queries.
	//
	// When query operation is successful, a pointer of AggregatorsResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, when it failed to parese the
	// response, or OpenTSDB is un-connectable right now.
	Aggregators() (*AggregatorsResponse, error)

	// Config is the implementation of 'GET /api/config' endpoint.
	// It returns information about the running configuration of the TSD.
	// It is read only and cannot be used to set configuration options.
	//
	// When query operation is successful, a pointer of ConfigResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, when it failed to parese the
	// response, or OpenTSDB is un-connectable right now.
	Config() (*ConfigResponse, error)

	// Serializers is the implementation of 'GET /api/serializers' endpoint.
	// It lists the serializer plugins loaded by the running TSD. Information given includes the name,
	// implemented methods, content types and methods.
	//
	// When query operation is successful, a pointer of SerialResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, when it failed to parese the
	// response, or OpenTSDB is un-connectable right now.
	Serializers() (*SerialResponse, error)

	// Stats is the implementation of 'GET /api/stats' endpoint.
	// It provides a list of statistics for the running TSD. These statistics are automatically recorded
	// by a running TSD every 5 minutes but may be accessed via this endpoint. All statistics are read only.
	//
	// When query operation is successful, a pointer of StatsResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, when it failed to parese the
	// response, or OpenTSDB is un-connectable right now.
	Stats() (*StatsResponse, error)

	// Suggest is the implementation of 'GET /api/suggest' endpoint.
	// It provides a means of implementing an "auto-complete" call that can be accessed repeatedly as a user
	// types a request in a GUI. It does not offer full text searching or wildcards, rather it simply matches
	// the entire string passed in the query on the first characters of the stored data.
	// For example, passing a query of type=metrics&q=sys will return the top 25 metrics in the system that start with sys.
	// Matching is case sensitive, so sys will not match System.CPU. Results are sorted alphabetically.
	//
	// sugParm is an instance of SuggestParam storing parameters by invoking /api/suggest.
	//
	// When query operation is successful, a pointer of SuggestResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	Suggest(sugParm SuggestParam) (*SuggestResponse, error)

	// Version is the implementation of 'GET /api/version' endpoint.
	// It returns information about the running version of OpenTSDB.
	//
	// When query operation is successful, a pointer of VersionResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, when it failed to parese the
	// response, or OpenTSDB is un-connectable right now.
	Version() (*VersionResponse, error)

	// Dropcaches is the implementation of 'GET /api/dropcaches' endpoint.
	// It purges the in-memory data cached in OpenTSDB. This includes all UID to name
	// and name to UID maps for metrics, tag names and tag values.
	//
	// When query operation is successful, a pointer of DropcachesResponse will be returned with the corresponding
	// status code and response info. Otherwise, an error instance will be returned, when it failed to parese the
	// response, or OpenTSDB is un-connectable right now.
	Dropcaches() (*DropcachesResponse, error)

	// QueryAnnotation is the implementation of 'GET /api/annotation' endpoint.
	// It retrieves a single annotation stored in the OpenTSDB backend.
	//
	// queryAnnoParam is a map storing parameters of a target queried annotation.
	// The key can be such as client.AnQueryStartTime, client.AnQueryTSUid.
	//
	// When query operation is handlering properly by the OpenTSDB backend, a pointer of AnnotationResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	QueryAnnotation(queryAnnoParam map[string]interface{}) (*AnnotationResponse, error)

	// UpdateAnnotation is the implementation of 'POST /api/annotation' endpoint.
	// It creates or modifies an annotation stored in the OpenTSDB backend.
	//
	// annotation is an annotation to be processed in the OpenTSDB backend.
	//
	// When modification operation is handlering properly by the OpenTSDB backend, a pointer of AnnotationResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	UpdateAnnotation(annotation Annotation) (*AnnotationResponse, error)

	// DeleteAnnotation is the implementation of 'DELETE /api/annotation' endpoint.
	// It deletes an annotation stored in the OpenTSDB backend.
	//
	// annotation is an annotation to be deleted in the OpenTSDB backend.
	//
	// When deleting operation is handlering properly by the OpenTSDB backend, a pointer of AnnotationResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	DeleteAnnotation(annotation Annotation) (*AnnotationResponse, error)

	// BulkUpdateAnnotations is the implementation of 'POST /api/annotation/bulk' endpoint.
	// It creates or modifies a list of annotation stored in the OpenTSDB backend.
	//
	// annotations is a list of annotations to be processed (to be created or modified) in the OpenTSDB backend.
	//
	// When bulk modification operation is handlering properly by the OpenTSDB backend, a pointer of BulkAnnotatResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	BulkUpdateAnnotations(annotations []Annotation) (*BulkAnnotatResponse, error)

	// BulkDeleteAnnotations is the implementation of 'DELETE /api/annotation/bulk' endpoint.
	// It deletes a list of annotation stored in the OpenTSDB backend.
	//
	// bulkDelParam contains the bulk deleting info in current invoking 'DELETE /api/annotation/bulk'.
	//
	// When bulk deleting operation is handlering properly by the OpenTSDB backend, a pointer of BulkAnnotatResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	BulkDeleteAnnotations(bulkDelParam BulkAnnoDeleteInfo) (*BulkAnnotatResponse, error)

	// QueryUIDMetaData is the implementation of 'GET /api/uid/uidmeta' endpoint.
	// It retrieves a single UIDMetaData stored in the OpenTSDB backend with the given query parameters.
	//
	// metaQueryParam is a map storing parameters of a target queried UIDMetaData.
	// It must contain two key/value pairs with the key "uid" and "type".
	// "type" should be one of client.TypeMetrics ("metric"), client.TypeTagk ("tagk"), and client.TypeTagv ("tagv")
	//
	// When query operation is handlering properly by the OpenTSDB backend, a pointer of UIDMetaDataResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	QueryUIDMetaData(metaQueryParam map[string]string) (*UIDMetaDataResponse, error)

	// UpdateUIDMetaData is the implementation of 'POST /api/uid/uidmeta' endpoint.
	// It modifies a UIDMetaData.
	//
	// uidMetaData is an instance of UIDMetaData to be modified
	//
	// When update operation is handlering properly by the OpenTSDB backend, a pointer of UIDMetaDataResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	UpdateUIDMetaData(uidMetaData UIDMetaData) (*UIDMetaDataResponse, error)

	// DeleteUIDMetaData is the implementation of 'DELETE /api/uid/uidmeta' endpoint.
	// It deletes a target UIDMetaData.
	//
	// uidMetaData is an instance of UIDMetaData whose correspance is to be deleted.
	// The values of uid and type in uidMetaData is required.
	//
	// When delete operation is handlering properly by the OpenTSDB backend, a pointer of UIDMetaDataResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	DeleteUIDMetaData(uidMetaData UIDMetaData) (*UIDMetaDataResponse, error)

	// AssignUID is the implementation of 'POST /api/uid/assigin' endpoint.
	// It enables assigning UIDs to new metrics, tag names and tag values. Multiple types and names can be provided
	// in a single call and the API will process each name individually, reporting which names were assigned UIDs
	// successfully, along with the UID assigned, and which failed due to invalid characters or had already been assigned.
	// Assignment can be performed via query string or content data.
	//
	// assignParam is an instance of UIDAssignParam holding the parameters to invoke 'POST /api/uid/assigin'.
	//
	// When assigin operation is handlering properly by the OpenTSDB backend, a pointer of UIDAssignResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	AssignUID(assignParam UIDAssignParam) (*UIDAssignResponse, error)

	// QueryTSMetaData is the implementation of 'GET /api/uid/tsmeta' endpoint.
	// It retrieves a single TSMetaData stored in the OpenTSDB backend with the given query parameters.
	//
	// tsuid is a tsuid of a target queried TSMetaData.
	//
	// When query operation is handlering properly by the OpenTSDB backend, a pointer of TSMetaDataResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, if the given parameter is invalid,
	// or when it failed to parese the response, or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	QueryTSMetaData(tsuid string) (*TSMetaDataResponse, error)

	// UpdateTSMetaData is the implementation of 'POST /api/uid/tsmeta' endpoint.
	// It modifies a target TSMetaData with the given fields.
	//
	// tsMetaData is an instance of UIDMetaData whose correspance is to be modified
	//
	// When update operation is handlering properly by the OpenTSDB backend, a pointer of TSMetaDataResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, when it failed to parese the response,
	// or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	UpdateTSMetaData(tsMetaData TSMetaData) (*TSMetaDataResponse, error)

	// DeleteTSMetaData is the implementation of 'DELETE /api/uid/tsmeta' endpoint.
	// It deletes a target TSMetaData.
	//
	// tsMetaData is an instance of UIDMetaData whose correspance is to be deleted
	//
	// When delete operation is handlering properly by the OpenTSDB backend, a pointer of TSMetaDataResponse
	// will be returned with the corresponding status code and response info (including the potential error
	// messages replied by OpenTSDB).
	//
	// Otherwise, an error instance will be returned, when it failed to parese the response,
	// or OpenTSDB is un-connectable right now.
	//
	// Note that: the returned non-nil error instance is only responsed by opentsdb-client, not the OpenTSDB backend.
	DeleteTSMetaData(tsMetaData TSMetaData) (*TSMetaDataResponse, error)
}

// NewClient creates an instance of http client which implements the
// pre-defined rest apis of OpenTSDB.
// A non-nil error instance returned means currently the target OpenTSDB
// designated with the given endpoint is not connectable.
func NewClient(opentsdbCfg config.OpenTSDBConfig) (Client, error) {
	opentsdbCfg.OpentsdbHost = strings.TrimSpace(opentsdbCfg.OpentsdbHost)
	if len(opentsdbCfg.OpentsdbHost) <= 0 {
		return nil, errors.New("The OpentsdbEndpoint of the given config should not be empty.")
	}
	transport := opentsdbCfg.Transport
	if transport == nil {
		transport = DefaultTransport
	}
	client := &http.Client{
		Transport: transport,
	}
	if opentsdbCfg.MaxPutPointsNum <= 0 {
		opentsdbCfg.MaxPutPointsNum = DefaultMaxPutPointsNum
	}
	if opentsdbCfg.DetectDeltaNum <= 0 {
		opentsdbCfg.DetectDeltaNum = DefaultDetectDeltaNum
	}
	if opentsdbCfg.MaxContentLength <= 0 {
		opentsdbCfg.MaxContentLength = DefaultMaxContentLength
	}
	tsdbEndpoint := fmt.Sprintf("http://%s", opentsdbCfg.OpentsdbHost)
	clientImpl := clientImpl{
		tsdbEndpoint: tsdbEndpoint,
		client:       client,
		opentsdbCfg:  opentsdbCfg,
	}
	return &clientImpl, nil
}

// The private implementation of Client interface.
type clientImpl struct {
	tsdbEndpoint string
	client       *http.Client
	opentsdbCfg  config.OpenTSDBConfig
}

// Response defines the common behaviours all the specific response for
// different rest-apis shound obey.
// Currently it is an abstraction used in (*clientImpl).sendRequest()
// to stored the different kinds of response contents for all the rest-apis.
type Response interface {

	// SetStatus can be used to set the actual http status code of
	// the related http response for the specific Response instance
	SetStatus(code int)

	// GetCustomParser can be used to retrive a custom-defined parser.
	// Returning nil means current specific Response instance doesn't
	// need a custom-defined parse process, and just uses the default
	// json unmarshal method to parse the contents of the http response.
	GetCustomParser() func(respCnt []byte) error

	// Return the contents of the specific Response instance with
	// the string format
	String() string
}

// sendRequest dispatches the http request with the given method name, url and body contents.
// reqBodyCnt is "" means there is no contents in the request body.
// If the tsdb server responses properly, the error is nil and parsedResp is the parsed
// response with the specific type. Otherwise, the returned error is not nil.
func (c *clientImpl) sendRequest(method, url, reqBodyCnt string, parsedResp Response) error {
	req, err := http.NewRequest(method, url, strings.NewReader(reqBodyCnt))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create request for %s %s: %v", method, url, err))
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := c.client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to send request for %s %s: %v", method, url, err))
	}
	defer resp.Body.Close()
	var jsonBytes []byte
	if jsonBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		return errors.New(fmt.Sprintf("Failed to read response for %s %s: %v", method, url, err))
	}

	parsedResp.SetStatus(resp.StatusCode)
	parser := parsedResp.GetCustomParser()
	if parser == nil {
		if err = json.Unmarshal(jsonBytes, parsedResp); err != nil {
			return errors.New(fmt.Sprintf("Failed to parse response for %s %s: %v", method, url, err))
		}
	} else {
		if err = parser(jsonBytes); err != nil {
			return err
		}
	}

	return nil
}

func (c *clientImpl) isValidOperateMethod(method string) bool {
	method = strings.TrimSpace(strings.ToUpper(method))
	if len(method) == 0 {
		return false
	}
	methods := []string{PostMethod, PutMethod, DeleteMethod}
	exists := false
	for _, item := range methods {
		if method == item {
			exists = true
			break
		}
	}
	return exists
}

func (c *clientImpl) Ping() error {
	conn, err := net.DialTimeout("tcp", c.opentsdbCfg.OpentsdbHost, DefaultDialTimeout)
	if err != nil {
		return errors.New(fmt.Sprintf("The target OpenTSDB is unreachable: %v", err))
	}
	if conn != nil {
		defer conn.Close()
	}
	return nil
}
