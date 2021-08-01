/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

import "github.com/blang/semver/v4"

// K8sIssue represents a known issue with a particular version of Kubernetes
type K8sIssue struct {
	// VersionAffected is the list of Kubernetes versions that has a particular issue
	VersionsAffected []string
	// Description is what will be printed to the user describing the issue
	Description string
	// URL is a link to an issue or documentation about the issue, optional.
	URL string
}

var k8sIssues = []K8sIssue{
	{
		// fixed by github.com/kubernetes/kubernetes/pull/99336
		VersionsAffected: []string{
			"1.18.16",
			"1.18.17",
			"1.18.18",
			"1.19.8",
			"1.19.9",
			"1.19.10",
			"1.20.3",
			"1.20.4",
			"1.20.5",
			"1.20.6",
			"1.21.0",
		},
		Description: "Kubernetes {{.version}} has a known performance issue on cluster startup. It might take 2 to 3 minutes for a cluster to start.",
		URL:         "https://github.com/kubernetes/kubeadm/issues/2395",
	},
}

// ProblematicK8sVersion checks for the supplied Kubernetes version and checks if there's a known issue with it.
func ProblematicK8sVersion(v semver.Version) *K8sIssue {
	for _, issue := range k8sIssues {
		for _, va := range issue.VersionsAffected {
			if va == v.String() {
				return &issue
			}
		}
	}
	return nil
}
