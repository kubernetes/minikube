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

// Package bsutil will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/ktmpl"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/util"
)

// kubeletConfigParams are the only allowed kubelet parameters for kubeadmin config file and not to be used as kubelet flags
// ref: https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/ - look for "DEPRECATED" flags
// ref: https://kubernetes.io/docs/tasks/administer-cluster/kubelet-config-file/
// ref: https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/#kubelet-config-k8s-io-v1beta1-KubeletConfiguration
var kubeletConfigParams = []string{
	"localStorageCapacityIsolation",
	"runtime-request-timeout",
	"hairpin-mode",
}

func extraKubeletOpts(mc config.ClusterConfig, nc config.Node, r cruntime.Manager) (map[string]string, error) {
	k8s := mc.KubernetesConfig
	version, err := util.ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parsing Kubernetes version")
	}

	extraOpts, err := extraConfigForComponent(Kubelet, k8s.ExtraOptions, version)
	if err != nil {
		return nil, errors.Wrap(err, "generating extra configuration for kubelet")
	}

	for k, v := range r.KubeletOptions() {
		extraOpts[k] = v
	}

	// avoid "Failed to start ContainerManager failed to initialise top level QOS containers" error (ref: https://github.com/kubernetes/kubernetes/issues/43856)
	// avoid "kubelet crashes with: root container [kubepods] doesn't exist" (ref: https://github.com/kubernetes/kubernetes/issues/95488)
	if mc.Driver == oci.Docker && mc.KubernetesConfig.ContainerRuntime == constants.CRIO {
		extraOpts["cgroups-per-qos"] = "false"
		extraOpts["enforce-node-allocatable"] = ""
	}

	if k8s.NetworkPlugin != "" {
		// Only CNI is supported in 1.24+, and it is the default
		if version.LT(semver.MustParse("1.24.0-alpha.2")) {
			extraOpts["network-plugin"] = k8s.NetworkPlugin
		} else if k8s.NetworkPlugin != "cni" && mc.KubernetesConfig.ContainerRuntime != constants.Docker {
			return nil, fmt.Errorf("invalid network plugin: %s", k8s.NetworkPlugin)
		}

		if k8s.NetworkPlugin == "kubenet" {
			extraOpts["pod-cidr"] = cni.DefaultPodCIDR
		}
	}

        // Pick node-ip based on requested IP family
        if _, ok := extraOpts["node-ip"]; !ok {
                family := strings.ToLower(k8s.IPFamily)
                switch family {
                case "ipv6":
                        if nc.IPv6 != "" {
                                extraOpts["node-ip"] = nc.IPv6
                        } else {
                                // fallback if IPv6 wasn’t wired yet
                                extraOpts["node-ip"] = nc.IP
                        }
                case "dual":
                        // Don’t set node-ip at all; kubelet will advertise both families.
                        // (If a user explicitly set node-ip, we honor it above.)
                default: // "ipv4" or empty
                        extraOpts["node-ip"] = nc.IP
                }
        }

	if _, ok := extraOpts["hostname-override"]; !ok {
		nodeName := KubeNodeName(mc, nc)
		extraOpts["hostname-override"] = nodeName
	}

	// Handled by CRI in 1.24+, and not by kubelet
	if version.LT(semver.MustParse("1.24.0-alpha.2")) {
		pauseImage := images.Pause(version, k8s.ImageRepository)
		if _, ok := extraOpts["pod-infra-container-image"]; !ok && k8s.ImageRepository != "" && pauseImage != "" && k8s.ContainerRuntime != remoteContainerRuntime {
			extraOpts["pod-infra-container-image"] = pauseImage
		}
	}

	// container-runtime-endpoint kubelet flag was deprecated but corresponding containerRuntimeEndpoint kubelet config field is "required" and supported from k8s v1.27
	// ref: https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/#options
	// ref: https://github.com/kubernetes/kubernetes/issues/118787
	if version.GTE(semver.MustParse("1.27.0")) {
		kubeletConfigParams = append(kubeletConfigParams, "container-runtime-endpoint")
	}

	// parses a map of the feature gates for kubelet
	_, kubeletFeatureArgs, err := parseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return nil, errors.Wrap(err, "parses feature gate config for kubelet")
	}

	if kubeletFeatureArgs != "" {
		extraOpts["feature-gates"] = kubeletFeatureArgs
	}

	// filter out non-flag extra kubelet config options
	for _, opt := range kubeletConfigParams {
		delete(extraOpts, opt)
	}

	return extraOpts, nil
}

// NewKubeletConfig generates a new systemd unit containing a configured kubelet
// based on the options present in the KubernetesConfig.
func NewKubeletConfig(mc config.ClusterConfig, nc config.Node, r cruntime.Manager) ([]byte, error) {
	b := bytes.Buffer{}
	extraOpts, err := extraKubeletOpts(mc, nc, r)
	if err != nil {
		return nil, err
	}
	k8s := mc.KubernetesConfig
	opts := struct {
		ExtraOptions     string
		ContainerRuntime string
		KubeletPath      string
	}{
		ExtraOptions:     convertToFlags(extraOpts),
		ContainerRuntime: k8s.ContainerRuntime,
		KubeletPath:      path.Join(binRoot(k8s.KubernetesVersion), "kubelet"),
	}
	if err := ktmpl.KubeletSystemdTemplate.Execute(&b, opts); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// NewKubeletService returns a generated systemd unit file for the kubelet
func NewKubeletService(cfg config.KubernetesConfig) ([]byte, error) {
	var b bytes.Buffer
	opts := struct{ KubeletPath string }{KubeletPath: path.Join(binRoot(cfg.KubernetesVersion), "kubelet")}
	if err := ktmpl.KubeletServiceTemplate.Execute(&b, opts); err != nil {
		return nil, errors.Wrap(err, "template execute")
	}
	return b.Bytes(), nil
}

// KubeNodeName returns the node name registered in Kubernetes
func KubeNodeName(cc config.ClusterConfig, n config.Node) string {
	if cc.Driver == driver.None {
		// Always use hostname for "none" driver
		hostname, _ := os.Hostname()
		return hostname
	}
	return config.MachineName(cc, n)
}
