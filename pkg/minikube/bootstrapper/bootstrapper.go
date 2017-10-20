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

package bootstrapper

import (
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

// Bootstrapper contains all the methods needed to bootstrap a kubernetes cluster
type Bootstrapper interface {
	StartCluster(KubernetesConfig) error
	UpdateCluster(KubernetesConfig) error
	RestartCluster(KubernetesConfig) error
	GetClusterLogs(follow bool) (string, error)
	SetupCerts(cfg KubernetesConfig) error
	GetClusterStatus() (string, error)
}

// KubernetesConfig contains the parameters used to configure the VM Kubernetes.
type KubernetesConfig struct {
	KubernetesVersion string
	NodeIP            string
	NodeName          string
	APIServerName     string
	DNSDomain         string
	ContainerRuntime  string
	NetworkPlugin     string
	FeatureGates      string
	ServiceCIDR       string
	ExtraOptions      util.ExtraOptionSlice

	ShouldLoadCachedImages bool
}

const (
	BootstrapperTypeLocalkube = "localkube"
	BootstrapperTypeKubeadm   = "kubeadm"
)

func GetCachedImageList(version string, bootstrapper string) []string {
	switch bootstrapper {
	case BootstrapperTypeLocalkube:
		return constants.LocalkubeCachedImages
	case BootstrapperTypeKubeadm:
		return constants.GetKubeadmCachedImages(version)
	default:
		return []string{}
	}
}
