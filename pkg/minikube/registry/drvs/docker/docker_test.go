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

package docker

import (
	"fmt"
	"testing"
)

type testCase struct {
	version, expect string
}

func appendVersionVariations(tc []testCase, v []int, reason string) []testCase {
	appendedTc := append(tc, testCase{
		version: fmt.Sprintf("linux-%02d.%02d", v[0], v[1]),
		expect:  reason,
	})

	// postfix string for unstable channel or patch. https://docs.docker.com/engine/install/
	patchPostFix := "20180720214833-f61e0f7"

	vs := fmt.Sprintf("%02d.%02d.%d", v[0], v[1], v[2])
	appendedTc = append(appendedTc, []testCase{
		{
			version: fmt.Sprintf("linux-%s", vs),
			expect:  reason,
		},
		{
			version: fmt.Sprintf("linux-%s-%s", vs, patchPostFix),
			expect:  reason,
		},
	}...,
	)

	return appendedTc
}

func TestCheckDockerVersion(t *testing.T) {
	tc := []testCase{
		{
			version: "windows-20.0.1",
			expect:  "PROVIDER_DOCKER_WINDOWS_CONTAINERS",
		},
		{
			version: fmt.Sprintf("linux-%02d.%02d", minDockerVersion[0], minDockerVersion[1]),
			expect:  "",
		},
		{
			version: fmt.Sprintf("linux-%02d.%02d.%02d", minDockerVersion[0], minDockerVersion[1], minDockerVersion[2]),
			expect:  "",
		},
	}

	for i := 0; i < 3; i++ {
		v := make([]int, 3)
		copy(v, minDockerVersion)

		v[i] = minDockerVersion[i] + 1
		tc = appendVersionVariations(tc, v, "")

		v[i] = minDockerVersion[i] - 1
		if v[2] < 0 {
			// skip test if patch version is negative number.
			continue
		}
		tc = appendVersionVariations(tc, v, "PROVIDER_DOCKER_VERSION_LOW")
	}

	for _, c := range tc {
		t.Run("checkDockerVersion test", func(t *testing.T) {
			s := checkDockerVersion(c.version)
			if c.expect != s.Reason {
				t.Errorf("Error %v expected. but got %q. (version string : %s)", c.expect, s.Reason, c.version)
			}
		})
	}
}
