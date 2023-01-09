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
	"bytes"
	_ "embed"
	"os/exec"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
)

// ref: https://github.com/flannel-io/flannel#deploying-flannel-manually: "For Kubernetes v1.17+"; multi-arch support

//go:embed flannel.yaml
var flannelYaml string

// https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml
var flannelTmpl = template.Must(template.New("flannel").Parse(flannelYaml))

type flannelTmplStruct struct {
	PodCIDR string
}

// Flannel is the Flannel CNI manager
type Flannel struct {
	cc config.ClusterConfig
}

// String returns a string representation of this CNI
func (c Flannel) String() string {
	return "Flannel"
}

// Apply enables the CNI
func (c Flannel) Apply(r Runner) error {
	// Mostly applicable to the 'none' driver
	_, err := r.RunCmd(exec.Command("stat", "/opt/cni/bin/portmap"))
	if err != nil {
		return errors.Wrap(err, "required 'portmap' CNI plug-in not found")
	}

	m, err := c.manifest()
	if err != nil {
		return errors.Wrap(err, "manifest")
	}
	return applyManifest(c.cc, r, m)
}

// manifest returns a Kubernetes manifest for a CNI
func (c Flannel) manifest() (assets.CopyableFile, error) {
	input := &flannelTmplStruct{
		PodCIDR: DefaultPodCIDR,
	}
	b := bytes.Buffer{}
	if err := flannelTmpl.Execute(&b, input); err != nil {
		return nil, err
	}
	return manifestAsset(b.Bytes()), nil
}

// CIDR returns the default CIDR used by this CNI
func (c Flannel) CIDR() string {
	return DefaultPodCIDR
}
