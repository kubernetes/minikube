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
// serializers.go contains the structs and methods for the implementation of /api/serializers.
//
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type SerialResponse struct {
	StatusCode  int
	Serializers []Serializer `json:"Serializers"`
}

type Serializer struct {
	SerializerName string   `json:"serializer"`
	Formatters     []string `json:"formatters"`
	Parsers        []string `json:"parsers"`
	Class          string   `json:"class,omitempty"`
	ResContType    string   `json:"response_content_type,omitempty"`
	ReqContType    string   `json:"request_content_type,omitempty"`
}

func (serialResp *SerialResponse) SetStatus(code int) {
	serialResp.StatusCode = code
}

func (serialResp *SerialResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		return json.Unmarshal([]byte(fmt.Sprintf("{%s:%s}", `"Serializers"`, string(respCnt))), &serialResp)
	}
}

func (serialResp *SerialResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(serialResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) Serializers() (*SerialResponse, error) {
	serialEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, SerializersPath)
	serialResp := SerialResponse{}
	if err := c.sendRequest(GetMethod, serialEndpoint, "", &serialResp); err != nil {
		return nil, err
	}
	return &serialResp, nil
}
