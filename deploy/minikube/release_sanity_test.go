/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"slices"
	"testing"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/util"
)

// TestReleasesJSON checks if all *GA* releases
//
//	enlisted in https://storage.googleapis.com/minikube/releases-v2.json
//	are available to download and have correct hashsum
func TestReleasesJSON(t *testing.T) {
	releases, err := notify.AllVersionsFromURL(notify.GithubMinikubeReleasesURL)
	if err != nil {
		t.Fatalf("Error getting releases.json: %v", err)
	}
	checkReleases(t, releases)
}

// TestBetaReleasesJSON checks if all *BETA* releases
//
//	enlisted in https://storage.googleapis.com/minikube/releases-beta-v2.json
//	are available to download and have correct hashsum
func TestBetaReleasesJSON(t *testing.T) {
	releases, err := notify.AllVersionsFromURL(notify.GithubMinikubeBetaReleasesURL)
	if err != nil {
		t.Fatalf("Error getting releases-bets.json: %v", err)
	}
	checkReleases(t, releases)
}

type binary struct {
	OS   string
	Arch string
	SHA  string
}

// validateChecksums verifies that the v1 flat fields and v2 nested amd64
// fields are identical.
func validateChecksums(r notify.Release) error {
	c := r.Checksums
	if c.AMD64 == nil {
		return nil
	}
	if c.Darwin != c.AMD64.Darwin {
		return fmt.Errorf("darwin amd64 checksum mismatch: v1=%s v2=%s", c.Darwin, c.AMD64.Darwin)
	}
	if c.Linux != c.AMD64.Linux {
		return fmt.Errorf("linux amd64 checksum mismatch: v1=%s v2=%s", c.Linux, c.AMD64.Linux)
	}
	if c.Windows != c.AMD64.Windows {
		return fmt.Errorf("windows amd64 checksum mismatch: v1=%s v2=%s", c.Windows, c.AMD64.Windows)
	}
	return nil
}

// releaseBinaries returns a deduplicated list of binaries with expected
// checksums for a release. The releases JSON contains entries in two formats:
//   - v1 (older releases): flat "darwin"/"linux"/"windows" fields, always amd64.
//   - v2 (newer releases): nested arch objects ("amd64", "arm64", …) plus the
//     same flat fields duplicated for backward compatibility.
//
// Both are normalized into a single list so each binary is checked exactly once.
func releaseBinaries(r notify.Release) []binary {
	c := r.Checksums
	var bins []binary

	if c.AMD64 != nil {
		// v2: nested arch objects.
		if c.AMD64.Darwin != "" {
			bins = append(bins, binary{"darwin", "amd64", c.AMD64.Darwin})
		}
		if c.AMD64.Linux != "" {
			bins = append(bins, binary{"linux", "amd64", c.AMD64.Linux})
		}
		if c.AMD64.Windows != "" {
			bins = append(bins, binary{"windows", "amd64", c.AMD64.Windows})
		}
	} else {
		// v1: flat fields, always amd64.
		if c.Darwin != "" {
			bins = append(bins, binary{"darwin", "amd64", c.Darwin})
		}
		if c.Linux != "" {
			bins = append(bins, binary{"linux", "amd64", c.Linux})
		}
		if c.Windows != "" {
			bins = append(bins, binary{"windows", "amd64", c.Windows})
		}
	}
	if c.ARM != nil && c.ARM.Linux != "" {
		bins = append(bins, binary{"linux", "arm", c.ARM.Linux})
	}
	if c.ARM64 != nil {
		if c.ARM64.Darwin != "" {
			bins = append(bins, binary{"darwin", "arm64", c.ARM64.Darwin})
		}
		if c.ARM64.Linux != "" {
			bins = append(bins, binary{"linux", "arm64", c.ARM64.Linux})
		}
	}
	if c.PPC64LE != nil && c.PPC64LE.Linux != "" {
		bins = append(bins, binary{"linux", "ppc64le", c.PPC64LE.Linux})
	}
	if c.S390X != nil && c.S390X.Linux != "" {
		bins = append(bins, binary{"linux", "s390x", c.S390X.Linux})
	}

	slices.SortFunc(bins, func(a, b binary) int {
		if c := cmp.Compare(a.OS, b.OS); c != 0 {
			return c
		}
		return cmp.Compare(a.Arch, b.Arch)
	})
	return bins
}

func checkReleases(t *testing.T, rs notify.Releases) {
	for _, r := range rs.Releases {
		fmt.Printf("Checking release: %s\n", r.Name)
		if err := validateChecksums(r); err != nil {
			t.Errorf("release %s: %v", r.Name, err)
			continue
		}
		for _, bin := range releaseBinaries(r) {
			fmt.Printf("Checking SHA for %s-%s.\n", bin.OS, bin.Arch)
			actualSha, err := getSHAFromURL(util.GetBinaryDownloadURL(r.Name, bin.OS, bin.Arch))
			if err != nil {
				t.Errorf("Error calculating SHA for %s-%s-%s. Error: %v", r.Name, bin.OS, bin.Arch, err)
				continue
			}
			if actualSha != bin.SHA {
				t.Errorf("ERROR: SHA does not match for version %s, os %s, arch %s. Expected %s, got %s.", r.Name, bin.OS, bin.Arch, bin.SHA, actualSha)
				continue
			}
		}
	}
}

func getSHAFromURL(url string) (string, error) {
	fmt.Println("Downloading: ", url)
	r, err := retryablehttp.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET %s: %w", url, err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", url, r.Status)
	}
	h := sha256.New()
	if _, err := io.Copy(h, r.Body); err != nil {
		return "", fmt.Errorf("GET %s: failed to copy body: %w", url, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
