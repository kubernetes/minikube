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

package detect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	dockerHubAuthURL      = "https://auth.docker.io/token"
	dockerHubRegistryURL  = "https://registry-1.docker.io"
	dockerHubPreviewRepo  = "ratelimitpreview/test"
	dockerHubPreviewScope = "repository:ratelimitpreview/test:pull"
)

// DockerHubRateLimitRemaining returns the remaining Docker Hub pulls for the calling IP.
// Based on https://docs.docker.com/docker-hub/usage/pulls/#view-pull-rate-and-limit
// calling this func will NOT reduce number of remaining pulls.
func DockerHubRateLimitRemaining(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token, err := dockerHubRateLimitToken(ctx)
	if err != nil {
		return 0, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/latest", dockerHubRegistryURL, dockerHubPreviewRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, manifestURL, nil)
	if err != nil {
		return 0, fmt.Errorf("building rate limit request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("querying rate limit: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, fmt.Errorf("docker hub rate limit exceeded (HTTP 429)")
	}

	// Go canonicalizes header keys, so both ratelimit-remaining and RateLimit-Remaining work.
	remainingHeader := resp.Header.Get("RateLimit-Remaining")
	if remainingHeader == "" {
		remainingHeader = resp.Header.Get("ratelimit-remaining")
	}
	if remainingHeader == "" {
		return 0, errors.New("docker hub RateLimit-Remaining header missing")
	}

	remaining, err := parseDockerHubRemaining(remainingHeader)
	if err != nil {
		return 0, err
	}

	return remaining, nil
}

// dockerHubRateLimitToken retrieves an authentication token from Docker Hub's authentication service.
// for non-logged in dockers it will still return a token but with lower rate limits.
func dockerHubRateLimitToken(ctx context.Context) (string, error) {
	values := url.Values{}
	values.Set("service", "registry.docker.io")
	values.Set("scope", dockerHubPreviewScope)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dockerHubAuthURL, nil)
	if err != nil {
		return "", fmt.Errorf("building auth request: %w", err)
	}
	req.URL.RawQuery = values.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth request returned %d", resp.StatusCode)
	}

	var parsed struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decoding auth response: %w", err)
	}
	if parsed.Token == "" {
		return "", errors.New("empty token from Docker Hub auth response")
	}
	return parsed.Token, nil
}

// parseDockerHubRemaining extracts the remaining API rate limit count from a Docker Hub
// RateLimit-Remaining header value.
func parseDockerHubRemaining(headerVal string) (int, error) {
	separators := func(r rune) bool {
		return r == ';' || r == ','
	}
	for _, part := range strings.FieldsFunc(headerVal, separators) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if remaining, err := strconv.Atoi(part); err == nil {
			return remaining, nil
		}
	}
	return 0, fmt.Errorf("unable to find integer value in RateLimit-Remaining header %q", headerVal)
}
