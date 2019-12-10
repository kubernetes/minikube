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

package constants

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/minikube/pkg/minikube/localpath"
	minikubeVersion "k8s.io/minikube/pkg/version"
)

const (
	// APIServerPort is the default API server port
	APIServerPort = 8443
	// APIServerName is the default API server name
	APIServerName = "minikubeCA"
	// ClusterDNSDomain is the default DNS domain
	ClusterDNSDomain = "cluster.local"
)

// DefaultMinipath is the default Minikube path (under the home directory)
var DefaultMinipath = filepath.Join(homedir.HomeDir(), ".minikube")

// KubeconfigPath is the path to the Kubernetes client config
var KubeconfigPath = clientcmd.RecommendedHomeFile

// KubeconfigEnvVar is the env var to check for the Kubernetes client config
var KubeconfigEnvVar = clientcmd.RecommendedConfigPathEnvVar

// DefaultMachineName is the default name for the VM
const DefaultMachineName = "minikube"

// DefaultNodeName is the default name for the kubeadm node within the VM
const DefaultNodeName = "minikube"

// MountProcessFileName is the filename of the mount process
var MountProcessFileName = ".mount-process"

const (
	// SHASuffix is the suffix of a SHA-256 checksum file
	SHASuffix = ".sha256"
)

// DefaultISOURL is the default location of the minikube.iso file
var DefaultISOURL = fmt.Sprintf("https://storage.googleapis.com/%s/minikube-%s.iso", minikubeVersion.GetISOPath(), minikubeVersion.GetISOVersion())

// DefaultISOSHAURL is the default location of the minikube.iso.sha256 file
var DefaultISOSHAURL = DefaultISOURL + SHASuffix

// DefaultKubernetesVersion is the default kubernetes version
var DefaultKubernetesVersion = "v1.17.0"

// NewestKubernetesVersion is the newest Kubernetes version to test against
var NewestKubernetesVersion = "v1.17.0"

// OldestKubernetesVersion is the oldest Kubernetes version to test against
var OldestKubernetesVersion = "v1.11.10"

const (
	// IsMinikubeChildProcess is the name of "is minikube child process" variable
	IsMinikubeChildProcess = "IS_MINIKUBE_CHILD_PROCESS"
)

// ImageRepositories contains all known image repositories
var ImageRepositories = map[string][]string{
	"global": {""},
	"cn":     {"registry.cn-hangzhou.aliyuncs.com/google_containers"},
}

// KubeadmBinaries are Kubernetes release binaries required for kubeadm
var KubeadmBinaries = []string{"kubelet", "kubeadm"}

// ImageCacheDir is the path to the image cache directory
var ImageCacheDir = localpath.MakeMiniPath("cache", "images")

const (
	// GvisorFilesPath is the path to the gvisor files saved by go-bindata
	GvisorFilesPath = "/tmp/gvisor"

	// GvisorConfigTomlTargetName is the go-bindata target name for the gvisor config.toml
	GvisorConfigTomlTargetName = "gvisor-config.toml"
)
