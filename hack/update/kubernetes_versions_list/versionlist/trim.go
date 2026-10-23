/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package versionlist

import (
	"fmt"

	"golang.org/x/mod/semver"
)

const (
	supportedKubernetesMinorVersions   = 3
	recentlyEOLKubernetesMinorVersions = 1
)

// TrimKubernetesVersions keeps the default minor, the two preceding supported minors, and the most recent EOL minor.
// versions must be sorted from newest to oldest.
func TrimKubernetesVersions(versions []string, defaultVersion string) ([]string, error) {
	if !semver.IsValid(defaultVersion) {
		return nil, fmt.Errorf("invalid default Kubernetes version %q", defaultVersion)
	}

	defaultMinor := semver.MajorMinor(defaultVersion)
	started := false
	lastMinor := ""
	oldestMinor := ""
	retainedMinors := 0
	minorLimit := supportedKubernetesMinorVersions + recentlyEOLKubernetesMinorVersions
	for _, version := range versions {
		minor := semver.MajorMinor(version)
		if !started {
			if minor != defaultMinor {
				continue
			}
			started = true
		}
		if minor == lastMinor {
			continue
		}
		lastMinor = minor
		oldestMinor = minor
		retainedMinors++
		if retainedMinors == minorLimit {
			break
		}
	}
	if !started {
		return nil, fmt.Errorf("default Kubernetes minor version %s is missing from release list", defaultMinor)
	}

	minimumMinor := oldestMinor + ".0"
	maximumMinor := defaultMinor + ".0"
	trimmed := make([]string, 0, len(versions))
	for _, version := range versions {
		minor := semver.MajorMinor(version) + ".0"
		if semver.Compare(minor, minimumMinor) >= 0 && semver.Compare(minor, maximumMinor) <= 0 {
			trimmed = append(trimmed, version)
		}
	}
	return trimmed, nil
}
