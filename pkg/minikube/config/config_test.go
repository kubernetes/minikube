/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"bytes"
	"reflect"
	"testing"
)

type configTestCase struct {
	data   string
	config map[string]interface{}
}

var configTestCases = []configTestCase{
	{
		data: `{
    "memory": 2
}`,
		config: map[string]interface{}{
			"memory": 2,
		},
	},
	{
		data: `{
    "ReminderWaitPeriodInHours": 99,
    "cpus": 4,
    "disk-size": "20g",
    "log_dir": "/etc/hosts",
    "show-libmachine-logs": true,
    "v": 5,
    "vm-driver": "kvm"
}`,
		config: map[string]interface{}{
			"vm-driver":                 "kvm",
			"cpus":                      4,
			"disk-size":                 "20g",
			"v":                         5,
			"show-libmachine-logs":      true,
			"log_dir":                   "/etc/hosts",
			"ReminderWaitPeriodInHours": 99,
		},
	},
}

func TestReadConfig(t *testing.T) {
	for _, tt := range configTestCases {
		r := bytes.NewBufferString(tt.data)
		config, err := decode(r)
		if reflect.DeepEqual(config, tt.config) || err != nil {
			t.Errorf("Did not read config correctly,\n\n wanted %+v, \n\n got %+v", tt.config, config)
		}
	}
}

func TestGet(t *testing.T) {
	cfg := `{
		"key": "val"
	}`

	config, err := decode(bytes.NewBufferString(cfg))
	if err != nil {
		t.Fatalf("Error decoding config : %v", err)
	}

	var testcases = []struct {
		key string
		val string
		err bool
	}{
		{"key", "val", false},
		{"badkey", "", true},
	}

	for _, tt := range testcases {
		val, err := get(tt.key, config)
		if (err != nil) && !tt.err {
			t.Errorf("Error fetching key: %s. Err: %v", tt.key, err)
			continue
		}
		if val != tt.val {
			t.Errorf("Expected %s, got %s", tt.val, val)
			continue
		}
	}
}
