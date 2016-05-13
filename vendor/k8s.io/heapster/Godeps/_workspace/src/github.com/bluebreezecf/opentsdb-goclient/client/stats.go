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
// stats.go contains the structs and methods for the implementation of /api/stats.
//
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type StatsResponse struct {
	StatusCode int
	Metrics    []MetricInfo `json:"Metrics"`
}

type MetricInfo struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     interface{}       `json:"value"`
	Tags      map[string]string `json:"tags"`
}

func (statsResp *StatsResponse) SetStatus(code int) {
	statsResp.StatusCode = code
}

func (statsResp *StatsResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		return json.Unmarshal([]byte(fmt.Sprintf("{%s:%s}", `"Metrics"`, string(respCnt))), &statsResp)
	}
}

func (statsResp *StatsResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(statsResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) Stats() (*StatsResponse, error) {
	statsEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, StatsPath)
	statsResp := StatsResponse{}
	if err := c.sendRequest(GetMethod, statsEndpoint, "", &statsResp); err != nil {
		return nil, err
	}
	return &statsResp, nil
}
