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
// suggest.go contains the structs and methods for the implementation of /api/suggest.
//
package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// SuggestParam is the structure used to hold
// the querying parameters when calling /api/suggest.
// Each attributes in SuggestParam matches the definition in
// (http://opentsdb.net/docs/build/html/api_http/suggest.html).
//
type SuggestParam struct {
	// The type of data to auto complete on.
	// Must be one of the following: metrics, tagk or tagv.
	// It is required.
	// Only the one of the three query type can be used:
	// TypeMetrics, TypeTagk, TypeTagv
	Type string `json:"type"`

	// An optional string value to match on for the given type
	Q string `json:"q,omitempty"`

	// An optional integer value presenting the maximum number of suggested
	// results to return. If it is set, it must be greater than 0.
	MaxResultNum int `json:"max,omitempty"`
}

func (sugParam *SuggestParam) String() string {
	contents, _ := json.Marshal(sugParam)
	return string(contents)
}

type SuggestResponse struct {
	StatusCode int
	ResultInfo []string `json:"ResultInfo"`
}

func (sugResp *SuggestResponse) SetStatus(code int) {
	sugResp.StatusCode = code
}

func (sugResp *SuggestResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		return json.Unmarshal([]byte(fmt.Sprintf("{%s:%s}", `"ResultInfo"`, string(respCnt))), &sugResp)
	}
}

func (sugResp *SuggestResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(sugResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) Suggest(sugParam SuggestParam) (*SuggestResponse, error) {
	if !isValidSuggestParam(&sugParam) {
		return nil, errors.New("The given suggest param is invalid.\n")
	}
	sugEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, SuggestPath)
	reqBodyCnt, err := getSuggestBodyContents(&sugParam)
	if err != nil {
		return nil, err
	}
	fmt.Println(reqBodyCnt)
	sugResp := SuggestResponse{}
	if err := c.sendRequest(PostMethod, sugEndpoint, reqBodyCnt, &sugResp); err != nil {
		return nil, err
	}
	return &sugResp, nil
}

func isValidSuggestParam(sugParam *SuggestParam) bool {
	if sugParam.Type == "" {
		return false
	}
	types := []string{TypeMetrics, TypeTagk, TypeTagv}
	sugParam.Type = strings.TrimSpace(sugParam.Type)
	for _, typeItem := range types {
		if sugParam.Type == typeItem {
			return true
		}
	}
	return false
}

func getSuggestBodyContents(sugParam *SuggestParam) (string, error) {
	result, err := json.Marshal(sugParam)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to marshal suggest param: %v\n", err))
	}
	return string(result), nil
}
