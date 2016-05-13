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
// config.go contains the structs and methods for the implementation of /api/config and /api/config/filters.
//
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type ConfigResponse struct {
	StatusCode int
	Configs    map[string]string `json:"configs"`
}

func (cfgResp *ConfigResponse) SetStatus(code int) {
	cfgResp.StatusCode = code
}

func (cfgResp *ConfigResponse) GetCustomParser() func(respCnt []byte) error {
	return func(respCnt []byte) error {
		return json.Unmarshal([]byte(fmt.Sprintf("{%s:%s}", `"Configs"`, string(respCnt))), &cfgResp)
	}
}

func (cfgResp *ConfigResponse) String() string {
	buffer := bytes.NewBuffer(nil)
	content, _ := json.Marshal(cfgResp)
	buffer.WriteString(fmt.Sprintf("%s\n", string(content)))
	return buffer.String()
}

func (c *clientImpl) Config() (*ConfigResponse, error) {
	configEndpoint := fmt.Sprintf("%s%s", c.tsdbEndpoint, ConfigPath)
	cfgResp := ConfigResponse{}
	if err := c.sendRequest(GetMethod, configEndpoint, "", &cfgResp); err != nil {
		return nil, err
	}
	return &cfgResp, nil
}
