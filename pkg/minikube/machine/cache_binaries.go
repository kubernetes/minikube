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

package machine

import (
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/drivers/kic/oci"
)

// isExcluded returns whether `binary` is expected to be excluded, based on `excludedBinaries`.
func isExcluded(binary string, excludedBinaries []string) bool {
	if excludedBinaries == nil {
		return false
	}
	for _, excludedBinary := range excludedBinaries {
		if binary == excludedBinary {
			return true
		}
	}
	return false
}

// getTargetArchitecture returns the target architecture for binary downloads.
// For remote Docker contexts, it queries the remote daemon's architecture.
// For local contexts, it uses the local machine's architecture.
func getTargetArchitecture() string {
	// Check if we're using a remote Docker context
	if oci.IsRemoteDockerContext() {
		klog.Infof("Detected remote Docker context, querying remote daemon architecture")

		// Directly get the architecture from Docker daemon
		dockerArch := getDockerArchitecture()
		if dockerArch != "" {
			klog.Infof("Using remote Docker daemon architecture: %s", dockerArch)
			return dockerArch
		}

		klog.Warningf("Could not determine remote architecture, falling back to local architecture")
	}

	return runtime.GOARCH
}

// getDockerArchitecture queries Docker daemon for its architecture and converts to Go format
func getDockerArchitecture() string {
	// Use direct Docker call to get the architecture
	dockerArch, err := getDockerArchitectureDirect()
	if err != nil {
		klog.Warningf("Failed to get Docker architecture directly: %v", err)
		return ""
	}

	// Convert Docker architecture names to Go architecture names
	switch dockerArch {
	case "x86_64":
		return "amd64"
	case "aarch64", "arm64":
		return "arm64"
	case "armv7l":
		return "arm"
	default:
		klog.Warningf("Unknown Docker architecture %q, falling back to local architecture", dockerArch)
		return runtime.GOARCH
	}
}

// getDockerArchitectureDirect directly queries Docker for architecture info
func getDockerArchitectureDirect() (string, error) {
	cmd := exec.Command("docker", "system", "info", "--format", "{{.Architecture}}")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "running docker system info")
	}

	arch := strings.TrimSpace(string(output))
	klog.Infof("Docker daemon architecture: %s", arch)
	return arch, nil
}

// CacheBinariesForBootstrapper will cache binaries for a bootstrapper
func CacheBinariesForBootstrapper(version string, excludeBinaries []string, binariesURL string) error {
	binaries := bootstrapper.GetCachedBinaryList()

	// Get the target architecture (local or remote)
	targetArch := getTargetArchitecture()
	klog.Infof("Caching binaries for architecture: %s", targetArch)

	var g errgroup.Group
	for _, bin := range binaries {
		if isExcluded(bin, excludeBinaries) {
			continue
		}
		bin := bin // https://go.dev/doc/faq#closures_and_goroutines
		g.Go(func() error {
			if _, err := download.Binary(bin, version, "linux", targetArch, binariesURL); err != nil {
				return errors.Wrapf(err, "caching binary %s", bin)
			}
			return nil
		})
	}
	return g.Wait()
}

// CopyBinary copies a locally cached binary to the guest VM
func CopyBinary(cr command.Runner, src string, dest string) error {
	f, err := assets.NewFileAsset(src, path.Dir(dest), path.Base(dest), "0755")
	if err != nil {
		return errors.Wrap(err, "new file asset")
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Warningf("error closing the file %s: %v", f.GetSourcePath(), err)
		}
	}()

	if err := cr.Copy(f); err != nil {
		return errors.Wrapf(err, "copy")
	}
	return nil
}
