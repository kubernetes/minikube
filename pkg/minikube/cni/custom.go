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

package cni

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

// Custom is a CNI manager than applies a user-specified manifest
type Custom struct {
	cc       config.ClusterConfig
	manifest string
}

// String returns a string representation of this CNI
func (c Custom) String() string {
	return c.manifest
}

// NewCustom returns a well-formed Custom CNI manager
func NewCustom(cc config.ClusterConfig, manifest string) (Custom, error) {
	_, err := os.Stat(manifest)
	if err != nil {
		return Custom{}, errors.Wrap(err, "stat")
	}

	return Custom{
		cc:       cc,
		manifest: manifest,
	}, nil
}

// Apply enables the CNI
func (c Custom) Apply(r Runner) error {
	if err := c.removeCNIConfig(r); err != nil {
		return errors.Wrap(err, "remove cni config")
	}
	m, err := assets.NewFileAsset(c.manifest, path.Dir(manifestPath()), path.Base(manifestPath()), "0644")
	if err != nil {
		return errors.Wrap(err, "manifest")
	}

	return applyManifest(c.cc, r, m)
}

// CIDR returns the default CIDR used by this CNI
func (c Custom) CIDR() string {
	return DefaultPodCIDR
}

func (c Custom) removeCNIConfig(r Runner) error {
	klog.Infof("Removing existing cni configuration (/etc/cni/net.d/*)")
	rm := fmt.Sprintf("rm -rf /etc/cni/net.d/*")
	if _, err := r.RunCmd(exec.Command("sudo", "/bin/bash", "-c", rm)); err != nil {
		klog.Errorf("netconf update failed: %v", err)
		return err
	}
	return sysinit.New(r).Restart(c.cc.KubernetesConfig.ContainerRuntime)

}
