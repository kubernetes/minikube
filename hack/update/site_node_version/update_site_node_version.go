/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
)

const (
	// ghListPerPage uses max value (100) for PerPage to avoid hitting the rate limits.
	// (ref: https://pkg.go.dev/github.com/google/go-github/github#hdr-Rate_Limiting)
	ghListPerPage = 100

	// ghSearchLimit limits the number of searched items to be <= N * ghListPerPage.
	ghSearchLimit = 300
)

var schema = map[string]update.Item{
	"netlify.toml": {
		Replace: map[string]string{
			`NODE_VERSION = ".*`: `NODE_VERSION = "{{.Version}}"`,
		},
	},
}

type Data struct {
	Version string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	version, err := latestNodeVersionByMajor(ctx, "v20")
	if err != nil {
		klog.Fatalf("Unable to get stable version: %v", err)
	}
	version = strings.TrimPrefix(version, "v")

	data := Data{Version: version}

	update.Apply(schema, data)
}

func latestNodeVersionByMajor(ctx context.Context, major string) (string, error) {
	ghc := github.NewClient(nil)

	// walk through the paginated list of up to ghSearchLimit newest releases
	opts := &github.ListOptions{PerPage: ghListPerPage}
	for (opts.Page+1)*ghListPerPage <= ghSearchLimit {
		rls, resp, err := ghc.Repositories.ListTags(ctx, "nodejs", "node", opts)
		if err != nil {
			return "", err
		}
		for _, rl := range rls {
			ver := rl.GetName()
			if !semver.IsValid(ver) {
				continue
			}
			if semver.Major(ver) == major {
				return ver, nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return "", fmt.Errorf("failed to find a version matching the provided major version %q", major)
}
