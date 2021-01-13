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
	"os"
	"path"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/ktmpl"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/util"
)

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
	if k8s.NetworkPlugin != "" {
		extraOpts["network-plugin"] = k8s.NetworkPlugin

		if k8s.NetworkPlugin == "kubenet" {
			extraOpts["pod-cidr"] = cni.DefaultPodCIDR
		}
	}

	if _, ok := extraOpts["node-ip"]; !ok {
		extraOpts["node-ip"] = nc.IP
	}
	if _, ok := extraOpts["hostname-override"]; !ok {
		nodeName := KubeNodeName(mc, nc)
		extraOpts["hostname-override"] = nodeName
	}

	pauseImage := images.Pause(version, k8s.ImageRepository)
	if _, ok := extraOpts["pod-infra-container-image"]; !ok && k8s.ImageRepository != "" && pauseImage != "" && k8s.ContainerRuntime != remoteContainerRuntime {
		extraOpts["pod-infra-container-image"] = pauseImage
	}

	// parses a map of the feature gates for kubelet
	_, kubeletFeatureArgs, err := parseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return nil, errors.Wrap(err, "parses feature gate config for kubelet")
	}

	if kubeletFeatureArgs != "" {
		extraOpts["feature-gates"] = kubeletFeatureArgs
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
