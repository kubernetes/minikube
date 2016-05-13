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
// annotation.go contains the structs and methods for the implementation of
// /api/annotation and /api/annotation/bulk.
//
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Annotation is the structure used to hold
// the querying parameters when calling /api/annotation.
// Each attributes in Annotation matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/annotation/index.html).
//
// Annotations are very basic objects used to record a note of an arbitrary
// event at some point, optionally associated with a timeseries. Annotations
// are not meant to be used as a tracking or event based system, rather they
// are useful for providing links to such systems by displaying a notice on
// graphs or via API query calls.
//
type Annotation struct {
	// A Unix epoch timestamp, in seconds, marking the time when the annotation event should be recorded.
	// The value is required with non-zero value.
	StartTime int64 `json:"startTime,omitempty"`

	// An optional end time for the event if it has completed or been resolved.
	EndTime int64 `json:"endTime,omitempty"`

	// A TSUID if the annotation is associated with a timeseries.
	// This may be optional if the note was for a global event
	Tsuid string `json:"tsuid,omitempty"`

	// An optional brief description of the event. As this may appear on GnuPlot graphs,
	// the description should be very short, ideally less than 25 characters.
	Description string `json:"description,omitempty"`

	// An optional detailed notes about the event
	Notes string `json:"notes,omitempty"`

	// An optional key/value map to store custom fields and values
	Custom map[string]string `json:"custom,omitempty"`
}

// AnnotationResponse acts as the implementation of Response in the /api/annotation scene.
// It holds the status code and the response values defined in the
// (http://opentsdb.net/docs/build/html/api_http/aggregators.html).
//
type AnnotationResponse struct {
	StatusCode int
	Annotation
	ErrorInfo map[string]interface{} `json:"error,omitempty"`
}

func (annotResp *AnnotationResponse) SetStatus(code int) {
	annotResp.StatusCode = code
}

func (annotResp *AnnotationResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		originContents := string(respCnt)
		var resultBytes []byte
		if strings.Contains(originContents, "startTime") ||
			strings.Contains(originContents, "error") {
			resultBytes = respCnt
		} else if annotResp.StatusCode == 204 {
			// The OpenTSDB deletes an annotation successfully and with no body content.
			return nil
		}
		return json.Unmarshal(resultBytes, &annotResp)
	}
}

func (annotResp *AnnotationResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(annotResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) QueryAnnotation(queryAnnoParam map[string]interface{}) (*AnnotationResponse, error) {
	if queryAnnoParam == nil || len(queryAnnoParam) == 0 {
		return nil, errors.New("The given query annotation param is nil")
	}
	buffer := bytes.NewBuffer(nil)
	size := len(queryAnnoParam)
	i := 0
	for k, v := range queryAnnoParam {
		buffer.WriteString(fmt.Sprintf("%s=%v", k, v))
		if i < size-1 {
			buffer.WriteString("&")
		} else {
			break
		}
		i++
	}
	annoEndpoint := fmt.Sprintf("%s%s?%s", c.tsdbEndpoint, AnnotationPath, buffer.String())
	annResp := AnnotationResponse{}
	if err := c.sendRequest(GetMethod, annoEndpoint, "", &annResp); err != nil {
		return nil, err
	}
	return &annResp, nil
}

func (c *clientImpl) UpdateAnnotation(annotation Annotation) (*AnnotationResponse, error) {
	return c.operateAnnotation(PostMethod, &annotation)
}

func (c *clientImpl) DeleteAnnotation(annotation Annotation) (*AnnotationResponse, error) {
	return c.operateAnnotation(DeleteMethod, &annotation)
}

func (c *clientImpl) operateAnnotation(method string, annotation *Annotation) (*AnnotationResponse, error) {
	if !c.isValidOperateMethod(method) {
		return nil, errors.New("The given method for operating an annotation is invalid.")
	}
	annoEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, AnnotationPath)
	resultBytes, err := json.Marshal(annotation)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to marshal annotation: %v", err))
	}
	annResp := AnnotationResponse{}
	if err = c.sendRequest(method, annoEndpoint, string(resultBytes), &annResp); err != nil {
		return nil, err
	}
	return &annResp, nil
}

// BulkAnnotatResponse acts as the implementation of Response in the /api/annotation/bulk scene.
// It holds the status code and the response values defined in the
// (http://opentsdb.net/docs/build/html/api_http/annotation/bulk.html)
// for both bulk update and delete scenes.
//
type BulkAnnotatResponse struct {
	StatusCode        int
	UpdateAnnotations []Annotation           `json:"InvolvedAnnotations,omitempty"`
	ErrorInfo         map[string]interface{} `json:"error,omitempty"`
	BulkDeleteResp
}

type BulkAnnoDeleteInfo struct {
	// A list of TSUIDs with annotations that should be deleted. This may be empty
	// or null (for JSON) in which case the global flag should be set.
	Tsuids []string `json:"tsuids,omitempty"`

	// A timestamp for the start of the request.
	StartTime int64 `json:"startTime,omitempty"`

	// An optional end time for the event if it has completed or been resolved.
	EndTime int64 `json:"endTime,omitempty"`

	// An optional flag indicating whether or not global annotations should be deleted for the range
	Global bool `json:"global,omitempty"`
}

type BulkDeleteResp struct {
	BulkAnnoDeleteInfo

	// Total number of annotations to be deleted successfully for current bulk
	// delete operation. The value is only used in the reponse of bulk deleting,
	// not in the bulk deleting parameters.
	TotalDeleted int64 `json:"totalDeleted,omitempty"`
}

func (bulkAnnotResp *BulkAnnotatResponse) SetStatus(code int) {
	bulkAnnotResp.StatusCode = code
}

func (bulkAnnotResp *BulkAnnotatResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		originContents := string(respCnt)
		var resultBytes []byte
		if strings.Contains(originContents, "startTime") {
			resultBytes = []byte(fmt.Sprintf("{%s:%s}", `"InvolvedAnnotations"`, originContents))
		} else if strings.Contains(originContents, "error") || strings.Contains(originContents, "totalDeleted") {
			resultBytes = respCnt
		} else {
			return errors.New(fmt.Sprintf("Unrecognized bulk annotation response info: %s", originContents))
		}
		return json.Unmarshal(resultBytes, &bulkAnnotResp)
	}
}

func (bulkAnnotResp *BulkAnnotatResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(bulkAnnotResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) BulkUpdateAnnotations(annotations []Annotation) (*BulkAnnotatResponse, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, errors.New("The given annotations are empty.")
	}
	bulkAnnoEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, BulkAnnotationPath)
	reqBodyCnt, err := marshalAnnotations(annotations)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to marshal annotations: %v", err))
	}
	bulkAnnoResp := BulkAnnotatResponse{}
	if err = c.sendRequest(PostMethod, bulkAnnoEndpoint, reqBodyCnt, &bulkAnnoResp); err != nil {
		return nil, err
	}
	return &bulkAnnoResp, nil
}

func (c *clientImpl) BulkDeleteAnnotations(bulkDelParam BulkAnnoDeleteInfo) (*BulkAnnotatResponse, error) {
	bulkAnnoEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, BulkAnnotationPath)
	resultBytes, err := json.Marshal(bulkDelParam)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to marshal bulk delete param: %v", err))
	}
	bulkAnnoResp := BulkAnnotatResponse{}
	if err = c.sendRequest(DeleteMethod, bulkAnnoEndpoint, string(resultBytes), &bulkAnnoResp); err != nil {
		return nil, err
	}
	return &bulkAnnoResp, nil
}

func marshalAnnotations(annotations []Annotation) (string, error) {
	buffer := bytes.NewBuffer(nil)
	size := len(annotations)
	buffer.WriteString("[")
	for index, item := range annotations {
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
