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
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/minikube/pkg/minikube/localpath"
	minikubeVersion "k8s.io/minikube/pkg/version"
)

const (
	APIServerPort    = 8443
	APIServerName    = "minikubeCA"
	ClusterDNSDomain = "cluster.local"
)

// DriverMock is a mock driver.
const DriverMock = "mock-driver"

// DriverNone is the none driver.
const DriverNone = "none"

// DriverKvm2 is the kvm2 driver option name for in linux
const DriverKvm2 = "kvm2"

// DriverVirtualbox is the virtualbox driver option name
const DriverVirtualbox = "virtualbox"

// DriverHyperkit is the hyperkit driver option name for mac os
const DriverHyperkit = "hyperkit"

// DriverVmware is the vmware driver option name
const DriverVmware = "vmware"

// DriverVmwareFusion is the vmware fusion driver option
const DriverVmwareFusion = "vmwarefusion"

// DriverHyperv is the hyperv driver option for windows
const DriverHyperv = "hyperv"

// DriverParallels is the parallels driver option name
const DriverParallels = "parallels"

// DefaultMinipath is the default Minikube path (under the home directory)
var DefaultMinipath = filepath.Join(homedir.HomeDir(), ".minikube")

// KubeconfigPath is the path to the Kubernetes client config
var KubeconfigPath = clientcmd.RecommendedHomeFile

// KubeconfigEnvVar is the env var to check for the Kubernetes client config
var KubeconfigEnvVar = clientcmd.RecommendedConfigPathEnvVar

// MinikubeContext is the kubeconfig context name used for minikube
const MinikubeContext = "minikube"

// MinikubeEnvPrefix is the prefix for the environmental variables
const MinikubeEnvPrefix = "MINIKUBE"

// DefaultMachineName is the default name for the VM
const DefaultMachineName = "minikube"

// DefaultNodeName is the default name for the kubeadm node within the VM
const DefaultNodeName = "minikube"

// DefaultStorageClassProvisioner is the name of the default storage class provisioner
const DefaultStorageClassProvisioner = "standard"

// Cache is used to modify the cache field in the config file
const Cache = "cache"

// MountProcessFileName is the filename of the mount process
var MountProcessFileName = ".mount-process"

const (
	// DefaultEmbedCerts  is if the certs should be embedded in the kubeconfig file
	DefaultEmbedCerts = false
	// DefaultKeepContext is if we should keep context by default
	DefaultKeepContext = false
	// SHASuffix is the suffix of a SHA-256 checksum file
	SHASuffix = ".sha256"
	// DefaultMemorySize is the default memory which will be allocated to minikube, in megabytes
	DefaultMemorySize = "2000mb"
	// MinimumMemorySize is the minimum memory size, in megabytes
	MinimumMemorySize = "1024mb"
	// DefaultCPUS is the default number of cpus of a host
	DefaultCPUS = 2
	// MinimumCPUS is the minimum number of cpus of a host
	MinimumCPUS = 2
	// DefaultDiskSize is the default disk image size, in megabytes
	DefaultDiskSize = "20000mb"
	// MinimumDiskSize is the minimum disk image size, in megabytes
	MinimumDiskSize = "2000mb"
	// DefaultVMDriver is the default virtual machine driver name
	DefaultVMDriver = DriverVirtualbox
	// DefaultStatusFormat is the default format of a host
	DefaultStatusFormat = `host: {{.Host}}
kubelet: {{.Kubelet}}
apiserver: {{.APIServer}}
kubectl: {{.Kubeconfig}}
`
	// DefaultAddonListFormat is the default format of addon list
	DefaultAddonListFormat = "- {{.AddonName}}: {{.AddonStatus}}\n"
	// DefaultConfigViewFormat is the default format of config view
	DefaultConfigViewFormat = "- {{.ConfigKey}}: {{.ConfigValue}}\n"
	// DefaultCacheListFormat is the default format of cache list
	DefaultCacheListFormat = "{{.CacheImage}}\n"
	// GithubMinikubeReleasesURL is the URL of the minikube github releases JSON file
	GithubMinikubeReleasesURL = "https://storage.googleapis.com/minikube/releases.json"
	// DefaultWait is the default wait time, in seconds
	DefaultWait = 20
	// DefaultInterval is the default interval, in seconds
	DefaultInterval = 6
	// DefaultK8sClientTimeout is the default kubernetes client timeout
	DefaultK8sClientTimeout = 60 * time.Second
	// DefaultClusterBootstrapper is the default cluster bootstrapper
	DefaultClusterBootstrapper = "kubeadm"
)

// DefaultISOURL is the default location of the minikube.iso file
var DefaultISOURL = fmt.Sprintf("https://storage.googleapis.com/%s/minikube-%s.iso", minikubeVersion.GetISOPath(), minikubeVersion.GetISOVersion())

// DefaultISOSHAURL is the default location of the minikube.iso.sha256 file
var DefaultISOSHAURL = DefaultISOURL + SHASuffix

// DefaultKubernetesVersion is the default kubernetes version
var DefaultKubernetesVersion = "v1.16.0"

// NewestKubernetesVersion is the newest Kubernetes version to test against
var NewestKubernetesVersion = "v1.16.0"

// OldestKubernetesVersion is the oldest Kubernetes version to test against
var OldestKubernetesVersion = "v1.11.10"

// ConfigFile is the path of the config file
var ConfigFile = localpath.MakeMiniPath("config", "config.json")

const (
	// KubeletServiceFile is the path to the kubelet systemd service
	KubeletServiceFile = "/lib/systemd/system/kubelet.service"
	// KubeletSystemdConfFile is the path to the kubelet systemd configuration
	KubeletSystemdConfFile = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	// DefaultCNIConfigPath is the path to the CNI configuration
	DefaultCNIConfigPath = "/etc/cni/net.d/k8s.conf"

	// GuestAddonsDir is the default path of the addons configuration
	GuestAddonsDir = "/etc/kubernetes/addons"
	// GuestManifestsDir is where the kubelet should look for static Pod manifests
	GuestManifestsDir = "/etc/kubernetes/manifests"
	// GuestEphemeralDir is the path where ephemeral data should be stored within the VM
	GuestEphemeralDir = "/var/tmp/minikube"
	// PersistentDir is the path where persistent data should be stored within the VM (not tmpfs)
	GuestPersistentDir = "/var/lib/minikube"
	// GuestCertsDir are where Kubernetes certificates are kept on the guest
	GuestCertsDir = GuestPersistentDir + "/certs"
	// DefaultUfsPort is the default port of UFS
	DefaultUfsPort = "5640"
	// DefaultUfsDebugLvl is the default debug level of UFS
	DefaultUfsDebugLvl = 0
	// DefaultMountEndpoint is the default mount endpoint
	DefaultMountEndpoint = "/minikube-host"
	// DefaultMsize is the default number of bytes to use for 9p packet payload
	DefaultMsize = 262144
	// DefaultMountVersion is the default 9p version to use for mount
	DefaultMountVersion = "9p2000.L"

	// IsMinikubeChildProcess is the name of "is minikube child process" variable
	IsMinikubeChildProcess = "IS_MINIKUBE_CHILD_PROCESS"
	// FileScheme is the file scheme
	FileScheme = "file"
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
	// ContainerdConfigTomlPath is the path to the containerd config.toml
	ContainerdConfigTomlPath = "/etc/containerd/config.toml"
	// StoredContainerdConfigTomlPath is the path where the default config.toml will be stored
	StoredContainerdConfigTomlPath = "/tmp/config.toml"

	// GvisorConfigTomlTargetName is the go-bindata target name for the gvisor config.toml
	GvisorConfigTomlTargetName = "gvisor-config.toml"

	// GvisorContainerdShimURL is the url to download gvisor-containerd-shim
	GvisorContainerdShimURL = "https://github.com/google/gvisor-containerd-shim/releases/download/v0.0.3/containerd-shim-runsc-v1.linux-amd64"
	// GvisorURL is the url to download gvisor
	GvisorURL = "https://storage.googleapis.com/gvisor/releases/nightly/2019-01-14/runsc"
)
