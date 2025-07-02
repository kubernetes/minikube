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

// Package bsutil package will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/drivers/kic/oci"
)

// getTargetArchitecture returns the target architecture for binary downloads.
// For remote Docker contexts, it queries the remote daemon's architecture.
// For local contexts, it uses the local machine's architecture.
func getTargetArchitecture() string {
	// Check if we're using a remote Docker context
	if oci.IsRemoteDockerContext() {
		klog.Infof("Detected remote Docker context, querying remote daemon architecture")

		// Get Docker daemon architecture directly
		dockerArch, err := getDockerArchitectureDirect()
		if err != nil {
			klog.Warningf("Failed to get Docker architecture directly, falling back to local architecture: %v", err)
			return runtime.GOARCH
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

	return runtime.GOARCH
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

// TransferBinaries transfers all required Kubernetes binaries
func TransferBinaries(cfg config.KubernetesConfig, c command.Runner, sm sysinit.Manager, binariesURL string) error {
	// Get the target architecture (local or remote)
	targetArch := getTargetArchitecture()
	klog.Infof("Target architecture for binaries: %s", targetArch)

	// Check if binaries exist and are the correct architecture
	ok, err := binariesExist(cfg, c)
	if err == nil && ok && !oci.IsRemoteDockerContext() {
		// For local contexts, trust existing binaries
		klog.Info("Found k8s binaries for local context, skipping transfer")
		return nil
	}

	if err == nil && ok && oci.IsRemoteDockerContext() {
		// For remote contexts, verify architecture of existing binaries
		klog.Info("Found k8s binaries, but using remote context - verifying architecture...")
		needRedownload := false

		// Check if we can run kubelet --version to verify it's the right architecture
		dir := binRoot(cfg.KubernetesVersion)
		kubeletPath := path.Join(dir, "kubelet")
		rr, err := c.RunCmd(exec.Command("sudo", kubeletPath, "--version"))
		if err != nil {
			klog.Warningf("Existing kubelet binary failed to run (likely wrong architecture): %v", err)
			needRedownload = true
		} else {
			klog.Infof("Existing kubelet works: %s", rr.Stdout.String())
		}

		if !needRedownload {
			return nil
		}
		klog.Info("Need to re-download binaries for correct architecture")
	}

	klog.Infof("Didn't find k8s binaries or need correct architecture: %v\nInitiating transfer...", err)

	dir := binRoot(cfg.KubernetesVersion)
	_, err = c.RunCmd(exec.Command("sudo", "mkdir", "-p", dir))
	if err != nil {
		return err
	}

	klog.Infof("Transferring binaries for architecture: %s", targetArch)

	var g errgroup.Group
	for _, name := range constants.KubernetesReleaseBinaries {
		name := name
		g.Go(func() error {
			src, err := download.Binary(name, cfg.KubernetesVersion, "linux", targetArch, binariesURL)
			if err != nil {
				return errors.Wrapf(err, "downloading %s", name)
			}

			if name == "kubelet" && sm.Active(name) {
				if err := sm.ForceStop(name); err != nil {
					klog.Errorf("unable to stop kubelet: %v", err)
				}
			}

			dst := path.Join(dir, name)
			if err := machine.CopyBinary(c, src, dst); err != nil {
				return errors.Wrapf(err, "copybinary %s -> %s", src, dst)
			}
			return nil
		})
	}
	return g.Wait()
}

// binariesExist returns true if the binaries already exist
func binariesExist(cfg config.KubernetesConfig, c command.Runner) (bool, error) {
	dir := binRoot(cfg.KubernetesVersion)
	rr, err := c.RunCmd(exec.Command("sudo", "ls", dir))
	if err != nil {
		return false, err
	}
	stdout := rr.Stdout.String()
	foundBinaries := map[string]struct{}{}
	for _, binary := range strings.Split(stdout, "\n") {
		foundBinaries[binary] = struct{}{}
	}
	for _, name := range constants.KubernetesReleaseBinaries {
		if _, ok := foundBinaries[name]; !ok {
			return false, fmt.Errorf("didn't find preexisting %s", name)
		}
	}
	return true, nil
}

// binRoot returns the persistent path binaries are stored in
func binRoot(version string) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", version)
}
