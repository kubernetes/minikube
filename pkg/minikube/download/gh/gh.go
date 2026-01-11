/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

// Package gh provides helper utilities for interacting with the GitHub API
package gh

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

// ReleaseAssets retrieves a GitHub release by tag from org/project.
// Try to not call this too often. preferably cache and reuse. to avoid rate limits.
func ReleaseAssets(org, project, tag string) ([]*github.ReleaseAsset, error) {
	ctx := context.Background()
	// Use an authenticated client when GITHUB_TOKEN is set to avoid low rate limits.
	httpClient := oauthClient(ctx, os.Getenv("GITHUB_TOKEN"))
	ghc := github.NewClient(httpClient)

	rel, _, err := ghc.Repositories.GetReleaseByTag(ctx, org, project, tag)
	if err != nil {
		return nil, err
	}
	return rel.Assets, nil
}

// AssetSHA256 returns the  SHA-256 digest for the asset with the given name
// from the provided release assets from github API.
// to avoid rate limits. encouraged to call pass results of ReleaseAssets here.
func AssetSHA256(assetName string, assets []*github.ReleaseAsset) ([]byte, error) {
	for _, asset := range assets {
		if asset.GetName() != assetName {
			continue
		}
		d := asset.GetDigest() // e.g. "sha256:fdcb..."
		if d == "" {
			return []byte(""), fmt.Errorf("asset %q has no digest; id=%d url=%s", assetName, asset.GetID(), asset.GetBrowserDownloadURL())
		}
		const prefix = "sha256:"
		d = strings.TrimPrefix(d, prefix)
		return []byte(d), nil
	}
	return []byte(""), fmt.Errorf("asset %q not found", assetName)
}

func oauthClient(ctx context.Context, token string) *http.Client {
	if token == "" {
		return nil // unauthenticated client (lower rate limit)
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, ts)
}
