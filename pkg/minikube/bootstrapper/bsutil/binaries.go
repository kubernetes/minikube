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
	"os/exec"
	"path"
	"runtime"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// TransferBinaries transfers all required Kubernetes binaries
func TransferBinaries(cfg config.KubernetesConfig, c command.Runner) error {
	dir := binRoot(cfg.KubernetesVersion)
	_, err := c.RunCmd(exec.Command("sudo", "mkdir", "-p", dir))
	if err != nil {
		return err
	}

	var g errgroup.Group
	for _, name := range constants.KubernetesReleaseBinaries {
		name := name
		g.Go(func() error {
			src, err := machine.CacheBinary(name, cfg.KubernetesVersion, "linux", runtime.GOARCH)
			if err != nil {
				return errors.Wrapf(err, "downloading %s", name)
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

// binRoot returns the persistent path binaries are stored in
func binRoot(version string) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", version)
}
