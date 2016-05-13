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
// dropcaches.go contains the structs and methods for the implementation of /api/dropcaches.
//
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type DropcachesResponse struct {
	StatusCode     int
	DropcachesInfo map[string]string `json:"DropcachesInfo"`
}

func (dropResp *DropcachesResponse) SetStatus(code int) {
	dropResp.StatusCode = code
}

func (dropResp *DropcachesResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		return json.Unmarshal([]byte(fmt.Sprintf("{%s:%s}", `"DropcachesInfo"`, string(respCnt))), &dropResp)
	}
}

func (dropResp *DropcachesResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(dropResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) Dropcaches() (*DropcachesResponse, error) {
	dropEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, DropcachesPath)
	dropResp := DropcachesResponse{}
	if err := c.sendRequest(GetMethod, dropEndpoint, "", &dropResp); err != nil {
		return nil, err
	}
	return &dropResp, nil
}
