/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package update

import (
	"context"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/google/go-github/v43/github"
)

const (
	// ghListPerPage uses max value (100) for PerPage to avoid hitting the rate limits.
	// (ref: https://pkg.go.dev/github.com/google/go-github/github#hdr-Rate_Limiting)
	ghListPerPage = 100

	// ghSearchLimit limits the number of searched items to be <= N * ghListPerPage.
	ghSearchLimit = 300
)

// GHReleases returns greatest current stable release and greatest latest rc or beta pre-release from GitHub owner/repo repository, and any error occurred.
// If latest pre-release version is lower than the current stable release, then it will return current stable release for both.
func GHReleases(ctx context.Context, owner, repo string) (stable, latest, edge string, err error) {
	ghc := github.NewClient(nil)

	// walk through the paginated list of up to ghSearchLimit newest releases
	opts := &github.ListOptions{PerPage: ghListPerPage}
	for (opts.Page+1)*ghListPerPage <= ghSearchLimit {
		rls, resp, err := ghc.Repositories.ListReleases(ctx, owner, repo, opts)
		if err != nil {
			return "", "", "", err
		}
		for _, rl := range rls {
			ver := rl.GetTagName()
			if !semver.IsValid(ver) {
				continue
			}
			// check if ver version is release (ie, 'v1.19.2') or pre-release (ie, 'v1.19.3-rc.0' or 'v1.19.0-beta.2')
			prerls := semver.Prerelease(ver)
			if prerls == "" {
				if semver.Compare(ver, stable) == 1 {
					stable = ver
				}
			} else if strings.HasPrefix(prerls, "-rc") || strings.HasPrefix(prerls, "-beta") {
				if semver.Compare(ver, latest) == 1 {
					latest = ver
				}
			} else if strings.Contains(prerls, "-alpha") {
				if semver.Compare(ver, edge) == 1 {
					edge = ver
				}
			}

			// make sure that latest >= stable
			if semver.Compare(latest, stable) == -1 {
				latest = stable
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return stable, latest, edge, nil
}
