// +build linux

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

package kvm

import (
	"testing"
)

var (
	emptyFile           = []byte(``)
	fileWithInvalidJSON = []byte(`{`)
	fileWithNoStatus    = []byte(`[
		
	]`)
	fileWithStatus = []byte(`[
		{
			"ip-address": "1.2.3.5",
			"mac-address": "a4:b5:c6:d7:e8:f9",
			"hostname": "host2",
			"client-id": "01:44:59:e7:fd:f4:d6",
			"expiry-time": 1558638717
		},
		{
			"ip-address": "1.2.3.4",
			"mac-address": "a1:b2:c3:d4:e5:f6",
			"hostname": "host1",
			"client-id": "01:ec:97:de:a2:86:81",
			"expiry-time": 1558639092
		}
	]`)
)

func TestParseStatusAndReturnIp(t *testing.T) {
	type args struct {
		mac      string
		statuses []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"emptyFile",
			args{"a1:b2:c3:d4:e5:f6", emptyFile},
			"",
			false,
		},
		{
			"fileWithStatus",
			args{"a1:b2:c3:d4:e5:f6", fileWithStatus},
			"1.2.3.4",
			false,
		},
		{
			"fileWithNoStatus",
			args{"a4:b5:c6:d7:e8:f9", fileWithNoStatus},
			"",
			false,
		},
		{
			"fileWithInvalidJSON",
			args{"a4:b5:c6:d7:e8:f9", fileWithInvalidJSON},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStatusAndReturnIP(tt.args.mac, tt.args.statuses)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStatusAndReturnIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseStatusAndReturnIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
