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
	"sort"

	"github.com/google/go-github/v43/github"
	"golang.org/x/mod/semver"
	"k8s.io/minikube/pkg/minikube/constants"
)

// supportedKubernetesVersions returns reverse-sort supported Kubernetes releases from GitHub that are in [constants.OldestKubernetesVersion, constants.NewestKubernetesVersion] range, including prereleases.
// in case it cannot get it from GitHub, in addition to [constants.NewestKubernetesVersion, constants.OldestKubernetesVersion], 'constants.DefaultKubernetesVersion' is also returned if different from 'constants.NewestKubernetesVersion'.
func supportedKubernetesVersions() (releases []string) {
	minver := constants.OldestKubernetesVersion
	defver := constants.DefaultKubernetesVersion
	maxver := constants.NewestKubernetesVersion

	ghc := github.NewClient(nil)

	opts := &github.ListOptions{PerPage: 100}
	for (opts.Page+1)*100 <= 300 {
		rls, resp, err := ghc.Repositories.ListReleases(context.Background(), "kubernetes", "kubernetes", opts)
		if err != nil {
			v := []string{maxver}
			if defver != maxver {
				v = append(v, defver)
			}
			v = append(v, minver)
			return v
		}
		for _, rl := range rls {
			ver := rl.GetTagName()
			if !semver.IsValid(ver) {
				continue
			}
			// skip out-of-range versions
			if (minver != "" && semver.Compare(minver, ver) == 1) || (maxver != "" && semver.Compare(ver, maxver) == 1) {
				continue
			}
			releases = append(releases, ver)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	sort.Slice(releases, func(i, j int) bool { return semver.Compare(releases[i], releases[j]) == 1 })
	return releases
}
