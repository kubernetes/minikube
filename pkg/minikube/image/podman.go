/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package image

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// PodmanWrite saves the image into podman as the given tag.
// same as github.com/google/go-containerregistry/pkg/v1/daemon
func PodmanWrite(ref name.Reference, img v1.Image, opts ...tarball.WriteOption) (string, error) {
	pr, pw := io.Pipe()
	go func() {
		_ = pw.CloseWithError(tarball.Write(ref, img, pw, opts...))
	}()

	// write the image in docker save format first, then load it
	cmd := exec.Command("sudo", "podman", "image", "load")
	cmd.Stdin = pr
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error loading image: %v", err)
	}
	// pull the image from the registry, to get the digest too
	// podman: "Docker references with both a tag and digest are currently not supported"
	cmd = exec.Command("sudo", "podman", "image", "pull", strings.Split(ref.Name(), "@")[0])
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error pulling image: %v", err)
	}
	return string(output), nil
}
