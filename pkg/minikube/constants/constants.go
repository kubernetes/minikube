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
	"strings"

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
	"vmwarefusion",
	"kvm",
	"xhyve",
	"hyperv",
	"hyperkit",
	"kvm2",
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
	DefaultStatusFormat = "minikube: {{.MinikubeStatus}}\n" +
		"cluster: {{.ClusterStatus}}\n" + "kubectl: {{.KubeconfigStatus}}\n"
	DefaultAddonListFormat     = "- {{.AddonName}}: {{.AddonStatus}}\n"
	DefaultConfigViewFormat    = "- {{.ConfigKey}}: {{.ConfigValue}}\n"
	DefaultCacheListFormat     = "{{.CacheImage}}\n"
	GithubMinikubeReleasesURL  = "https://storage.googleapis.com/minikube/releases.json"
	KubernetesVersionGCSURL    = "https://storage.googleapis.com/minikube/k8s_releases.json"
	DefaultWait                = 20
	DefaultInterval            = 6
	DefaultClusterBootstrapper = "kubeadm"
)

var DefaultIsoUrl = fmt.Sprintf("https://storage.googleapis.com/%s/minikube-%s.iso", minikubeVersion.GetIsoPath(), minikubeVersion.GetIsoVersion())
var DefaultIsoShaUrl = DefaultIsoUrl + ShaSuffix

var DefaultKubernetesVersion = "v1.10.0"

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
	KubeletServiceFile     = "/lib/systemd/system/kubelet.service"
	KubeletSystemdConfFile = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	KubeadmConfigFile      = "/var/lib/kubeadm.yaml"
)

var Preflights = []string{
	// We use --ignore-preflight-errors=DirAvailable since we have our own custom addons
	// that we also stick in /etc/kubernetes/manifests
	"DirAvailable--etc-kubernetes-manifests",
	"DirAvailable--data-minikube",
	"Port-10250",
	"FileAvailable--etc-kubernetes-manifests-kube-scheduler.yaml",
	"FileAvailable--etc-kubernetes-manifests-kube-apiserver.yaml",
	"FileAvailable--etc-kubernetes-manifests-kube-controller-manager.yaml",
	"FileAvailable--etc-kubernetes-manifests-etcd.yaml",
	// We use --ignore-preflight-errors=Swap since minikube.iso allocates a swap partition.
	// (it should probably stop doing this, though...)
	"Swap",
	// We use --ignore-preflight-errors=CRI since /var/run/dockershim.sock is not present.
	// (because we start kubelet with an invalid config)
	"CRI",
}

const (
	DefaultUfsPort       = "5640"
	DefaultUfsDebugLvl   = 0
	DefaultMountEndpoint = "/minikube-host"
	DefaultMsize         = 262144
	DefaultMountVersion  = "9p2000.u"
)

func GetKubernetesReleaseURL(binaryName, version string) string {
	return fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/amd64/%s", version, binaryName)
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

	gt_v1_10 := semver.MustParseRange(">=1.11.0")
	v1_10 := semver.MustParseRange(">=1.10.0 <1.11.0")
	v1_9 := semver.MustParseRange(">=1.9.0 <1.10.0")
	v1_8 := semver.MustParseRange(">=1.8.0 <1.9.0")

	kubernetesVersion, err := semver.Make(strings.TrimPrefix(kubernetesVersionStr, minikubeVersion.VersionPrefix))
	if err != nil {
		glog.Errorln("Error parsing version semver: ", err)
	}

	if v1_10(kubernetesVersion) || gt_v1_10(kubernetesVersion) {
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
		"k8s.gcr.io/kubernetes-dashboard-amd64:v1.8.1",
		"k8s.gcr.io/kube-addon-manager:v8.6",
		"gcr.io/k8s-minikube/storage-provisioner:v1.8.1",
	}...)

	return images
}

var ImageCacheDir = MakeMiniPath("cache", "images")
