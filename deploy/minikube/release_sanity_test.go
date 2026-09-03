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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
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
	checkReleasesV2(t, releases)
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
	checkReleasesV2(t, releases)
}

func checkReleasesV1(t *testing.T, r notify.Release) {
	checksums := map[string]string{
		"darwin":  r.Checksums.Darwin,
		"linux":   r.Checksums.Linux,
		"windows": r.Checksums.Windows,
	}
	for platform, sha := range checksums {
		if sha == "" {
			continue
		}
		fmt.Printf("Checking SHA for %s.\n", platform)
		actualSha, err := getSHAFromURL(util.GetBinaryDownloadURL(r.Name, platform, "amd64"))
		if err != nil {
			t.Errorf("Error calculating SHA for %s-%s. Error: %v", r.Name, platform, err)
			continue
		}
		if actualSha != sha {
			t.Errorf("ERROR: SHA does not match for version %s, platform %s. Expected %s, got %s.", r.Name, platform, sha, actualSha)
			continue
		}
	}
}

func getSHAMap(r notify.Release) map[string]map[string]string {
	c := r.Checksums
	m := map[string]map[string]string{
		"darwin":  {},
		"linux":   {},
		"windows": {},
	}
	if c.AMD64 != nil {
		if c.AMD64.Darwin != "" {
			m["darwin"]["amd64"] = c.AMD64.Darwin
		}
		if c.AMD64.Linux != "" {
			m["linux"]["amd64"] = c.AMD64.Linux
		}
		if c.AMD64.Windows != "" {
			m["windows"]["amd64"] = c.AMD64.Windows
		}
	}
	if c.ARM != nil && c.ARM.Linux != "" {
		m["linux"]["arm"] = c.ARM.Linux
	}
	if c.ARM64 != nil {
		if c.ARM64.Darwin != "" {
			m["darwin"]["arm64"] = c.ARM64.Darwin
		}
		if c.ARM64.Linux != "" {
			m["linux"]["arm64"] = c.ARM64.Linux
		}
	}
	if c.PPC64LE != nil && c.PPC64LE.Linux != "" {
		m["linux"]["ppc64le"] = c.PPC64LE.Linux
	}
	if c.S390X != nil && c.S390X.Linux != "" {
		m["linux"]["s390x"] = c.S390X.Linux
	}
	return m
}

func checkReleasesV2(t *testing.T, rs notify.Releases) {
	for _, r := range rs.Releases {
		fmt.Printf("Checking release: %s\n", r.Name)
		checkReleasesV1(t, r)
		release := getSHAMap(r)
		for os, archs := range release {
			for arch, sha := range archs {
				fmt.Printf("Checking SHA for %s-%s.\n", os, arch)
				actualSha, err := getSHAFromURL(util.GetBinaryDownloadURL(r.Name, os, arch))
				if err != nil {
					t.Errorf("Error calculating SHA for %s-%s-%s. Error: %v", r.Name, os, arch, err)
					continue
				}
				if actualSha != sha {
					t.Errorf("ERROR: SHA does not match for version %s, os %s, arch %s. Expected %s, got %s.", r.Name, os, arch, sha, actualSha)
					continue
				}
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
