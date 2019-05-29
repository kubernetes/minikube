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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	minikubeVersion "k8s.io/minikube/pkg/version"
)

// APIServerPort is the port that the API server should listen on.
const (
	APIServerName    = "minikubeCA"
	ClusterDNSDomain = "cluster.local"
)

// MinikubeHome is the name of the minikube home directory variable.
const MinikubeHome = "MINIKUBE_HOME"

// GetMinipath returns the path to the user's minikube dir
func GetMinipath() string {
	if os.Getenv(MinikubeHome) == "" {
		return DefaultMinipath
	}
	if filepath.Base(os.Getenv(MinikubeHome)) == ".minikube" {
		return os.Getenv(MinikubeHome)
	}
	return filepath.Join(os.Getenv(MinikubeHome), ".minikube")
}

// ArchTag returns the archtag for images
func ArchTag(hasTag bool) string {
	if runtime.GOARCH == "amd64" && !hasTag {
		return ":"
	}
	return "-" + runtime.GOARCH + ":"
}

// SupportedVMDrivers is a list of supported drivers on all platforms. Currently
// used in gendocs.
var SupportedVMDrivers = [...]string{
	"virtualbox",
	"parallels",
	"vmwarefusion",
	"kvm",
	"xhyve",
	"hyperv",
	"hyperkit",
	"kvm2",
	"vmware",
	"none",
}

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

// TunnelRegistryPath returns the path to the runnel registry file
func TunnelRegistryPath() string {
	return filepath.Join(GetMinipath(), "tunnels.json")
}

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName ...string) string {
	args := []string{GetMinipath()}
	args = append(args, fileName...)
	return filepath.Join(args...)
}

// MountProcessFileName is the filename of the mount process
var MountProcessFileName = ".mount-process"

const (
	// DefaultKeepContext is if we should keep context by default
	DefaultKeepContext = false
	// SHASuffix is the suffix of a SHA-256 checksum file
	SHASuffix = ".sha256"
	// DefaultMemory is the default memory of a host, in megabytes
	DefaultMemory = 2048
	// DefaultCPUS is the default number of cpus of a host
	DefaultCPUS = 2
	// DefaultDiskSize is the default disk image size, parseable
	DefaultDiskSize = "20g"
	// MinimumDiskSizeMB is the minimum disk image size, in megabytes
	MinimumDiskSizeMB = 2000
	// DefaultVMDriver is the default virtual machine driver name
	DefaultVMDriver = "virtualbox"
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
var DefaultKubernetesVersion = "v1.14.2"

// NewestKubernetesVersion is the newest Kubernetes version to test against
var NewestKubernetesVersion = "v1.14.2"

// OldestKubernetesVersion is the oldest Kubernetes version to test against
var OldestKubernetesVersion = "v1.10.13"

// ConfigFilePath is the path of the config directory
var ConfigFilePath = MakeMiniPath("config")

// ConfigFile is the path of the config file
var ConfigFile = MakeMiniPath("config", "config.json")

// GetProfileFile returns the Minikube profile config file
func GetProfileFile(profile string) string {
	return filepath.Join(GetMinipath(), "profiles", profile, "config.json")
}

// DockerAPIVersion is the API version implemented by Docker running in the minikube VM.
const DockerAPIVersion = "1.39"

// ReportingURL is the URL for reporting a minikube error
const ReportingURL = "https://clouderrorreporting.googleapis.com/v1beta1/projects/k8s-minikube/events:report?key=AIzaSyACUwzG0dEPcl-eOgpDKnyKoUFgHdfoFuA"

// AddonsPath is the default path of the addons configuration
const AddonsPath = "/etc/kubernetes/addons"

// FilesPath is the default path of files
const FilesPath = "/files"

const (
	// KubeletServiceFile is the path to the kubelet systemd service
	KubeletServiceFile = "/lib/systemd/system/kubelet.service"
	// KubeletSystemdConfFile is the path to the kubelet systemd configuration
	KubeletSystemdConfFile = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	// KubeadmConfigFile is the path to the kubeadm configuration
	KubeadmConfigFile = "/var/lib/kubeadm.yaml"
	// DefaultCNIConfigPath is the path to the CNI configuration
	DefaultCNIConfigPath = "/etc/cni/net.d/k8s.conf"
	// DefaultRktNetConfigPath is the path to the rkt net configuration
	DefaultRktNetConfigPath = "/etc/rkt/net.d/k8s.conf"
)

const (
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
)

// ImageRepositories contains all known image repositories
var ImageRepositories = map[string][]string{
	"global": {""},
	"cn":     {"registry.cn-hangzhou.aliyuncs.com/google_containers"},
}

// GetKubernetesReleaseURL gets the location of a kubernetes client
func GetKubernetesReleaseURL(binaryName, version, osName, archName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/%s", version, osName, archName, binaryName)
}

// GetKubernetesReleaseURLSHA1 gets the location of a kubernetes client checksum
func GetKubernetesReleaseURLSHA1(binaryName, version, osName, archName string) string {
	return fmt.Sprintf("%s.sha1", GetKubernetesReleaseURL(binaryName, version, osName, archName))
}

// IsMinikubeChildProcess is the name of "is minikube child process" variable
const IsMinikubeChildProcess = "IS_MINIKUBE_CHILD_PROCESS"

// DriverNone is the none driver
const DriverNone = "none"

// FileScheme is the file scheme
const FileScheme = "file"

// GetKubeadmCachedBinaries gets the binaries to cache for kubeadm
func GetKubeadmCachedBinaries() []string {
	return []string{"kubelet", "kubeadm"}
}

// GetKubeadmCachedImages gets the images to cache for kubeadm for a version
func GetKubeadmCachedImages(imageRepository string, kubernetesVersionStr string) (string, []string) {
	minikubeRepository := imageRepository
	if imageRepository == "" {
		imageRepository = "k8s.gcr.io"
		minikubeRepository = "gcr.io/k8s-minikube"
	}
	if !strings.HasSuffix(imageRepository, "/") {
		imageRepository += "/"
	}
	if !strings.HasSuffix(minikubeRepository, "/") {
		minikubeRepository += "/"
	}

	v1_14plus := semver.MustParseRange(">=1.14.0")
	v1_13 := semver.MustParseRange(">=1.13.0 <1.14.0")
	v1_12 := semver.MustParseRange(">=1.12.0 <1.13.0")
	v1_11 := semver.MustParseRange(">=1.11.0 <1.12.0")
	v1_10 := semver.MustParseRange(">=1.10.0 <1.11.0")
	v1_9 := semver.MustParseRange(">=1.9.0 <1.10.0")
	v1_8 := semver.MustParseRange(">=1.8.0 <1.9.0")
	v1_12plus := semver.MustParseRange(">=1.12.0")

	kubernetesVersion, err := semver.Make(strings.TrimPrefix(kubernetesVersionStr, minikubeVersion.VersionPrefix))
	if err != nil {
		glog.Errorln("Error parsing version semver: ", err)
	}

	var images []string
	if v1_12plus(kubernetesVersion) {
		images = append(images, []string{
			imageRepository + "kube-proxy" + ArchTag(false) + kubernetesVersionStr,
			imageRepository + "kube-scheduler" + ArchTag(false) + kubernetesVersionStr,
			imageRepository + "kube-controller-manager" + ArchTag(false) + kubernetesVersionStr,
			imageRepository + "kube-apiserver" + ArchTag(false) + kubernetesVersionStr,
		}...)
	} else {
		images = append(images, []string{
			imageRepository + "kube-proxy" + ArchTag(true) + kubernetesVersionStr,
			imageRepository + "kube-scheduler" + ArchTag(true) + kubernetesVersionStr,
			imageRepository + "kube-controller-manager" + ArchTag(true) + kubernetesVersionStr,
			imageRepository + "kube-apiserver" + ArchTag(true) + kubernetesVersionStr,
		}...)
	}

	var podInfraContainerImage string
	if v1_14plus(kubernetesVersion) {
		podInfraContainerImage = imageRepository + "pause:3.1"
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.13",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.13",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.13",
			imageRepository + "etcd" + ArchTag(false) + "3.3.10",
			imageRepository + "coredns" + ArchTag(false) + "1.3.1",
		}...)

	} else if v1_13(kubernetesVersion) {
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.1"
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.8",
			imageRepository + "etcd" + ArchTag(false) + "3.2.24",
			imageRepository + "coredns:1.2.6",
		}...)

	} else if v1_12(kubernetesVersion) {
		podInfraContainerImage = imageRepository + "pause:3.1"
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.8",
			imageRepository + "etcd" + ArchTag(false) + "3.2.24",
			imageRepository + "coredns:1.2.2",
		}...)

	} else if v1_11(kubernetesVersion) {
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.1"
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.8",
			imageRepository + "etcd" + ArchTag(true) + "3.2.18",
			imageRepository + "coredns:1.1.3",
		}...)

	} else if v1_10(kubernetesVersion) {
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.1"
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.8",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.8",
			imageRepository + "etcd" + ArchTag(true) + "3.1.12",
		}...)

	} else if v1_9(kubernetesVersion) {
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.0"
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.7",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.7",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.7",
			imageRepository + "etcd" + ArchTag(true) + "3.1.10",
		}...)

	} else if v1_8(kubernetesVersion) {
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.0"
		images = append(images, []string{
			podInfraContainerImage,
			imageRepository + "k8s-dns-kube-dns" + ArchTag(true) + "1.14.5",
			imageRepository + "k8s-dns-dnsmasq-nanny" + ArchTag(true) + "1.14.5",
			imageRepository + "k8s-dns-sidecar" + ArchTag(true) + "1.14.5",
			imageRepository + "etcd" + ArchTag(true) + "3.0.17",
		}...)

	} else {
		podInfraContainerImage = imageRepository + "pause" + ArchTag(false) + "3.0"
	}

	images = append(images, []string{
		imageRepository + "kubernetes-dashboard" + ArchTag(true) + "v1.10.1",
		imageRepository + "kube-addon-manager" + ArchTag(false) + "v9.0",
		minikubeRepository + "storage-provisioner" + ArchTag(false) + "v1.8.1",
	}...)

	return podInfraContainerImage, images
}

// ImageCacheDir is the path to the image cache directory
var ImageCacheDir = MakeMiniPath("cache", "images")

const (
	// GvisorFilesPath is the path to the gvisor files saved by go-bindata
	GvisorFilesPath = "/tmp/gvisor"
	// ContainerdConfigTomlPath is the path to the containerd config.toml
	ContainerdConfigTomlPath = "/etc/containerd/config.toml"
	// GvisorContainerdShimTomlPath is the path to gvisor-containerd-shim.toml
	GvisorContainerdShimTomlPath = "/etc/containerd/gvisor-containerd-shim.toml"
	// StoredContainerdConfigTomlPath is the path where the default config.toml will be stored
	StoredContainerdConfigTomlPath = "/tmp/config.toml"

	//GvisorConfigTomlTargetName is the go-bindata target name for the gvisor config.toml
	GvisorConfigTomlTargetName = "gvisor-config.toml"
	// GvisorContainerdShimTargetName is the go-bindata target name for gvisor-containerd-shim
	GvisorContainerdShimTargetName = "gvisor-containerd-shim.toml"

	// GvisorContainerdShimURL is the url to download gvisor-containerd-shim
	GvisorContainerdShimURL = "https://github.com/google/gvisor-containerd-shim/releases/download/v0.0.1-rc.0/gvisor-containerd-shim-v0.0.1-rc.0.linux-amd64"
	// GvisorURL is the url to download gvisor
	GvisorURL = "https://storage.googleapis.com/gvisor/releases/nightly/2018-12-07/runsc"
)
