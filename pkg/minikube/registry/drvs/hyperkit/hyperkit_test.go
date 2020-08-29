// +build darwin

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package hyperkit

import (
	"testing"
)

var (
	versionOutputType1 = "hyperkit: v0.20190201-11-gc0dd46\n\nHomepage: https://github.com/docker/hyperkit\nLicense: BSD\n\n"
	versionOutputType2 = "hyperkit: 0.20190201\n\nHomepage: https://github.com/docker/hyperkit\nLicense: BSD\n\n"
)

func TestSplitHyperKitVersion(t *testing.T) {
	tc := []struct {
		desc, version, expect string
	}{
		{
			desc:    "split type1 output to YYYYMMDD format",
			version: "v.20190201-gc0dd46",
			expect:  "20190201",
		},
		{
			desc:    "split type2 output to YYYYMMDD format",
			version: "0.20190201",
			expect:  "20190201",
		},
		{
			desc:    "non split YYYYMMDD output to YYYYMMDD format",
			version: "20190201",
			expect:  "20190201",
		},
		{
			desc:    "split semver output to YYYYMMDD format",
			version: "v1.0.0",
			expect:  "0",
		},
	}
	for _, test := range tc {
		t.Run(test.desc, func(t *testing.T) {
			version := splitHyperKitVersion(test.version)
			if version != test.expect {
				t.Fatalf("Error %v expected but result %v", test.expect, version)
			}
		})
	}
}

func TestConvertVersionToDate(t *testing.T) {
	tc := []struct {
		desc, versionOutput, expect string
	}{
		{
			desc:          "split type1 output to YYYYMMDD format",
			versionOutput: versionOutputType1,
			expect:        "20190201",
		},
		{
			desc:          "split type2 output to YYYYMMDD format",
			versionOutput: versionOutputType2,
			expect:        "20190201",
		},
		{
			desc:          "split semver output to YYYYMMDD format",
			versionOutput: "hyperkit: v1.0.0\n\nHomepage: https://github.com/docker/hyperkit\nLicense: BSD\n\n",
			expect:        "0",
		},
	}
	for _, test := range tc {
		t.Run(test.desc, func(t *testing.T) {
			version := convertVersionToDate(test.versionOutput)
			if version != test.expect {
				t.Fatalf("Error %v expected but result %v", test.expect, version)
			}
		})
	}
}

func TestIsNewerVersion(t *testing.T) {
	tc := []struct {
		desc, currentVersion, specificVersion string
		isNew                                 bool
	}{
		{
			desc:            "version check newer",
			currentVersion:  "29991231",
			specificVersion: "20190802",
			isNew:           true,
		},
		{
			desc:            "version check equal",
			currentVersion:  "20190802",
			specificVersion: "20190802",
			isNew:           true,
		},
		{
			desc:            "version check older",
			currentVersion:  "20190201",
			specificVersion: "20190802",
			isNew:           false,
		},
	}
	for _, test := range tc {
		t.Run(test.desc, func(t *testing.T) {
			isNew, err := isNewerVersion(test.currentVersion, test.specificVersion)
			if err != nil {
				t.Fatalf("Got unexpected error %v", err)
			}
			if isNew != test.isNew {
				t.Fatalf("Error %v expected but result %v", test.isNew, isNew)
			}
		})
	}
}
