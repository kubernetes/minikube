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
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/ktmpl"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
)

// Container runtimes
const remoteContainerRuntime = "remote"

// GenerateKubeadmYAML generates the kubeadm.yaml file
func GenerateKubeadmYAML(cc config.ClusterConfig, n config.Node, r cruntime.Manager) ([]byte, error) {
	k8s := cc.KubernetesConfig
	version, err := util.ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parsing Kubernetes version")
	}

	// parses a map of the feature gates for kubeadm and component
	kubeadmFeatureArgs, componentFeatureArgs, err := parseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return nil, errors.Wrap(err, "parses feature gate config for kubeadm and component")
	}

	// In case of no port assigned, use default
	cp, err := config.PrimaryControlPlane(&cc)
	if err != nil {
		return nil, errors.Wrap(err, "getting control plane")
	}
	nodePort := cp.Port
	if nodePort <= 0 {
		nodePort = constants.APIServerPort
	}

	cgroupDriver, err := r.CGroupDriver()
	if err != nil {
		if driver.BareMetal(cc.Driver) && r.Name() == "Docker" && strings.Contains(err.Error(), "panic") {
			return nil, oci.ErrDockerNotRunning
		}
		return nil, errors.Wrap(err, "getting cgroup driver")
	}

	componentOpts, err := createExtraComponentConfig(k8s.ExtraOptions, version, componentFeatureArgs, cp)
	if err != nil {
		return nil, errors.Wrap(err, "generating extra component config for kubeadm")
	}

	cnm, err := cni.New(cc)
	if err != nil {
		return nil, errors.Wrap(err, "cni")
	}

	podCIDR := cnm.CIDR()
	overrideCIDR := k8s.ExtraOptions.Get("pod-network-cidr", Kubeadm)
	if overrideCIDR != "" {
		podCIDR = overrideCIDR
	}
	klog.Infof("Using pod CIDR: %s", podCIDR)

	opts := struct {
		CertDir             string
		ServiceCIDR         string
		PodSubnet           string
		AdvertiseAddress    string
		APIServerPort       int
		KubernetesVersion   string
		EtcdDataDir         string
		EtcdExtraArgs       map[string]string
		ClusterName         string
		NodeName            string
		DNSDomain           string
		CRISocket           string
		ImageRepository     string
		ComponentOptions    []componentOptions
		FeatureArgs         map[string]bool
		NoTaintMaster       bool
		NodeIP              string
		CgroupDriver        string
		ClientCAFile        string
		StaticPodPath       string
		ControlPlaneAddress string
		KubeProxyOptions    map[string]string
	}{
		CertDir:           vmpath.GuestKubernetesCertsDir,
		ServiceCIDR:       constants.DefaultServiceCIDR,
		PodSubnet:         podCIDR,
		AdvertiseAddress:  n.IP,
		APIServerPort:     nodePort,
		KubernetesVersion: k8s.KubernetesVersion,
		EtcdDataDir:       EtcdDataDir(),
		EtcdExtraArgs:     etcdExtraArgs(k8s.ExtraOptions),
		ClusterName:       cc.Name,
		// kubeadm uses NodeName as the --hostname-override parameter, so this needs to be the name of the machine
		NodeName:            KubeNodeName(cc, n),
		CRISocket:           r.SocketPath(),
		ImageRepository:     k8s.ImageRepository,
		ComponentOptions:    componentOpts,
		FeatureArgs:         kubeadmFeatureArgs,
		NoTaintMaster:       false, // That does not work with k8s 1.12+
		DNSDomain:           k8s.DNSDomain,
		NodeIP:              n.IP,
		CgroupDriver:        cgroupDriver,
		ClientCAFile:        path.Join(vmpath.GuestKubernetesCertsDir, "ca.crt"),
		StaticPodPath:       vmpath.GuestManifestsDir,
		ControlPlaneAddress: constants.ControlPlaneAlias,
		KubeProxyOptions:    createKubeProxyOptions(k8s.ExtraOptions),
	}

	if k8s.ServiceCIDR != "" {
		opts.ServiceCIDR = k8s.ServiceCIDR
	}

	opts.NoTaintMaster = true
	b := bytes.Buffer{}
	configTmpl := ktmpl.V1Alpha3
	// v1beta1 works in v1.13, but isn't required until v1.14.
	if version.GTE(semver.MustParse("1.14.0-alpha.0")) {
		configTmpl = ktmpl.V1Beta1
	}
	// v1beta2 isn't required until v1.17.
	if version.GTE(semver.MustParse("1.17.0")) {
		configTmpl = ktmpl.V1Beta2
	}
	klog.Infof("kubeadm options: %+v", opts)
	if err := configTmpl.Execute(&b, opts); err != nil {
		return nil, err
	}
	klog.Infof("kubeadm config:\n%s\n", b.String())
	return b.Bytes(), nil
}

// These are the components that can be configured
// through the "extra-config"
const (
	Apiserver         = "apiserver"
	ControllerManager = "controller-manager"
	Scheduler         = "scheduler"
	Etcd              = "etcd"
	Kubeadm           = "kubeadm"
	Kubeproxy         = "kube-proxy"
	Kubelet           = "kubelet"
)

// KubeadmExtraConfigOpts is a list of allowed "extra-config" components
var KubeadmExtraConfigOpts = []string{
	Apiserver,
	ControllerManager,
	Scheduler,
	Etcd,
	Kubeadm,
	Kubelet,
	Kubeproxy,
}

// InvokeKubeadm returns the invocation command for Kubeadm
func InvokeKubeadm(version string) string {
	return fmt.Sprintf("sudo env PATH=%s:$PATH kubeadm", binRoot(version))
}

// EtcdDataDir is where etcd data is stored.
func EtcdDataDir() string {
	return path.Join(vmpath.GuestPersistentDir, "etcd")
}

func etcdExtraArgs(extraOpts config.ExtraOptionSlice) map[string]string {
	args := map[string]string{}
	for _, eo := range extraOpts {
		if eo.Component != Etcd {
			continue
		}
		args[eo.Key] = eo.Value
	}
	return args
}
