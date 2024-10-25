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
	"path"
	"path/filepath"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

var (
	// SupportedArchitectures is the list of supported architectures
	SupportedArchitectures = [5]string{"amd64", "arm", "arm64", "ppc64le", "s390x"}
)

const (
	// DefaultKubernetesVersion is the default Kubernetes version
	DefaultKubernetesVersion = "v1.31.2"
	// NewestKubernetesVersion is the newest Kubernetes version to test against
	// NOTE: You may need to update coreDNS & etcd versions in pkg/minikube/bootstrapper/images/images.go
	NewestKubernetesVersion = "v1.31.2"
	// OldestKubernetesVersion is the oldest Kubernetes version to test against
	OldestKubernetesVersion = "v1.20.0"
	// NoKubernetesVersion is the version used when users does NOT want to install kubernetes
	NoKubernetesVersion = "v0.0.0"

	// DefaultClusterName is the default name for the k8s cluster
	DefaultClusterName = "minikube"
	// DockerDaemonPort is the port Docker daemon listening inside a minikube node (vm or container).
	DockerDaemonPort = 2376
	// APIServerPort is the default API server port
	APIServerPort = 8443
	// AutoPauseProxyPort is the port to be used as a reverse proxy for apiserver port
	AutoPauseProxyPort = 32443

	// SSHPort is the SSH serviceport on the node vm and container
	SSHPort = 22
	// RegistryAddonPort os the default registry addon port
	RegistryAddonPort = 5000

	// Containerd is the default name and spelling for the containerd container runtime
	Containerd = "containerd"
	// CRIO is the default name and spelling for the cri-o container runtime
	CRIO = "crio"
	// Docker is the default name and spelling for the docker container runtime
	Docker = "docker"
	// DefaultContainerRuntime is our default container runtime
	DefaultContainerRuntime = ""

	// cgroup drivers
	DefaultCgroupDriver  = "systemd"
	CgroupfsCgroupDriver = "cgroupfs"
	SystemdCgroupDriver  = "systemd"
	UnknownCgroupDriver  = ""

	// APIServerName is the default API server name
	APIServerName = "minikubeCA"
	// ClusterDNSDomain is the default DNS domain
	ClusterDNSDomain = "cluster.local"
	// DefaultServiceCIDR is The CIDR to be used for service cluster IPs
	DefaultServiceCIDR = "10.96.0.0/12"
	// HostAlias is a DNS alias to the container/VM host IP
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
	// SSHAuthSock is used for docker-env
	SSHAuthSock = "SSH_AUTH_SOCK"
	// SSHAgentPID is used for docker-env
	SSHAgentPID = "SSH_AGENT_PID"
	// PodmanVarlinkBridgeEnv is used for podman settings
	PodmanVarlinkBridgeEnv = "PODMAN_VARLINK_BRIDGE"
	// PodmanContainerHostEnv is used for podman settings
	PodmanContainerHostEnv = "CONTAINER_HOST"
	// PodmanContainerSSHKeyEnv is used for podman settings
	PodmanContainerSSHKeyEnv = "CONTAINER_SSHKEY"
	// MinikubeActivePodmanEnv holds the podman service that the user's shell is pointing at
	// value would be profile or empty if pointing to the user's host.
	MinikubeActivePodmanEnv = "MINIKUBE_ACTIVE_PODMAN"
	// MinikubeForceSystemdEnv is used to force systemd as cgroup manager for the container runtime
	MinikubeForceSystemdEnv = "MINIKUBE_FORCE_SYSTEMD"
	// TestDiskUsedEnv is used in integration tests for insufficient storage with 'minikube status' (in %)
	TestDiskUsedEnv = "MINIKUBE_TEST_STORAGE_CAPACITY"
	// TestDiskAvailableEnv is used in integration tests for insufficient storage with 'minikube status' (in GiB)
	TestDiskAvailableEnv = "MINIKUBE_TEST_AVAILABLE_STORAGE"
	// MinikubeRootlessEnv is used to force Rootless Docker/Podman driver
	MinikubeRootlessEnv = "MINIKUBE_ROOTLESS"

	// scheduled stop constants

	// ScheduledStopEnvFile is the environment file for scheduled-stop
	ScheduledStopEnvFile = "/var/lib/minikube/scheduled-stop/environment"
	// ScheduledStopSystemdService is the service file for scheduled-stop
	ScheduledStopSystemdService = "minikube-scheduled-stop"

	// MinikubeExistingPrefix is used to save the original environment when executing docker-env
	MinikubeExistingPrefix = "MINIKUBE_EXISTING_"

	// ExistingDockerHostEnv is used to save original docker environment
	ExistingDockerHostEnv = MinikubeExistingPrefix + "DOCKER_HOST"
	// ExistingDockerCertPathEnv is used to save original docker environment
	ExistingDockerCertPathEnv = MinikubeExistingPrefix + "DOCKER_CERT_PATH"
	// ExistingDockerTLSVerifyEnv is used to save original docker environment
	ExistingDockerTLSVerifyEnv = MinikubeExistingPrefix + "DOCKER_TLS_VERIFY"

	// ExistingContainerHostEnv is used to save original podman environment
	ExistingContainerHostEnv = MinikubeExistingPrefix + "CONTAINER_HOST"

	// TimeFormat is the format that should be used when outputting time
	TimeFormat = time.RFC822
	// MaxResources is the value that can be passed into the memory and cpus flags to specify to use maximum resources
	MaxResources = "max"
	// NoLimit is the value that can be passed into the memory and cpus flags to specify to not set the resource limit on the container (Docker & Podman only)
	NoLimit = "no-limit"

	// DefaultCertExpiration is the amount of time in the future a certificate will expire in by default, which is 3 years
	DefaultCertExpiration = time.Hour * 24 * 365 * 3

	// Mount9PVersionFlag is the flag used to set the mount 9P version
	Mount9PVersionFlag = "9p-version"
	// MountGIDFlag is the flag used to set the mount GID
	MountGIDFlag = "gid"
	// MountIPFlag is the flag used to set the mount IP
	MountIPFlag = "ip"
	// MountMSizeFlag is the flag used to set the mount msize
	MountMSizeFlag = "msize"
	// MountOptionsFlag is the flag used to set the mount options
	MountOptionsFlag = "options"
	// MountPortFlag is the flag used to set the mount port
	MountPortFlag = "port"
	// MountTypeFlag is the flag used to set the mount type
	MountTypeFlag = "type"
	// MountUIDFlag is the flag used to set the mount UID
	MountUIDFlag = "uid"

	// Mirror CN
	AliyunMirror = "registry.cn-hangzhou.aliyuncs.com/google_containers"
)

var (
	// IsMinikubeChildProcess is the name of "is minikube child process" variable
	IsMinikubeChildProcess = "IS_MINIKUBE_CHILD_PROCESS"
	// MountProcessFileName is the filename of the mount process
	MountProcessFileName = ".mount-process"

	// SHASuffix is the suffix of a SHA-256 checksum file
	SHASuffix = ".sha256"

	// DockerDaemonEnvs is list of docker-daemon related environment variables.
	DockerDaemonEnvs = [3]string{DockerHostEnv, DockerTLSVerifyEnv, DockerCertPathEnv}
	// ExistingDockerDaemonEnvs is list of docker-daemon related environment variables.
	ExistingDockerDaemonEnvs = [3]string{ExistingDockerHostEnv, ExistingDockerTLSVerifyEnv, ExistingDockerCertPathEnv}

	// PodmanRemoteEnvs is list of podman-remote related environment variables.
	PodmanRemoteEnvs = [2]string{PodmanVarlinkBridgeEnv, PodmanContainerHostEnv}

	// DefaultMinipath is the default minikube path (under the home directory)
	DefaultMinipath = filepath.Join(homedir.HomeDir(), ".minikube")

	// KubeconfigEnvVar is the env var to check for the Kubernetes client config
	KubeconfigEnvVar = clientcmd.RecommendedConfigPathEnvVar
	// KubeconfigPath is the path to the Kubernetes client config
	KubeconfigPath = clientcmd.RecommendedHomeFile

	// ImageRepositories contains all known image repositories
	ImageRepositories = map[string][]string{
		"global": {""},
		"cn":     {AliyunMirror},
	}
	// KubernetesReleaseBinaries are Kubernetes release binaries required for
	// kubeadm (kubelet, kubeadm) and the addon manager (kubectl)
	KubernetesReleaseBinaries = []string{"kubelet", "kubeadm", "kubectl"}

	// DefaultNamespaces are Kubernetes namespaces used by minikube, including addons
	DefaultNamespaces = []string{
		"kube-system",
		"kubernetes-dashboard",
		"storage-gluster",
		"istio-operator",
	}

	// ErrMachineMissing is returned when virtual machine does not exist due to user interrupt cancel(i.e. Ctrl + C)
	ErrMachineMissing = errors.New("machine does not exist")

	// KubeadmYamlPath is the path to the kubeadm configuration
	KubeadmYamlPath = path.Join(vmpath.GuestEphemeralDir, "kubeadm.yaml")
)
