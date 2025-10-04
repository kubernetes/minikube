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
	"strings"
	"time"
	"os/exec"

	// goembed needs this
	_ "embed"
	"text/template"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

// https://docs.projectcalico.org/manifests/calico.yaml
//
//go:embed calico.yaml
var calicoYaml string

// calicoTmpl is from https://docs.projectcalico.org/manifests/calico.yaml
var calicoTmpl = template.Must(template.New("calico").Parse(calicoYaml))

// Calico is the Calico CNI manager
type Calico struct {
	cc config.ClusterConfig
}

type calicoTmplStruct struct {
	// IPv4/IPv6/dual inputs for the template
	IPFamily                  string
	PodCIDR                   string
	PodCIDRv6                 string
	ServiceCIDR               string
	ServiceCIDRv6             string
	ControlPlaneAlias         string
	APIServerPort             int
	DeploymentImageName       string
	DaemonSetImageName        string
	BinaryImageName           string
	LegacyPodDisruptionBudget bool
	EnableIPv4     bool
    	EnableIPv6     bool
}

// String returns a string representation of this CNI
func (c Calico) String() string {
	return "Calico"
}

// manifest returns a Kubernetes manifest for a CNI
func (c Calico) manifest() (assets.CopyableFile, error) {
	k8sVersion, err := util.ParseKubernetesVersion(c.cc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Kubernetes version: %v", err)
	}
	
	k := c.cc.KubernetesConfig
	ipFamily := strings.ToLower(k.IPFamily)
	// Defaults/fallbacks to stay safe if flags weren’t provided
	v4Pod := k.PodCIDR
	if v4Pod == "" {
		v4Pod = DefaultPodCIDR
	}
	v6Pod := k.PodCIDRv6
       // Ensure an IPv6 pod CIDR exists when the ily is v6 or dual
       if v6Pod == "" && (ipFamily == "ipv6" || ipFamily == "dual") {
               v6Pod = constants.DefaultPodCIDRv6
       }
	svcV4 := k.ServiceCIDR
	if svcV4 == "" {
		svcV4 = constants.DefaultServiceCIDR
	}
	svcV6 := k.ServiceCIDRv6
	if svcV6 == "" && ipFamily != "ipv4" {
		svcV6 = constants.DefaultServiceCIDRv6
	}
	apiPort := c.cc.APIServerPort
        if apiPort == 0 {
                apiPort = 8443
        }

	enablev4 := (ipFamily == "ipv4" || ipFamily == "dual")
	enablev6 := (ipFamily == "ipv6" || ipFamily == "dual")

	input := &calicoTmplStruct{
		IPFamily:                  ipFamily,
		PodCIDR:                   v4Pod,
		PodCIDRv6:                 v6Pod,
		ServiceCIDR:               svcV4,
		ServiceCIDRv6:             svcV6,
		ControlPlaneAlias:         constants.ControlPlaneAlias,
		APIServerPort:             apiPort,
		DeploymentImageName:       images.CalicoDeployment(k.ImageRepository),
		DaemonSetImageName:        images.CalicoDaemonSet(k.ImageRepository),
		BinaryImageName:           images.CalicoBin(k.ImageRepository),
		LegacyPodDisruptionBudget: k8sVersion.LT(semver.Version{Major: 1, Minor: 25}),
		EnableIPv4: 			enablev4,
		EnableIPv6: 			enablev6,

	}

	b := bytes.Buffer{}
	if err := calicoTmpl.Execute(&b, input); err != nil {
		return nil, err
	}
	return manifestAsset(b.Bytes()), nil
}

// Apply enables the CNI
func (c Calico) Apply(r Runner) error {
	m, err := c.manifest()
	if err != nil {
		return errors.Wrap(err, "manifest")
	}
	// Phase 1: apply core Calico (includes CRDs)
	if err := applyManifest(c.cc, r, m); err != nil {
		return err
	}

	// Phase 2: wait for CRD to be Established, then apply IPPools
	if err := waitForCRDEstablished(r, c.cc.KubernetesConfig.KubernetesVersion, "ippools.crd.projectcalico.org", 90*time.Second); err != nil {
		// Non-fatal: log and try to continue; but usually this must succeed.
		return errors.Wrap(err, "waiting for Calico IPPool CRD")
	}

	ipPoolsYAML := renderCalicoIPPools(c.cc.KubernetesConfig)
	if ipPoolsYAML == "" {
		return nil
	}
	poolsAsset := assets.NewMemoryAssetTarget([]byte(ipPoolsYAML), "/var/tmp/minikube/calico-ippools.yaml", "0644")
	return applyManifest(c.cc, r, poolsAsset)
}

// waitForCRDEstablished waits until the given CRD reports Established=True.
func waitForCRDEstablished(r Runner, k8sVersion string, crd string, to time.Duration) error {
	kubectlPath := fmt.Sprintf("/var/lib/minikube/binaries/%s/kubectl", k8sVersion)
	cmd := exec.Command("sudo", kubectlPath, "wait",
		"--kubeconfig=/var/lib/minikube/kubeconfig",
		"--for=condition=Established",
		fmt.Sprintf("--timeout=%ds", int(to.Seconds())),
		"crd/"+crd,
	)
	_, err := r.RunCmd(cmd)
	return err
}

// renderCalicoIPPools returns a small manifest for IPv4/IPv6 IPPools based on IPFamily and PodCIDRs.
func renderCalicoIPPools(k config.KubernetesConfig) string {
    fam := strings.ToLower(k.IPFamily)
    var b strings.Builder
    emit := func(name, cidr string, nat bool, isV6 bool) {
    if cidr == "" {
        return
    }
    if b.Len() > 0 {
        b.WriteString("\n---\n")
    }

        // IPPool instances are served under projectcalico.org/v3.
        fmt.Fprintf(&b,
            "apiVersion: projectcalico.org/v3\nkind: IPPool\nmetadata:\n  name: %s\nspec:\n  cidr: %q\n  disabled: false\n  natOutgoing: %t\n  nodeSelector: \"all()\"\n",
            name, cidr, nat)
        if !isV6 {
            // Match calico-node env (IPIP Always, VXLAN Never) for IPv4.
            b.WriteString("  ipipMode: Always\n  vxlanMode: Never\n")
        } else {
            // For IPv6, IPIP is IPv4-only; keep VXLAN off to match DS env.
            b.WriteString("  vxlanMode: Never\n")
        }
     }

// IPv4 pool unless explicitly ipv6-only
	if  fam != "ipv6" {
		cidr := k.PodCIDR
		if cidr == "" {
			cidr = DefaultPodCIDR
		}
		emit("default-ipv4-ippool", cidr, true, false)
	}
	// IPv6 pool for ipv6 or dual
	if  fam == "ipv6" || fam == "dual" {
		cidr := k.PodCIDRv6
		if cidr == "" {
			// default provided by your constants if you prefer:
			// cidr = constants.DefaultPodCIDRv6
			// but safer to require user/normalizer to have set it.
			cidr = constants.DefaultPodCIDRv6
		}
		emit("default-ipv6-ippool", cidr, false, true)
	}
	return b.String()
}

// CIDR returns the default CIDR used by this CNI
func (c Calico) CIDR() string {
	// Calico docs specify 192.168.0.0/16 - but we do this for compatibility with other CNI's.
	k := c.cc.KubernetesConfig
	 fam := strings.ToLower(k.IPFamily)
	// Prefer explicitly-set CIDRs; for ipv6-only prefer v6
	if k.PodCIDRv6 != "" && ( fam == "ipv6" || k.PodCIDR == "") {
		return k.PodCIDRv6
	}
	if k.PodCIDR != "" {
		return k.PodCIDR
	}
	// fallback for legacy behavior
	return DefaultPodCIDR
}
