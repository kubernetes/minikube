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
	"io"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// Bootstrapper contains all the methods needed to bootstrap a kubernetes cluster
type Bootstrapper interface {
	StartCluster(config.KubernetesConfig) error
	UpdateCluster(config.KubernetesConfig) error
	RestartCluster(config.KubernetesConfig) error
	GetClusterLogsTo(follow bool, out io.Writer) error
	SetupCerts(cfg config.KubernetesConfig) error
	GetClusterStatus() (string, error)
}

const (
	BootstrapperTypeKubeadm = "kubeadm"
)

func GetCachedImageList(version string, bootstrapper string) []string {
	switch bootstrapper {
	case BootstrapperTypeKubeadm:
		return constants.GetKubeadmCachedImages(version)
	default:
		return []string{}
	}
}
