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
	"net"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// LogOptions are options to be passed to LogCommands
type LogOptions struct {
	// Lines is the number of recent log lines to include, as in tail -n.
	Lines int
	// Follow is whether or not to actively follow the logs, as in tail -f.
	Follow bool
}

// Bootstrapper contains all the methods needed to bootstrap a kubernetes cluster
type Bootstrapper interface {
	// PullImages pulls images necessary for a cluster. Success should not be required.
	PullImages(config.KubernetesConfig) error
	StartCluster(config.KubernetesConfig) error
	UpdateCluster(config.KubernetesConfig) error
	RestartCluster(config.KubernetesConfig) error
	DeleteCluster(config.KubernetesConfig) error
	// LogCommands returns a map of log type to a command which will display that log.
	LogCommands(LogOptions) map[string]string
	SetupCerts(cfg config.KubernetesConfig) error
	GetKubeletStatus() (string, error)
	GetAPIServerStatus(net.IP, int) (string, error)
}

const (
	// BootstrapperTypeKubeadm is the kubeadm bootstrapper type
	BootstrapperTypeKubeadm = "kubeadm"
)

// GetCachedBinaryList returns the list of binaries
func GetCachedBinaryList(bootstrapper string) []string {
	switch bootstrapper {
	case BootstrapperTypeKubeadm:
		return constants.GetKubeadmCachedBinaries()
	default:
		return []string{}
	}
}

// GetCachedImageList returns the list of images for a version
func GetCachedImageList(imageRepository string, version string, bootstrapper string) []string {
	switch bootstrapper {
	case BootstrapperTypeKubeadm:
		_, images := constants.GetKubeadmCachedImages(imageRepository, version)
		return images
	default:
		return []string{}
	}
}
