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
// aggregators.go contains the structs and methods for the implementation of /api/aggregators.
//
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// AggregatorsResponse acts as the implementation of Response in the /api/aggregators scene.
// It holds the status code and the response values defined in the
// (http://opentsdb.net/docs/build/html/api_http/aggregators.html).
//
type AggregatorsResponse struct {
	StatusCode  int
	Aggregators []string `json:"aggregators"`
}

func (aggreResp *AggregatorsResponse) SetStatus(code int) {
	aggreResp.StatusCode = code
}

func (aggreResp *AggregatorsResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		return json.Unmarshal([]byte(fmt.Sprintf("{%s:%s}", `"aggregators"`, string(respCnt))), &aggreResp)
	}
}

func (aggreResp *AggregatorsResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(aggreResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) Aggregators() (*AggregatorsResponse, error) {
	aggregatorsEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, AggregatorPath)
	aggreResp := AggregatorsResponse{}
	if err := c.sendRequest(GetMethod, aggregatorsEndpoint, "", &aggreResp); err != nil {
		return nil, err
	}
	return &aggreResp, nil
}
