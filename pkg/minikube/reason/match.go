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

package reason

import (
	"regexp"

	"github.com/golang/glog"
)

// match matches a known issue within minikube
type match struct {
	// Inherit ID, ExitCode, and Style from reason.Kind
	Kind

	// Regexp is which regular expression this issue matches
	Regexp *regexp.Regexp
	// Operating systems this error is specific to
	GOOS []string
}

func knownIssues() []match {
	ps := []match{}
	// This is intentionally in dependency order
	ps = append(ps, programIssues...)
	ps = append(ps, resourceIssues...)
	ps = append(ps, hostIssues...)
	ps = append(ps, providerIssues...)
	ps = append(ps, driverIssues...)
	ps = append(ps, localNetworkIssues...)
	ps = append(ps, internetIssues...)
	ps = append(ps, guestIssues...)
	ps = append(ps, runtimeIssues...)
	ps = append(ps, controlPlaneIssues...)
	ps = append(ps, serviceIssues...)
	return ps
}

// FindMatch returns a known issue from an error on an OS
func MatchKnownIssue(r Kind, err error, goos string) *Kind {
	// The kind passed in has specified that it should not be rematched
	if r.NoMatch {
		return nil
	}

	var genericMatch *Kind

	for _, ki := range knownIssues() {
		ki := ki
		if ki.Regexp == nil {
			glog.Errorf("known issue has no regexp: %+v", ki)
			continue
		}

		if !ki.Regexp.MatchString(err.Error()) {
			continue
		}

		// Does this match require an OS matchup?
		if len(ki.GOOS) > 0 {
			for _, o := range ki.GOOS {
				if o == goos {
					return &ki.Kind
				}
			}
		}
		if genericMatch == nil {
			genericMatch = &ki.Kind
		}
	}

	return genericMatch
}
