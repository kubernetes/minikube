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

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// TransferBinaries transfers all required Kubernetes binaries
func TransferBinaries(cfg config.KubernetesConfig, c command.Runner, sm sysinit.Manager, binariesURL string) error {
	// Skip binary transfer in --no-kubernetes mode
	if viper.GetBool("no-kubernetes") {
		klog.Info("Skipping Kubernetes binary transfer due to --no-kubernetes flag")
		return nil
	}

	ok, err := binariesExist(cfg, c)
	if err == nil && ok {
		klog.Info("Found k8s binaries, skipping transfer")
		return nil
	}
	klog.Infof("Didn't find k8s binaries: %v\nInitiating transfer...", err)

	dir := binRoot(cfg.KubernetesVersion)
	_, err = c.RunCmd(exec.Command("sudo", "mkdir", "-p", dir))
	if err != nil {
		return err
	}

	var g errgroup.Group
	for _, name := range constants.KubernetesReleaseBinaries {
		name := name
		g.Go(func() error {
			src, err := download.Binary(name, cfg.KubernetesVersion, "linux", runtime.GOARCH, binariesURL)
			if err != nil {
				return fmt.Errorf("downloading %s: %w", name, err)
			}

			if name == "kubelet" && sm.Active(name) {
				if err := sm.ForceStop(name); err != nil {
					klog.Errorf("unable to stop kubelet: %v", err)
				}
			}

			dst := path.Join(dir, name)
			if err := copyBinary(c, src, dst); err != nil {
				return fmt.Errorf("copybinary %s -> %s: %w", src, dst, err)
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

// copyBinary copies a locally cached binary to the guest VM
func copyBinary(cr command.Runner, src, dest string) error {
	f, err := assets.NewFileAsset(src, path.Dir(dest), path.Base(dest), "0755")
	if err != nil {
		return fmt.Errorf("new file asset: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Warningf("error closing the file %s: %v", f.GetSourcePath(), err)
		}
	}()

	if err := cr.Copy(f); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}
