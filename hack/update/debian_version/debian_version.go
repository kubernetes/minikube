/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"regexp"
	"strings"

	"k8s.io/klog/v2"

	"k8s.io/minikube/hack/update"
)

var (
	schema = map[string]update.Item{
		"deploy/kicbase/Dockerfile": {
			Replace: map[string]string{
				`KICBASE_IMAGE=.*`: `KICBASE_IMAGE="{{.LatestVersion}}"`,
			},
		},
	}
)

// Data holds the latest dated Debian slim tag.
type Data struct {
	LatestVersion string
}

// dateTag matches the current Debian slim date-stamped tags
// (for example, trixie-20260824-slim).
var dateTag = regexp.MustCompile(`^trixie-\d{8}-slim$`)

// latestSlimTag returns the newest dated slim tag so the resulting
// image digest stays stable between updater runs.
func latestSlimTag(tags []string) (string, error) {
	var newestDateTag string
	for _, tag := range tags {
		// Skip anything that isn't a trixie slim tag to avoid matching other
		// Debian variants.
		if !strings.HasPrefix(tag, "trixie-") || !strings.HasSuffix(tag, "-slim") {
			continue
		}

		// Track the lexicographically greatest dated tag, which corresponds to
		// the most recent date stamp provided by Debian.
		if dateTag.MatchString(tag) {
			if newestDateTag == "" || tag > newestDateTag {
				newestDateTag = tag
			}
		}
	}

	if newestDateTag != "" {
		return newestDateTag, nil
	}
	return "", fmt.Errorf("no dated tag found that matches: %s", dateTag.String())
}

func main() {
	tags, err := update.ImageTagsFromDockerHub("library/debian")
	if err != nil {
		klog.Fatal(err)
	}
	tag, err := latestSlimTag(tags)
	if err != nil {
		klog.Fatal(err)
	}
	data := Data{LatestVersion: fmt.Sprintf("debian:%s", tag)}

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}
}
