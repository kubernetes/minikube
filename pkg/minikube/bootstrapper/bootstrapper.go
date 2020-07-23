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
	"time"

	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

// LogOptions are options to be passed to LogCommands
type LogOptions struct {
	// Lines is the number of recent log lines to include, as in tail -n.
	Lines int
	// Follow is whether or not to actively follow the logs, as in tail -f.
	Follow bool
}

// Bootstrapper contains all the methods needed to bootstrap a Kubernetes cluster
type Bootstrapper interface {
	StartCluster(config.ClusterConfig) error
	UpdateCluster(config.ClusterConfig) error
	DeleteCluster(config.KubernetesConfig) error
	WaitForNode(config.ClusterConfig, config.Node, time.Duration) error
	JoinCluster(config.ClusterConfig, config.Node, string) error
	UpdateNode(config.ClusterConfig, config.Node, cruntime.Manager) error
	GenerateToken(config.ClusterConfig) (string, error)
	// LogCommands returns a map of log type to a command which will display that log.
	LogCommands(config.ClusterConfig, LogOptions) map[string]string
	SetupCerts(config.KubernetesConfig, config.Node) error
}

const (
	// Kubeadm is the kubeadm bootstrapper type
	Kubeadm = "kubeadm"
)

// GetCachedBinaryList returns the list of binaries
func GetCachedBinaryList(bootstrapper string) []string {
	return constants.KubernetesReleaseBinaries
}

// GetCachedImageList returns the list of images for a version
func GetCachedImageList(imageRepository string, version string, bootstrapper string) ([]string, error) {
	return images.Kubeadm(imageRepository, version)
}
