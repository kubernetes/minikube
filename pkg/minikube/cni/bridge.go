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
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
)

// bridge is what minikube defaulted to when `--enable-default-cni=true`
// ref: https://www.cni.dev/plugins/current/main/bridge/
// ref: https://www.cni.dev/plugins/current/meta/portmap/
// ref: https://www.cni.dev/plugins/current/meta/firewall/


// renderBridgeConflist builds a bridge CNI config that supports IPv4-only, IPv6-only, or dual-stack.
func renderBridgeConflist(k8s config.KubernetesConfig) ([]byte, error) {
       // minimal structs for JSON marshal
       type rng struct{ Subnet string `json:"subnet"` }
       type ipam struct {
               Type   string    `json:"type"`
               Subnet string    `json:"subnet,omitempty"` // single-stack (v4 or v6)
               Ranges [][]rng   `json:"ranges,omitempty"` // dual-stack
       }
       type bridge struct {
               Type             string `json:"type"`
               Bridge           string `json:"bridge"`
               IsDefaultGateway bool   `json:"isDefaultGateway"`
               HairpinMode      bool   `json:"hairpinMode"`
               IPMasq           bool   `json:"ipMasq"`
               IPAM             ipam   `json:"ipam"`
       }
       type plugin struct {
               Type         string          `json:"type"`
               Capabilities map[string]bool `json:"capabilities,omitempty"`
       }
       type conflist struct {
               CNIVersion string        `json:"cniVersion"`
               Name       string        `json:"name"`
               Plugins    []interface{} `json:"plugins"`
       }

       v4 := k8s.PodCIDR   != ""
       v6 := k8s.PodCIDRv6 != ""

       cfgIPAM := ipam{Type: "host-local"}
       switch {
       case v4 && v6:
               cfgIPAM.Ranges = [][]rng{{{Subnet: k8s.PodCIDR}}, {{Subnet: k8s.PodCIDRv6}}}
       case v6:
               cfgIPAM.Subnet = k8s.PodCIDRv6
       default:
               // fall back to previous default if unset upstream
               cidr := k8s.PodCIDR
               if cidr == "" {
                       cidr = DefaultPodCIDR
               }
               cfgIPAM.Subnet = cidr
       }

       // NAT generally not desired for IPv6; keep masquerade only for v4
       ipMasq := v4 && !v6

       br := bridge{
               Type: "bridge", Bridge: "cni0",
               IsDefaultGateway: true,
               HairpinMode: true,
               IPMasq: ipMasq,
               IPAM: cfgIPAM,
       }
       portmap := plugin{Type: "portmap", Capabilities: map[string]bool{"portMappings": true}}
       firewall := plugin{Type: "firewall"}

       out := conflist{
               CNIVersion: "1.0.0",
               Name:       "k8s-pod-network",
               Plugins:    []interface{}{br, portmap, firewall},
       }
       return json.MarshalIndent(out, "", "  ")
}

// Bridge is a simple CNI manager for single-node usage
type Bridge struct {
	cc config.ClusterConfig
}

// String returns a string representation of this CNI
func (c Bridge) String() string {
	return "bridge CNI"
}

func (c Bridge) netconf() (assets.CopyableFile, error) {
	cfgBytes, err := renderBridgeConflist(c.cc.KubernetesConfig)
        if err != nil {
                return nil, err
        }
        return assets.NewMemoryAssetTarget(cfgBytes, "/etc/cni/net.d/1-k8s.conflist", "0644"), nil
}

// Apply enables the CNI
func (c Bridge) Apply(r Runner) error {
	if len(c.cc.Nodes) > 1 {
		return fmt.Errorf("bridge CNI is incompatible with multi-node clusters")
	}

	if _, err := r.RunCmd(exec.Command("sudo", "mkdir", "-p", "/etc/cni/net.d")); err != nil {
		return err
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

       // Prefer explicitly-set CIDRs from the cluster config.
       k := c.cc.KubernetesConfig
       if k.PodCIDRv6 != "" && (strings.ToLower(k.IPFamily) == "ipv6" || k.PodCIDR == "") {
               return k.PodCIDRv6
       }	

       if k.PodCIDR != "" {
               return k.PodCIDR
       }
       return DefaultPodCIDR
}
