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

// bsutil package will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"bytes"
	"fmt"
	"path"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
)

// Container runtimes
const remoteContainerRuntime = "remote"

// GenerateKubeadmYAML generates the kubeadm.yaml file
func GenerateKubeadmYAML(k8s config.KubernetesConfig, r cruntime.Manager) ([]byte, error) {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parsing kubernetes version")
	}

	// parses a map of the feature gates for kubeadm and component
	kubeadmFeatureArgs, componentFeatureArgs, err := parseFeatureArgs(k8s.FeatureGates)
	if err != nil {
		return nil, errors.Wrap(err, "parses feature gate config for kubeadm and component")
	}

	extraComponentConfig, err := createExtraComponentConfig(k8s.ExtraOptions, version, componentFeatureArgs)
	if err != nil {
		return nil, errors.Wrap(err, "generating extra component config for kubeadm")
	}

	// In case of no port assigned, use util.APIServerPort
	nodePort := k8s.NodePort
	if nodePort <= 0 {
		nodePort = constants.APIServerPort
	}

	opts := struct {
		CertDir           string
		ServiceCIDR       string
		PodSubnet         string
		AdvertiseAddress  string
		APIServerPort     int
		KubernetesVersion string
		EtcdDataDir       string
		NodeName          string
		DNSDomain         string
		CRISocket         string
		ImageRepository   string
		ExtraArgs         []ComponentExtraArgs
		FeatureArgs       map[string]bool
		NoTaintMaster     bool
	}{
		CertDir:           vmpath.GuestCertsDir,
		ServiceCIDR:       util.DefaultServiceCIDR,
		PodSubnet:         k8s.ExtraOptions.Get("pod-network-cidr", Kubeadm),
		AdvertiseAddress:  k8s.NodeIP,
		APIServerPort:     nodePort,
		KubernetesVersion: k8s.KubernetesVersion,
		EtcdDataDir:       etcdDataDir(),
		NodeName:          k8s.NodeName,
		CRISocket:         r.SocketPath(),
		ImageRepository:   k8s.ImageRepository,
		ExtraArgs:         extraComponentConfig,
		FeatureArgs:       kubeadmFeatureArgs,
		NoTaintMaster:     false, // That does not work with k8s 1.12+
		DNSDomain:         k8s.DNSDomain,
	}

	if k8s.ServiceCIDR != "" {
		opts.ServiceCIDR = k8s.ServiceCIDR
	}

	opts.NoTaintMaster = true
	b := bytes.Buffer{}
	configTmpl := ConfigTmplV1Alpha1
	if version.GTE(semver.MustParse("1.12.0")) {
		configTmpl = ConfigTmplV1Alpha3
	}
	// v1beta1 works in v1.13, but isn't required until v1.14.
	if version.GTE(semver.MustParse("1.14.0-alpha.0")) {
		configTmpl = ConfigTmplV1Beta1
	}
	if err := configTmpl.Execute(&b, opts); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// NewKubeletConfig generates a new systemd unit containing a configured kubelet
// based on the options present in the KubernetesConfig.
func NewKubeletConfig(k8s config.KubernetesConfig, r cruntime.Manager) ([]byte, error) {
	version, err := ParseKubernetesVersion(k8s.KubernetesVersion)
	if err != nil {
		return nil, errors.Wrap(err, "parsing kubernetes version")
	}

	extraOpts, err := ExtraConfigForComponent(Kubelet, k8s.ExtraOptions, version)
	if err != nil {
		return nil, errors.Wrap(err, "generating extra configuration for kubelet")
	}

	for k, v := range r.KubeletOptions() {
		extraOpts[k] = v
	}
	if k8s.NetworkPlugin != "" {
		extraOpts["network-plugin"] = k8s.NetworkPlugin
	}
	if _, ok := extraOpts["node-ip"]; !ok {
		extraOpts["node-ip"] = k8s.NodeIP
	}

	pauseImage := images.Pause(k8s.ImageRepository)
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

	b := bytes.Buffer{}
	opts := struct {
		ExtraOptions     string
		ContainerRuntime string
		KubeletPath      string
	}{
		ExtraOptions:     convertToFlags(extraOpts),
		ContainerRuntime: k8s.ContainerRuntime,
		KubeletPath:      path.Join(binRoot(k8s.KubernetesVersion), "kubelet"),
	}
	if err := KubeletSystemdTemplate.Execute(&b, opts); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// NewKubeletService returns a generated systemd unit file for the kubelet
func NewKubeletService(cfg config.KubernetesConfig) ([]byte, error) {
	var b bytes.Buffer
	opts := struct{ KubeletPath string }{KubeletPath: path.Join(binRoot(cfg.KubernetesVersion), "kubelet")}
	if err := kubeletServiceTemplate.Execute(&b, opts); err != nil {
		return nil, errors.Wrap(err, "template execute")
	}
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

// ExtraConfigForComponent generates a map of flagname-value pairs for a k8s
// component.
func ExtraConfigForComponent(component string, opts config.ExtraOptionSlice, version semver.Version) (map[string]string, error) {
	versionedOpts, err := DefaultOptionsForComponentAndVersion(component, version)
	if err != nil {
		return nil, errors.Wrapf(err, "setting version specific options for %s", component)
	}

	for _, opt := range opts {
		if opt.Component == component {
			if val, ok := versionedOpts[opt.Key]; ok {
				glog.Infof("Overwriting default %s=%s with user provided %s=%s for component %s", opt.Key, val, opt.Key, opt.Value, component)
			}
			versionedOpts[opt.Key] = opt.Value
		}
	}

	return versionedOpts, nil
}

// binRoot returns the persistent path binaries are stored in
func binRoot(version string) string {
	return path.Join(vmpath.GuestPersistentDir, "binaries", version)
}

// InvokeKubeadm returns the invocation command for Kubeadm
func InvokeKubeadm(version string) string {
	return fmt.Sprintf("sudo env PATH=%s:$PATH kubeadm", binRoot(version))
}

// etcdDataDir is where etcd data is stored.
func etcdDataDir() string {
	return path.Join(vmpath.GuestPersistentDir, "etcd")
}
