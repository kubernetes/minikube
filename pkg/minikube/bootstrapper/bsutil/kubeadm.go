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
	"path"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/ktmpl"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

// Container runtimes
const remoteContainerRuntime = "remote"

// GenerateKubeadmYAML generates the kubeadm.yaml file
func GenerateKubeadmYAML(mc config.ClusterConfig, r cruntime.Manager, n config.Node) ([]byte, error) {
	k8s := mc.KubernetesConfig
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parsing kubernetes version")
	}

	// parses a map of the feature gates for kubeadm and component
	kubeadmFeatureArgs, componentFeatureArgs, err := parseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return nil, errors.Wrap(err, "parses feature gate config for kubeadm and component")
	}

	// In case of no port assigned, use default
	cp, err := config.PrimaryControlPlane(mc)
	if err != nil {
		return nil, errors.Wrap(err, "getting control plane")
	}
	nodePort := cp.Port
	if nodePort <= 0 {
		nodePort = constants.APIServerPort
	}

	componentOpts, err := createExtraComponentConfig(k8s.ExtraOptions, version, componentFeatureArgs, cp)
	if err != nil {
		return nil, errors.Wrap(err, "generating extra component config for kubeadm")
	}

	opts := struct {
		CertDir           string
		ServiceCIDR       string
		PodSubnet         string
		AdvertiseAddress  string
		APIServerPort     int
		KubernetesVersion string
		EtcdDataDir       string
		ClusterName       string
		NodeName          string
		DNSDomain         string
		CRISocket         string
		ImageRepository   string
		ComponentOptions  []componentOptions
		FeatureArgs       map[string]bool
		NoTaintMaster     bool
		NodeIP            string
		ControlPlaneIP    string
	}{
		CertDir:           vmpath.GuestKubernetesCertsDir,
		ServiceCIDR:       constants.DefaultServiceCIDR,
		PodSubnet:         k8s.ExtraOptions.Get("pod-network-cidr", Kubeadm),
		AdvertiseAddress:  cp.IP,
		APIServerPort:     nodePort,
		KubernetesVersion: k8s.KubernetesVersion,
		EtcdDataDir:       EtcdDataDir(),
		ClusterName:       k8s.ClusterName,
		NodeName:          cp.Name,
		CRISocket:         r.SocketPath(),
		ImageRepository:   k8s.ImageRepository,
		ComponentOptions:  componentOpts,
		FeatureArgs:       kubeadmFeatureArgs,
		NoTaintMaster:     false, // That does not work with k8s 1.12+
		DNSDomain:         k8s.DNSDomain,
		NodeIP:            n.IP,
		ControlPlaneIP:    cp.IP,
	}

	if k8s.ServiceCIDR != "" {
		opts.ServiceCIDR = k8s.ServiceCIDR
	}

	opts.NoTaintMaster = true
	b := bytes.Buffer{}
	configTmpl := ktmpl.V1Alpha1
	if version.GTE(semver.MustParse("1.12.0")) {
		configTmpl = ktmpl.V1Alpha3
	}
	// v1beta1 works in v1.13, but isn't required until v1.14.
	if version.GTE(semver.MustParse("1.14.0-alpha.0")) {
		configTmpl = ktmpl.V1Beta1
	}
	// v1beta2 isn't required until v1.17.
	if version.GTE(semver.MustParse("1.17.0")) {
		configTmpl = ktmpl.V1Beta2
	}
	glog.Infof("kubeadm options: %+v", opts)
	if err := configTmpl.Execute(&b, opts); err != nil {
		return nil, err
	}
	glog.Infof("kubeadm config:\n%s\n", b.String())
	return b.Bytes(), nil
}

// These are the components that can be configured
// through the "extra-config"
const (
	Kubelet           = "kubelet"
	Kubeadm           = "kubeadm"
	Apiserver         = "apiserver"
	Scheduler         = "scheduler"
	ControllerManager = "controller-manager"
)

// InvokeKubeadm returns the invocation command for Kubeadm
func InvokeKubeadm(version string) string {
	return fmt.Sprintf("sudo env PATH=%s:$PATH kubeadm", binRoot(version))
}

// EtcdDataDir is where etcd data is stored.
func EtcdDataDir() string {
	return path.Join(vmpath.GuestPersistentDir, "etcd")
}
