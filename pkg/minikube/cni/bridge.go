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
	"fmt"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
)

// bridge is what minikube defaulted to when `--enable-default-cni=true`
// https://github.com/containernetworking/plugins/blob/master/plugins/main/bridge/README.md

var bridgeConf = template.Must(template.New("bridge").Parse(`
{
  "cniVersion": "0.3.1",
  "name": "bridge",
  "type": "bridge",
  "bridge": "bridge",
  "addIf": "true",
  "isDefaultGateway": true,
  "forceAddress": false,
  "ipMasq": true,
  "hairpinMode": true,
  "ipam": {
      "type": "host-local",
      "subnet": "{{.PodCIDR}}"
  }
}
`))

// Bridge is a simple CNI manager for single-node usage
type Bridge struct {
	cc config.ClusterConfig
}

// String returns a string representation of this CNI
func (c Bridge) String() string {
	return "bridge CNI"
}

func (c Bridge) netconf() (assets.CopyableFile, error) {
	input := &tmplInput{PodCIDR: DefaultPodCIDR}

	b := bytes.Buffer{}
	if err := bridgeConf.Execute(&b, input); err != nil {
		return nil, err
	}

	return assets.NewMemoryAssetTarget(b.Bytes(), "/etc/cni/net.d/1-k8s.conf", "0644"), nil
}

// Apply enables the CNI
func (c Bridge) Apply(r Runner) error {
	if len(c.cc.Nodes) > 1 {
		return fmt.Errorf("bridge CNI is incompatible with multi-node clusters")
	}

	f, err := c.netconf()
	if err != nil {
		return errors.Wrap(err, "netconf")
	}

	if err := r.Copy(f); err != nil {
		return errors.Wrapf(err, "copy")
	}

	return nil
}

// CIDR returns the default CIDR used by this CNI
func (c Bridge) CIDR() string {
	return DefaultPodCIDR
}

// Images returns the list of images used by this CNI
func (c Bridge) Images() []string {
	return []string{}
}
