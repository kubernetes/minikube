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

const MinikubeHome = "MINIKUBE_HOME"

// Minipath is the path to the user's minikube dir
func GetMinipath() string {
	if os.Getenv(MinikubeHome) == "" {
		return DefaultMinipath
	}
	if filepath.Base(os.Getenv(MinikubeHome)) == ".minikube" {
		return os.Getenv(MinikubeHome)
	}
	return filepath.Join(os.Getenv(MinikubeHome), ".minikube")
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

// The name of the default storage class provisioner
const DefaultStorageClassProvisioner = "standard"

// Used to modify the cache field in the config file
const Cache = "cache"

func TunnelRegistryPath() string {
	return filepath.Join(GetMinipath(), "tunnels.json")
}

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName ...string) string {
	args := []string{GetMinipath()}
	args = append(args, fileName...)
	return filepath.Join(args...)
}

var MountProcessFileName = ".mount-process"

const (
	DefaultKeepContext  = false
	ShaSuffix           = ".sha256"
	DefaultMemory       = 2048
	DefaultCPUS         = 2
	DefaultDiskSize     = "20g"
	MinimumDiskSizeMB   = 2000
	DefaultVMDriver     = "virtualbox"
	DefaultStatusFormat = `host: {{.Host}}
kubelet: {{.Kubelet}}
apiserver: {{.ApiServer}}
kubectl: {{.Kubeconfig}}
`
	DefaultAddonListFormat     = "- {{.AddonName}}: {{.AddonStatus}}\n"
	DefaultConfigViewFormat    = "- {{.ConfigKey}}: {{.ConfigValue}}\n"
	DefaultCacheListFormat     = "{{.CacheImage}}\n"
	GithubMinikubeReleasesURL  = "https://storage.googleapis.com/minikube/releases.json"
	DefaultWait                = 20
	DefaultInterval            = 6
	DefaultK8sClientTimeout    = 60 * time.Second
	DefaultClusterBootstrapper = "kubeadm"
)

var DefaultIsoUrl = fmt.Sprintf("https://storage.googleapis.com/%s/minikube-%s.iso", minikubeVersion.GetIsoPath(), minikubeVersion.GetIsoVersion())
var DefaultIsoShaUrl = DefaultIsoUrl + ShaSuffix

var DefaultKubernetesVersion = "v1.13.3"

var ConfigFilePath = MakeMiniPath("config")
var ConfigFile = MakeMiniPath("config", "config.json")

// GetProfileFile returns the Minikube profile config file
func GetProfileFile(profile string) string {
	return filepath.Join(GetMinipath(), "profiles", profile, "config.json")
}

// DockerAPIVersion is the API version implemented by Docker running in the minikube VM.
const DockerAPIVersion = "1.35"

const ReportingURL = "https://clouderrorreporting.googleapis.com/v1beta1/projects/k8s-minikube/events:report?key=AIzaSyACUwzG0dEPcl-eOgpDKnyKoUFgHdfoFuA"

const AddonsPath = "/etc/kubernetes/addons"
const FilesPath = "/files"

const (
	KubeletServiceFile      = "/lib/systemd/system/kubelet.service"
	KubeletSystemdConfFile  = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	KubeadmConfigFile       = "/var/lib/kubeadm.yaml"
	DefaultCNIConfigPath    = "/etc/cni/net.d/k8s.conf"
	DefaultRktNetConfigPath = "/etc/rkt/net.d/k8s.conf"
)

const (
	DefaultUfsPort       = "5640"
	DefaultUfsDebugLvl   = 0
	DefaultMountEndpoint = "/minikube-host"
	DefaultMsize         = 262144
	DefaultMountVersion  = "9p2000.u"
)

func GetKubernetesReleaseURL(binaryName, version string) string {
	return fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/%s", version, runtime.GOARCH, binaryName)
}

func GetKubernetesReleaseURLSha1(binaryName, version string) string {
	return fmt.Sprintf("%s.sha1", GetKubernetesReleaseURL(binaryName, version))
}

const IsMinikubeChildProcess = "IS_MINIKUBE_CHILD_PROCESS"
const DriverNone = "none"
const FileScheme = "file"

func GetKubeadmCachedImages(kubernetesVersionStr string) []string {

	var images = []string{
		"k8s.gcr.io/kube-proxy-amd64:" + kubernetesVersionStr,
		"k8s.gcr.io/kube-scheduler-amd64:" + kubernetesVersionStr,
		"k8s.gcr.io/kube-controller-manager-amd64:" + kubernetesVersionStr,
		"k8s.gcr.io/kube-apiserver-amd64:" + kubernetesVersionStr,
	}

	gt_v1_11 := semver.MustParseRange(">=1.12.0")
	v1_11 := semver.MustParseRange(">=1.11.0 <1.12.0")
	v1_10 := semver.MustParseRange(">=1.10.0 <1.11.0")
	v1_9 := semver.MustParseRange(">=1.9.0 <1.10.0")
	v1_8 := semver.MustParseRange(">=1.8.0 <1.9.0")

	kubernetesVersion, err := semver.Make(strings.TrimPrefix(kubernetesVersionStr, minikubeVersion.VersionPrefix))
	if err != nil {
		glog.Errorln("Error parsing version semver: ", err)
	}

	if v1_11(kubernetesVersion) || gt_v1_11(kubernetesVersion) {
		images = append(images, []string{
			"k8s.gcr.io/pause-amd64:3.1",
			"k8s.gcr.io/pause:3.1",
			"k8s.gcr.io/k8s-dns-kube-dns-amd64:1.14.8",
			"k8s.gcr.io/k8s-dns-dnsmasq-nanny-amd64:1.14.8",
			"k8s.gcr.io/k8s-dns-sidecar-amd64:1.14.8",
			"k8s.gcr.io/etcd-amd64:3.2.18",
			"k8s.gcr.io/coredns:1.1.3",
			"k8s.gcr.io/coredns:1.2.2",
		}...)

	} else if v1_10(kubernetesVersion) {
		images = append(images, []string{
			"k8s.gcr.io/pause-amd64:3.1",
			"k8s.gcr.io/k8s-dns-kube-dns-amd64:1.14.8",
			"k8s.gcr.io/k8s-dns-dnsmasq-nanny-amd64:1.14.8",
			"k8s.gcr.io/k8s-dns-sidecar-amd64:1.14.8",
			"k8s.gcr.io/etcd-amd64:3.1.12",
		}...)

	} else if v1_9(kubernetesVersion) {
		images = append(images, []string{
			"k8s.gcr.io/pause-amd64:3.0",
			"k8s.gcr.io/k8s-dns-kube-dns-amd64:1.14.7",
			"k8s.gcr.io/k8s-dns-dnsmasq-nanny-amd64:1.14.7",
			"k8s.gcr.io/k8s-dns-sidecar-amd64:1.14.7",
			"k8s.gcr.io/etcd-amd64:3.1.10",
		}...)

	} else if v1_8(kubernetesVersion) {
		images = append(images, []string{
			"k8s.gcr.io/pause-amd64:3.0",
			"k8s.gcr.io/k8s-dns-kube-dns-amd64:1.14.5",
			"k8s.gcr.io/k8s-dns-dnsmasq-nanny-amd64:1.14.5",
			"k8s.gcr.io/k8s-dns-sidecar-amd64:1.14.5",
			"k8s.gcr.io/etcd-amd64:3.0.17",
		}...)
	}

	images = append(images, []string{
		"k8s.gcr.io/kubernetes-dashboard-amd64:v1.10.1",
		"k8s.gcr.io/kube-addon-manager:v8.6",
		"gcr.io/k8s-minikube/storage-provisioner:v1.8.1",
	}...)

	return images
}

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
