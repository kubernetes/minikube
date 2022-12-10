/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"os"
	"os/exec"
	"strings"

	"k8s.io/minikube/pkg/minikube/image"
)

// ToDriverCache
// calls OCIBIN's load command at specified path:
// loads the archived container image at provided PATH.
func ArchiveToDriverCache(ociBin, path string) error {
	cmd := exec.Command(ociBin, "load", "-i", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// IsInCache
// searches in OCIBIN's cache for the IMG; returns true if found. no error handling
func IsImageInCache(ociBin, img string) bool {
	res, err := runCmd(exec.Command(ociBin, "images", "--format", "{{.Repository}}:{{.Tag}}@{{.Digest}}"))
	if err != nil {
		// only the docker binary seems to have this issue..
		// the docker.io/ substring is cut from the output and formatting doesn't help
		if ociBin == Docker {
			img = image.TrimDockerIO(img)
		}

		if strings.Contains(res.Stdout.String(), img) {
			return true
		}
	}

	return false
}
