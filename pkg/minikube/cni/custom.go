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
	"context"

	"k8s.io/minikube/pkg/minikube/assets"
)

// defaultCNIConfig is an default CNI config
const defaultCNIConfig = `
{
  "cniVersion": "0.3.0",
  "name": "rkt.kubernetes.io",
  "type": "bridge",
  "bridge": "mybridge",
  "mtu": 1460,
  "addIf": "true",
  "isGateway": true,
  "ipMasq": true,
  "ipam": {
    "type": "host-local",
    "subnet": "10.1.0.0/16",
    "gateway": "10.1.0.1",
    "routes": [
      {
        "dst": "0.0.0.0/0"
      }
    ]
  }
}
`

// Custom is a CNI manager than does nothing
type Custom struct{}

// Assets returns a list of assets necessary to enable this CNI
func (n Custom) Assets() ([]assets.CopyableFile, error) {
	return []assets.CopyableFile{assets.NewMemoryAssetTarget([]byte(defaultCNIConfig), "/etc/cni/net.d/1-k8s.conf", "0644")}, nil
}

// Apply enables the CNI
func (n Custom) Apply(context.Context, Runner) error {
	return nil
}

// CIDR returns the default CIDR used by this CNI
func (n Custom) CIDR() string {
	return ""
}
