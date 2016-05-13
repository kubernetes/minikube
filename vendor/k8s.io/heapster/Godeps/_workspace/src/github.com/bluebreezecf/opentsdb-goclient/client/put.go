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
// put.go contains the structs and methods for the implementation of /api/put.
//
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// DataPoint is the structure used to hold
// the values of a metric item. Each attributes
// in DataPoint matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/put.html).
//
type DataPoint struct {
	// The name of the metric which is about to be stored, and is required with non-empty value.
	Metric string `json:"metric"`

	// A Unix epoch style timestamp in seconds or milliseconds.
	// The timestamp must not contain non-numeric characters.
	// One can use time.Now().Unix() to set this attribute.
	// This attribute is also required with non-zero value.
	Timestamp int64 `json:"timestamp"`

	// The real type of Value only could be int, int64, float64, or string, and is required.
	Value interface{} `json:"value"`

	// A map of tag name/tag value pairs. At least one pair must be supplied.
	// Don't use too many tags, keep it to a fairly small number, usually up to 4 or 5 tags
	// (By default, OpenTSDB supports a maximum of 8 tags, which can be modified by add
	// configuration item 'tsd.storage.max_tags' in opentsdb.conf).
	Tags map[string]string `json:"tags"`
}

func (data *DataPoint) String() string {
	content, _ := json.Marshal(data)
	return string(content)
}

// PutError holds the error message for each putting DataPoint instance.
// Only calling PUT() with "details" query parameter, the reponse of
// the failed put data operation can contain an array PutError instance
// to show the details for each failure.
type PutError struct {
	Data     DataPoint `json:"datapoint"`
	ErrorMsg string    `json:"error"`
}

func (putErr *PutError) String() string {
	return fmt.Sprintf("%s:%s", putErr.ErrorMsg, putErr.Data.String())
}

// PutResponse acts as the implementation of Response
// in the /api/put scene.
// It holds the status code and the response values defined in
// the (http://opentsdb.net/docs/build/html/api_http/put.html).
type PutResponse struct {
	StatusCode int
	Failed     int64      `json:"failed"`
	Success    int64      `json:"success"`
	Errors     []PutError `json:"errors,omitempty"`
}

func (putResp *PutResponse) SetStatus(code int) {
	putResp.StatusCode = code
}

func (putResp *PutResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(putResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (putResp *PutResponse) GetCustomParser() func(respCnt []byte) error {
	return nil
}

func (c *clientImpl) Put(datas []DataPoint, queryParam string) (*PutResponse, error) {
	err := validateDataPoint(datas)
	if err != nil {
		return nil, err
	}
	if !isValidPutParam(queryParam) {
		return nil, errors.New("The given query param is invalid.")
	}
	var putEndpoint = ""
	if !isEmptyPutParam(queryParam) {
		putEndpoint = fmt.Sprintf("%s%s?%s", c.tsdbEndpoint, PutPath, queryParam)
	} else {
		putEndpoint = fmt.Sprintf("%s%s", c.tsdbEndpoint, PutPath)
	}

	dataGroups, err := c.splitProperGroups(datas)
	if err != nil {
		return nil, err
	}

	responses := make([]PutResponse, 0)
	for _, datapoints := range dataGroups {
		// The datas have been marshalled successfully in splitProperGroups(),
		// so now the returned error is always nil.
		reqBodyCnt, _ := getPutBodyContents(datapoints)
		putResp := PutResponse{}
		if err = c.sendRequest(PostMethod, putEndpoint, reqBodyCnt, &putResp); err != nil {
			// This kind of error only occurs during the process of sending request,
			// not including the scene of inserting datapoints into opentsdb.
			// So just return error once it happens.
			return nil, err
		}
		responses = append(responses, putResp)
	}

	globalResp := PutResponse{}
	globalResp.StatusCode = 200
	for _, resp := range responses {
		globalResp.Failed = globalResp.Failed + resp.Failed
		globalResp.Success = globalResp.Success + resp.Success
		globalResp.Errors = append(globalResp.Errors, resp.Errors...)
		if resp.StatusCode != 200 && globalResp.StatusCode == 200 {
			globalResp.StatusCode = resp.StatusCode
		}
	}
	if globalResp.StatusCode == 200 {
		return &globalResp, nil
	}
	return nil, parsePutErrorMsg(&globalResp)
}

// splitProperGroups splits the given datapoints into groups, whose content size is
// not larger than c.opentsdbCfg.MaxContentLength.
// This method is an assurement of avoiding Put failure, when the content length of
// the given datapoints in a single /api/put request exceeded the value of
// tsd.http.request.max_chunk in the opentsdb config file.
func (c *clientImpl) splitProperGroups(datapoints []DataPoint) ([][]DataPoint, error) {
	datasBytes, err := json.Marshal(&datapoints)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal the datapoints to be put: %v", err)
	}
	datapointGroups := make([][]DataPoint, 0)
	if len(datasBytes) > c.opentsdbCfg.MaxContentLength {
		datapointsSize := len(datapoints)
		endIndex := datapointsSize
		if endIndex > c.opentsdbCfg.MaxPutPointsNum {
			endIndex = c.opentsdbCfg.MaxPutPointsNum
		}
		startIndex := 0
		for endIndex <= datapointsSize {
			tempdps := datapoints[startIndex:endIndex]
			tempSize := len(tempdps)
			// After successful unmarshal, the above marshal is definitly without error
			tempdpsBytes, _ := json.Marshal(&tempdps)
			if len(tempdpsBytes) <= c.opentsdbCfg.MaxContentLength {
				datapointGroups = append(datapointGroups, tempdps)
				startIndex = endIndex
				endIndex = startIndex + tempSize
				if endIndex > datapointsSize {
					endIndex = datapointsSize
				}
			} else {
				endIndex = endIndex - c.opentsdbCfg.DetectDeltaNum
			}
			if startIndex >= datapointsSize {
				break
			}
		}
	} else {
		datapointGroups = append(datapointGroups, datapoints)
	}
	return datapointGroups, nil
}

func parsePutErrorMsg(resp *PutResponse) error {
	buf := bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("Failed to put %d datapoint(s) into opentsdb, statuscode %d:\n", resp.Failed, resp.StatusCode))
	if len(resp.Errors) > 0 {
		for _, putError := range resp.Errors {
			buf.WriteString(fmt.Sprintf("\t%s\n", putError.String()))
		}
	}
	return errors.New(buf.String())
}

func getPutBodyContents(datas []DataPoint) (string, error) {
	if len(datas) == 1 {
		result, err := json.Marshal(datas[0])
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed to marshal datapoint: %v", err))
		}
		return string(result), nil
	} else {
		reqBodyCnt, err := marshalDataPoints(datas)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Failed to marshal datapoint: %v", err))
		}
		return reqBodyCnt, nil
	}
}

func marshalDataPoints(datas []DataPoint) (string, error) {
	buffer := bytes.NewBuffer(nil)
	size := len(datas)
	buffer.WriteString("[")
	for index, item := range datas {
		result, err := json.Marshal(item)
		if err != nil {
			return "", err
		}
		buffer.Write(result)
		if index < size-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")
	return buffer.String(), nil
}

func validateDataPoint(datas []DataPoint) error {
	if datas == nil || len(datas) == 0 {
		return errors.New("The given datapoint is empty.")
	}
	for _, data := range datas {
		if !isValidDataPoint(&data) {
			return errors.New("The value of the given datapoint is invalid.")
		}
	}
	return nil
}

func isValidDataPoint(data *DataPoint) bool {
	if data.Metric == "" || data.Timestamp == 0 || len(data.Tags) < 1 || data.Value == nil {
		return false
	}
	switch data.Value.(type) {
	case int64:
		return true
	case int:
		return true
	case float64:
		return true
	case string:
		return true
	default:
		return false
	}
}

func isValidPutParam(param string) bool {
	if isEmptyPutParam(param) {
		return true
	}
	param = strings.TrimSpace(param)
	if param != PutRespWithSummary && param != PutRespWithDetails {
		return false
	}
	return true
}

func isEmptyPutParam(param string) bool {
	return strings.TrimSpace(param) == ""
}
