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
	"os"
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
	if err := c.addFakeConfig(r); err != nil {
		return errors.Wrap(err, "create fake cni config")
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

// addFakeConfig adds a fake CNI configuration file to the node to prevent
// podman CNI to serve ip addresses to pods in case 3rd party CNI is enabled
func (c Custom) addFakeConfig(r Runner) error {
	klog.Infof("creating fake CNI config")
	fakeCNI := `{
  "cniVersion": "0.4.0",
  "name": "fake",
  "type": "host-local"
}`

	b := []byte(fakeCNI)
	fakeCNIAsset := assets.NewMemoryAssetTarget(b, "/etc/cni/net.d/86-fake.conf", "0644")
	if err := r.Copy(fakeCNIAsset); err != nil {
		return errors.Wrapf(err, "copy")
	}

	// crio service must be restarted otherwise the podman cni is still used
	return sysinit.New(r).Restart(c.cc.KubernetesConfig.ContainerRuntime)
}
