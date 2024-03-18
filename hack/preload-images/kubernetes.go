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

package main

import (
	"context"
	"strings"

	"github.com/google/go-github/v60/github"

	"k8s.io/klog/v2"
)

// recentK8sVersions returns the most recent k8s version, usually around 100.
func recentK8sVersions() ([]string, error) {
	const k8s = "kubernetes"
	client := github.NewClient(nil)
	list, _, err := client.Repositories.ListReleases(context.Background(), k8s, k8s, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, err
	}
	var releases []string
	for _, r := range list {
		// Exclude "alpha" releases.
		if strings.Contains(r.GetTagName(), "alpha") {
			continue
		}
		releases = append(releases, r.GetTagName())
	}
	klog.InfoS("Got releases", "releases", releases)
	return releases, nil
}
