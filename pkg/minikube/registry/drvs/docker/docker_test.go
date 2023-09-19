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
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/registry"
)

type testCase struct {
	version, expect   string
	expectFixContains string
}

func appendVersionVariations(tc []testCase, v []int, reason string) []testCase {
	appendedTc := tc
	appendedTc = append(appendedTc, testCase{
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

func stringToIntSlice(t *testing.T, s string) []int {
	sem, err := semver.ParseTolerant(s)
	if err != nil {
		t.Fatalf("failed to parse %s to semver: %v", s, err)
	}
	return []int{int(sem.Major), int(sem.Minor), int(sem.Patch)}
}

func TestCheckDockerEngineVersion(t *testing.T) {
	recParts := stringToIntSlice(t, recommendedDockerVersion)
	minParts := stringToIntSlice(t, minDockerVersion)

	tc := []testCase{
		{
			version: "windows-20.0.1",
			expect:  "PROVIDER_DOCKER_WINDOWS_CONTAINERS",
		},
		{
			version: fmt.Sprintf("linux-%02d.%02d", recParts[0], recParts[1]),
			expect:  "",
		},
		{
			version: fmt.Sprintf("linux-%s", recommendedDockerVersion),
			expect:  "",
		},
	}

	for i := 0; i < 3; i++ {
		v := make([]int, 3)
		copy(v, minParts)

		v[i] = minParts[i] + 1
		tc = appendVersionVariations(tc, v, "")

		v[i] = minParts[i] - 1
		if v[2] < 0 {
			// skip test if patch version is negative number.
			continue
		}
		tc = appendVersionVariations(tc, v, "PROVIDER_DOCKER_VERSION_LOW")
	}

	recommendedSupported := fmt.Sprintf("Minimum recommended version is %s, minimum supported version is %s", recommendedDockerVersion, minDockerVersion)
	install := fmt.Sprintf("Install the official release of %s (%s, current version is %%s)", driver.FullName(driver.Docker), recommendedSupported)
	update := fmt.Sprintf("Upgrade %s to a newer version (%s, current version is %%s)", driver.FullName(driver.Docker), recommendedSupported)
	tc = append(tc, []testCase{
		{
			// "dev" is set when Docker (Moby) was installed with `make binary && make install`
			version:           "linux-dev",
			expect:            "",
			expectFixContains: fmt.Sprintf(install, "dev"),
		},
		{
			// "library-import" is set when Docker (Moby) was installed with `go build github.com/docker/docker/cmd/dockerd` (unrecommended, but valid)
			version:           "linux-library-import",
			expect:            "",
			expectFixContains: fmt.Sprintf(install, "library-import"),
		},
		{
			// "foo.bar.baz" is a triplet that cannot be parsed as "%02d.%02d.%d"
			version:           "linux-foo.bar.baz",
			expect:            "",
			expectFixContains: fmt.Sprintf(install, "foo.bar.baz"),
		},
		{
			// "linux-18.09.9" is older than minimum recommended version
			version:           "linux-18.09.9",
			expect:            "",
			expectFixContains: fmt.Sprintf(update, "18.09.9"),
		},
		{
			// "linux-18.06.2" is older than minimum required version
			version:           "linux-18.06.2",
			expect:            "PROVIDER_DOCKER_VERSION_LOW",
			expectFixContains: fmt.Sprintf(update, "18.06.2"),
		},
	}...)

	for _, c := range tc {
		t.Run("checkDockerEngineVersion test", func(t *testing.T) {
			s := checkDockerEngineVersion(c.version)
			if s.Error != nil {
				if c.expect != s.Reason {
					t.Errorf("Error %v expected. but got %q. (version string : %s)", c.expect, s.Reason, c.version)
				}
			}
			if c.expectFixContains != "" {
				if !strings.Contains(s.Fix, c.expectFixContains) {
					t.Errorf("Error expected Fix to contain %q, but got %q", c.expectFixContains, s.Fix)
				}
			}
		})
	}
}

func TestCheckDockerDesktopVersion(t *testing.T) {
	tests := []struct {
		input             string
		shouldReturnError bool
	}{
		{"Docker Desktop", false},
		{"Cat Desktop 4.16.0", false},
		{"Docker Playground 4.16.0", false},
		{"Docker Desktop 4.15.0", false},
		{"Docker Desktop 4.16.0", true},
		{"  Docker  Desktop  4.16.0  ", true},
	}
	for _, tt := range tests {
		state := checkDockerDesktopVersion(tt.input)
		err := state.Error
		if (err == nil && tt.shouldReturnError) || (err != nil && !tt.shouldReturnError) {
			t.Errorf("checkDockerDesktopVersion(%q) = %+v; expected shouldReturnError = %t", tt.input, state, tt.shouldReturnError)
		}
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		input             string
		shouldReturnError bool
	}{
		{"linux-20.10.22:Docker Desktop 4.16.2 (95914)", false},
		{"noDashHere:Docker Desktop 4.16.2 (95914)", true},
		{"linux-20.10.22:Docker Desktop 4.16.0 (95914)", true},
		{"", true},
	}
	for _, tt := range tests {
		dockerVersionOrState = func() (string, registry.State) { return tt.input, registry.State{} }
		oci.CachedDaemonInfo = func(string) (oci.SysInfo, error) { return oci.SysInfo{}, nil }
		state := status()
		err := state.Error
		if (err == nil && tt.shouldReturnError) || (err != nil && !tt.shouldReturnError) {
			t.Errorf("status(%q) = %+v; expected shouldReturnError = %t", tt.input, state, tt.shouldReturnError)
		}
	}
}
