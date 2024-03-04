/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"context"
	"net/http"

	"github.com/google/go-github/v60/github"
	"golang.org/x/mod/semver"
	"k8s.io/minikube/pkg/minikube/constants"
)

// supportedKubernetesVersions returns reverse-sort supported Kubernetes releases that are in [constants.OldestKubernetesVersion, constants.NewestKubernetesVersion] range, including prereleases.
func supportedKubernetesVersions() (releases []string) {
	minver := constants.OldestKubernetesVersion
	maxver := constants.NewestKubernetesVersion

	for _, ver := range constants.ValidKubernetesVersions {
		if (minver != "" && semver.Compare(minver, ver) == 1) || (maxver != "" && semver.Compare(ver, maxver) == 1) {
			continue
		}
		releases = append(releases, ver)
	}
	return releases
}

// IsInGitHubKubernetesVersions checks whether ver is in the GitHub list of K8s versions
func IsInGitHubKubernetesVersions(ver string) (bool, error) {
	ghc := github.NewClient(nil)

	_, resp, err := ghc.Repositories.GetReleaseByTag(context.Background(), "kubernetes", "kubernetes", ver)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
