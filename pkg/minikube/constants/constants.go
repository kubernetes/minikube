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
	"errors"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/minikube/pkg/minikube/localpath"
)

const (
	// DefaultKubernetesVersion is the default Kubernetes version
	DefaultKubernetesVersion = "v1.19.0-rc.4"
	// NewestKubernetesVersion is the newest Kubernetes version to test against
	// NOTE: You may need to update coreDNS & etcd versions in pkg/minikube/bootstrapper/images/images.go
	NewestKubernetesVersion = "v1.19.0-rc.4"
	// OldestKubernetesVersion is the oldest Kubernetes version to test against
	OldestKubernetesVersion = "v1.13.0"
	// DefaultClusterName is the default nane for the k8s cluster
	DefaultClusterName = "minikube"
	// DockerDaemonPort is the port Docker daemon listening inside a minikube node (vm or container).
	DockerDaemonPort = 2376
	// APIServerPort is the default API server port
	APIServerPort = 8443
	// SSHPort is the SSH serviceport on the node vm and container
	SSHPort = 22
	// RegistryAddonPort os the default registry addon port
	RegistryAddonPort = 5000
	// CRIO is the default name and spelling for the cri-o container runtime
	CRIO = "crio"

	// APIServerName is the default API server name
	APIServerName = "minikubeCA"
	// ClusterDNSDomain is the default DNS domain
	ClusterDNSDomain = "cluster.local"
	// DefaultServiceCIDR is The CIDR to be used for service cluster IPs
	DefaultServiceCIDR = "10.96.0.0/12"
	// HostAlias is a DNS alias to the the container/VM host IP
	HostAlias = "host.minikube.internal"
	// ControlPlaneAlias is a DNS alias pointing to the apiserver frontend
	ControlPlaneAlias = "control-plane.minikube.internal"

	// DockerHostEnv is used for docker daemon settings
	DockerHostEnv = "DOCKER_HOST"
	// DockerCertPathEnv is used for docker daemon settings
	DockerCertPathEnv = "DOCKER_CERT_PATH"
	// DockerTLSVerifyEnv is used for docker daemon settings
	DockerTLSVerifyEnv = "DOCKER_TLS_VERIFY"
	// MinikubeActiveDockerdEnv holds the docker daemon which user's shell is pointing at
	// value would be profile or empty if pointing to the user's host daemon.
	MinikubeActiveDockerdEnv = "MINIKUBE_ACTIVE_DOCKERD"
	// PodmanVarlinkBridgeEnv is used for podman settings
	PodmanVarlinkBridgeEnv = "PODMAN_VARLINK_BRIDGE"
	// MinikubeActivePodmanEnv holds the podman service that the user's shell is pointing at
	// value would be profile or empty if pointing to the user's host.
	MinikubeActivePodmanEnv = "MINIKUBE_ACTIVE_PODMAN"
	// MinikubeForceSystemdEnv is used to force systemd as cgroup manager for the container runtime
	MinikubeForceSystemdEnv = "MINIKUBE_FORCE_SYSTEMD"
	// TestDiskUsedEnv is used in integration tests for insufficient storage with 'minikube status'
	TestDiskUsedEnv = "MINIKUBE_TEST_STORAGE_CAPACITY"
)

var (
	// IsMinikubeChildProcess is the name of "is minikube child process" variable
	IsMinikubeChildProcess = "IS_MINIKUBE_CHILD_PROCESS"
	// GvisorConfigTomlTargetName is the go-bindata target name for the gvisor config.toml
	GvisorConfigTomlTargetName = "gvisor-config.toml"
	// MountProcessFileName is the filename of the mount process
	MountProcessFileName = ".mount-process"

	// SHASuffix is the suffix of a SHA-256 checksum file
	SHASuffix = ".sha256"

	// DockerDaemonEnvs is list of docker-daemon related environment variables.
	DockerDaemonEnvs = [3]string{DockerHostEnv, DockerTLSVerifyEnv, DockerCertPathEnv}
	// PodmanRemoteEnvs is list of podman-remote related environment variables.
	PodmanRemoteEnvs = [1]string{PodmanVarlinkBridgeEnv}

	// DefaultMinipath is the default minikube path (under the home directory)
	DefaultMinipath = filepath.Join(homedir.HomeDir(), ".minikube")

	// KubeconfigEnvVar is the env var to check for the Kubernetes client config
	KubeconfigEnvVar = clientcmd.RecommendedConfigPathEnvVar
	// KubeconfigPath is the path to the Kubernetes client config
	KubeconfigPath = clientcmd.RecommendedHomeFile

	// ImageRepositories contains all known image repositories
	ImageRepositories = map[string][]string{
		"global": {""},
		"cn":     {"registry.cn-hangzhou.aliyuncs.com/google_containers"},
	}
	// KubernetesReleaseBinaries are Kubernetes release binaries required for
	// kubeadm (kubelet, kubeadm) and the addon manager (kubectl)
	KubernetesReleaseBinaries = []string{"kubelet", "kubeadm", "kubectl"}
	// ImageCacheDir is the path to the image cache directory
	ImageCacheDir = localpath.MakeMiniPath("cache", "images")

	// DefaultNamespaces are Kubernetes namespaces used by minikube, including addons
	DefaultNamespaces = []string{
		"kube-system",
		"kubernetes-dashboard",
		"storage-gluster",
		"istio-operator",
	}

	// ErrMachineMissing is returned when virtual machine does not exist due to user interrupt cancel(i.e. Ctrl + C)
	ErrMachineMissing = errors.New("machine does not exist")
)
