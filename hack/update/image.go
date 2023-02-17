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

package update

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetImageSHA gets the SHA of the provided image
func GetImageSHA(image string) (string, error) {
	if o, err := exec.Command("docker", "pull", image).CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to pull image: %v: %s", err, o)
	}
	o, err := exec.Command("docker", "inspect", "--format={{index .RepoDigests 0}}", image).Output()
	if err != nil {
		return "", fmt.Errorf("failed to inspect image: %v: %s", err, o)
	}
	digest := strings.TrimSpace(string(o))
	if !strings.Contains(digest, "@") {
		return "", fmt.Errorf("digest doesn't contain a SHA: %s", digest)
	}
	return strings.Split(digest, "@")[1], nil
}
