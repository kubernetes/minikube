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

// Data holds latest Ubuntu jammy version in semver format.
type Data struct {
	LatestVersion string
}

// bookwormDateTag matches Debian bookworm slim tags that include an 8-digit
// date stamp (for example, bookworm-20250929-slim).
var bookwormDateTag = regexp.MustCompile(`^bookworm-\d{8}-slim$`)

// latestBookwormSlimTag returns the newest bookworm slim tag that includes a
// date suffix. The updater now requires a dated tag to be present so that the
// resulting image digest remains stable and predictable between runs.
func latestBookwormSlimTag(tags []string) (string, error) {
	var newestDateTag string
	for _, tag := range tags {
		// Skip anything that isn't a bookworm slim tag to avoid matching other
		// Debian variants.
		if !strings.HasPrefix(tag, "bookworm-") || !strings.HasSuffix(tag, "-slim") {
			continue
		}

		// Track the lexicographically greatest dated tag, which corresponds to
		// the most recent date stamp provided by Debian.
		if bookwormDateTag.MatchString(tag) {
			if newestDateTag == "" || tag > newestDateTag {
				newestDateTag = tag
			}
		}
	}

	if newestDateTag != "" {
		return newestDateTag, nil
	}
	return "", fmt.Errorf("no dated tag found that matches: %s", bookwormDateTag.String())
}

func main() {
	tags, err := update.ImageTagsFromDockerHub("library/debian")
	if err != nil {
		klog.Fatal(err)
	}
	tag, err := latestBookwormSlimTag(tags)
	if err != nil {
		klog.Fatal(err)
	}
	data := Data{LatestVersion: fmt.Sprintf("debian:%s", tag)}

	if err := update.Apply(schema, data); err != nil {
		klog.Fatalf("unable to apply update: %v", err)
	}
}
