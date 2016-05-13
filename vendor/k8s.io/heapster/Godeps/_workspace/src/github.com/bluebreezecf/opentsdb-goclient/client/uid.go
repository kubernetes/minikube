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
// Every metric, tag name and tag value is associated with a unique identifier (UID).
// Internally, the UID is a binary array assigned to a text value the first time it is
// encountered or via an explicit assignment request. This endpoint provides utilities
// for managing UIDs and their associated data. Please see the UID endpoint TOC below
// for information on what functions are implemented.
//
// UIDs exposed via the API are encoded as hexadecimal strings. The UID 42 would be expressed
// as 00002A given the default UID width of 3 bytes.
// You may also edit meta data associated with timeseries or individual UID objects via the UID endpoint.
//
// uid.go contains the structs and methods for the implementation of
// /api/uid/tsmeta, /api/uid/assign, /api/uid/uidmeta.
//
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// UIDMetaData is the structure used to hold
// the parameters when calling (POST,PUT) /api/uid/uidmeta.
// Each attributes in UIDMetaData matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/uid/uidmeta.html).
//
type UIDMetaData struct {
	// A required hexadecimal representation of the UID
	Uid string `json:"uid,omitempty"`

	// A required type of UID, must be metric, tagk or tagv
	Type string `json:"type,omitempty"`

	// An optional brief description of what the UID represents
	Description string `json:"description,omitempty"`

	// An optional short name that can be displayed in GUIs instead of the default name
	DisplayName string `json:"displayName,omitempty"`

	// An optional detailed notes about what the UID represents
	Notes string `json:"notes,omitempty"`

	// An optional key/value map to store custom fields and values
	Custom map[string]string `json:"custom,omitempty"`
}

// UIDMetaDataResponse acts as the implementation of Response in the /api/uid/uidmeta scene.
// It holds the status code and the response values defined in the
// (http://opentsdb.net/docs/build/html/api_http/uid/uidmeta.html).
//
type UIDMetaDataResponse struct {
	UIDMetaData

	StatusCode int

	// The name of the UID as given when the data point was stored or the UID assigned
	Name string `json:"name,omitempty"`

	// A Unix epoch timestamp in seconds when the UID was first created.
	// If the meta data was not stored when the UID was assigned, this value may be 0.
	Created int64 `json:"created,omitempty"`

	ErrorInfo map[string]interface{} `json:"error,omitempty"`
}

func (uidMetaDataResp *UIDMetaDataResponse) SetStatus(code int) {
	uidMetaDataResp.StatusCode = code
}

func (uidMetaDataResp *UIDMetaDataResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		var resultBytes []byte
		if uidMetaDataResp.StatusCode == 204 || // The OpenTSDB deletes a UIDMetaData successfully, or
			uidMetaDataResp.StatusCode == 304 { // no changes were present, and with no body content.
			return nil
		} else {
			resultBytes = respCnt
		}
		return json.Unmarshal(resultBytes, &uidMetaDataResp)
	}
}

func (uidMetaDataResp *UIDMetaDataResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(uidMetaDataResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) QueryUIDMetaData(metaQueryParam map[string]string) (*UIDMetaDataResponse, error) {
	if !isValidUIDMetaDataQueryParam(metaQueryParam) {
		return nil, errors.New("The given query uid metadata is invalid.")
	}
	queryParam := fmt.Sprintf("%s=%v&%s=%v", "uid", metaQueryParam["uid"], "type", metaQueryParam["type"])
	queryUIDMetaEndpoint := fmt.Sprintf("%s%s?%s", c.tsdbEndpoint, UIDMetaDataPath, queryParam)
	uidMetaDataResp := UIDMetaDataResponse{}
	if err := c.sendRequest(GetMethod, queryUIDMetaEndpoint, "", &uidMetaDataResp); err != nil {
		return nil, err
	}
	return &uidMetaDataResp, nil
}

func (c *clientImpl) UpdateUIDMetaData(uidMetaData UIDMetaData) (*UIDMetaDataResponse, error) {
	return c.operateUIDMetaData(PostMethod, &uidMetaData)
}

func (c *clientImpl) DeleteUIDMetaData(uidMetaData UIDMetaData) (*UIDMetaDataResponse, error) {
	return c.operateUIDMetaData(DeleteMethod, &uidMetaData)
}

func (c *clientImpl) operateUIDMetaData(method string, uidMetaData *UIDMetaData) (*UIDMetaDataResponse, error) {
	if !c.isValidOperateMethod(method) {
		return nil, errors.New("The given method for operating a uid metadata is invalid.")
	}
	uidMetaEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, UIDMetaDataPath)
	resultBytes, err := json.Marshal(uidMetaData)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to marshal uidMetaData: %v", err))
	}
	uidMetaDataResp := UIDMetaDataResponse{}
	if err = c.sendRequest(method, uidMetaEndpoint, string(resultBytes), &uidMetaDataResp); err != nil {
		return nil, err
	}
	return &uidMetaDataResp, nil
}

func isValidUIDMetaDataQueryParam(metaQueryParam map[string]string) bool {
	if metaQueryParam == nil || len(metaQueryParam) != 2 {
		return false
	}
	checkKeys := []string{"uid", "type"}
	for _, checkKey := range checkKeys {
		_, exists := metaQueryParam[checkKey]
		if !exists {
			return false
		}
	}
	typeValue := metaQueryParam["type"]
	typeCheckItems := []string{TypeMetrics, TypeTagk, TypeTagv}
	for _, checkItem := range typeCheckItems {
		if typeValue == checkItem {
			return true
		}
	}
	return false
}

// UIDAssignParam is the structure used to hold
// the parameters when calling POST /api/uid/assign.
// Each attributes in UIDAssignParam matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/uid/assign.html).
//
type UIDAssignParam struct {
	// An optional list of metric names for assignment
	Metric []string `json:"metric,omitempty"`

	// An optional list of tag names for assignment
	Tagk []string `json:"tagk,omitempty"`

	// An optional list of tag values for assignment
	Tagv []string `json:"tagv,omitempty"`
}

// UIDAssignResponse acts as the implementation of Response in the POST /api/uid/assign scene.
// It holds the status code and the response values defined in the
// (http://opentsdb.net/docs/build/html/api_http/uid/assign.html).
//
type UIDAssignResponse struct {
	StatusCode   int
	Metric       map[string]string `json:"metric"`
	MetricErrors map[string]string `json:"metric_errors,omitempty"`
	Tagk         map[string]string `json:"tagk"`
	TagkErrors   map[string]string `json:"tagk_errors,omitempty"`
	Tagv         map[string]string `json:"tagv"`
	TagvErrors   map[string]string `json:"tagv_errors,omitempty"`
}

func (uidAssignResp *UIDAssignResponse) SetStatus(code int) {
	uidAssignResp.StatusCode = code
}

func (uidAssignResp *UIDAssignResponse) GetCustomParser() func(respCnt []byte) error {
	return nil
}

func (uidAssignResp *UIDAssignResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(uidAssignResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) AssignUID(assignParam UIDAssignParam) (*UIDAssignResponse, error) {
	assignUIDEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, UIDAssignPath)
	resultBytes, err := json.Marshal(assignParam)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to marshal UIDAssignParam: %v", err))
	}
	uidAssignResp := UIDAssignResponse{}
	if err = c.sendRequest(PostMethod, assignUIDEndpoint, string(resultBytes), &uidAssignResp); err != nil {
		return nil, err
	}
	return &uidAssignResp, nil
}

// TSMetaData is the structure used to hold
// the parameters when calling (POST,PUT,DELETE) /api/uid/tsmeta.
// Each attributes in TSMetaData matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/uid/tsmeta.html).
//
type TSMetaData struct {
	// A required hexadecimal representation of the timeseries UID
	Tsuid string `json:"tsuid,omitempty"`

	// An optional brief description of what the UID represents
	Description string `json:"description,omitempty"`

	// An optional short name that can be displayed in GUIs instead of the default name
	DisplayName string `json:"displayName,omitempty"`

	// An optional detailed notes about what the UID represents
	Notes string `json:"notes,omitempty"`

	// An optional key/value map to store custom fields and values
	Custom map[string]string `json:"custom,omitempty"`

	// An optional value reflective of the data stored in the timeseries, may be used in GUIs or calculations
	Units string `json:"units,omitempty"`

	// The kind of data stored in the timeseries such as counter, gauge, absolute, etc.
	// These may be defined later but they should be similar to Data Source Types in an RRD.
	// Its value is optional
	DataType string `json:"dataType,omitempty"`

	// The number of days of data points to retain for the given timeseries. Not Implemented.
	// When set to 0, the default, data is retained indefinitely.
	// Its value is optional
	Retention int64 `json:"retention,omitempty"`

	// An optional maximum value for this timeseries that may be used in calculations such as
	// percent of maximum. If the default of NaN is present, the value is ignored.
	Max float64 `json:"max,omitempty"`

	// An optional minimum value for this timeseries that may be used in calculations such as
	// percent of minimum. If the default of NaN is present, the value is ignored.
	Min float64 `json:"min,omitempty"`
}

type TSMetaDataResponse struct {
	StatusCode int
	TSMetaData
	Metric          UIDMetaData            `json:"metric,omitempty"`
	Tags            []UIDMetaData          `json:"tags,omitempty"`
	Created         int64                  `json:"created,omitempty"`
	LastReceived    int64                  `json:"lastReceived,omitempty"`
	TotalDatapoints int64                  `json:"totalDatapoints,omitempty"`
	ErrorInfo       map[string]interface{} `json:"error,omitempty"`
}

func (tsMetaDataResp *TSMetaDataResponse) SetStatus(code int) {
	tsMetaDataResp.StatusCode = code
}

func (tsMetaDataResp *TSMetaDataResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		var resultBytes []byte
		if tsMetaDataResp.StatusCode == 204 || // The OpenTSDB deletes a TSMetaData successfully, or
			tsMetaDataResp.StatusCode == 304 { // no changes were present, and with no body content.
			return nil
		} else {
			resultBytes = respCnt
		}
		return json.Unmarshal(resultBytes, &tsMetaDataResp)
	}
}

func (tsMetaDataResp *TSMetaDataResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(tsMetaDataResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) QueryTSMetaData(tsuid string) (*TSMetaDataResponse, error) {
	tsuid = strings.TrimSpace(tsuid)
	if len(tsuid) == 0 {
		return nil, errors.New("The given query tsuid is empty.")
	}
	queryTSMetaEndpoint := fmt.Sprintf("%s%s?tsuid=%s", c.tsdbEndpoint, TSMetaDataPath, tsuid)
	tsMetaDataResp := TSMetaDataResponse{}
	if err := c.sendRequest(GetMethod, queryTSMetaEndpoint, "", &tsMetaDataResp); err != nil {
		return nil, err
	}
	return &tsMetaDataResp, nil
}

func (c *clientImpl) UpdateTSMetaData(tsMetaData TSMetaData) (*TSMetaDataResponse, error) {
	return c.operateTSMetaData(PostMethod, &tsMetaData)
}

func (c *clientImpl) DeleteTSMetaData(tsMetaData TSMetaData) (*TSMetaDataResponse, error) {
	return c.operateTSMetaData(DeleteMethod, &tsMetaData)
}

func (c *clientImpl) operateTSMetaData(method string, tsMetaData *TSMetaData) (*TSMetaDataResponse, error) {
	if !c.isValidOperateMethod(method) {
		return nil, errors.New("The given method for operating a uid metadata is invalid.")
	}
	tsMetaEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, TSMetaDataPath)
	resultBytes, err := json.Marshal(tsMetaData)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to marshal uidMetaData: %v", err))
	}
	tsMetaDataResp := TSMetaDataResponse{}
	if err = c.sendRequest(method, tsMetaEndpoint, string(resultBytes), &tsMetaDataResp); err != nil {
		return nil, err
	}
	return &tsMetaDataResp, nil
}
