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

	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/util/homedir"
	"k8s.io/kubernetes/pkg/version"
	minikubeVersion "k8s.io/minikube/pkg/version"
)

// MachineName is the name to use for the VM.
const MachineName = "minikube"

// APIServerPort is the port that the API server should listen on.
const APIServerPort = 8443

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

var DefaultMinipath = filepath.Join(homedir.HomeDir(), ".minikube")

// KubeconfigPath is the path to the Kubernetes client config
var KubeconfigPath = clientcmd.RecommendedHomeFile

// KubeconfigEnvVar is the env var to check for the Kubernetes client config
var KubeconfigEnvVar = clientcmd.RecommendedConfigPathEnvVar

// MinikubeContext is the kubeconfig context name used for minikube
const MinikubeContext = "minikube"

// MinikubeEnvPrefix is the prefix for the environmental variables
const MinikubeEnvPrefix = "MINIKUBE"

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName ...string) string {
	args := []string{GetMinipath()}
	args = append(args, fileName...)
	return filepath.Join(args...)
}

// Only pass along these flags to localkube.
var LogFlags = [...]string{
	"v",
	"vmodule",
}

const (
	DefaultKeepContext  = false
	ShaSuffix           = ".sha256"
	DefaultMemory       = 2048
	DefaultCPUS         = 2
	DefaultDiskSize     = "20g"
	MinimumDiskSizeMB   = 2000
	DefaultVMDriver     = "virtualbox"
	DefaultStatusFormat = "minikubeVM: {{.MinikubeStatus}}\n" +
		"localkube: {{.LocalkubeStatus}}\n"
	DefaultAddonListFormat    = "- {{.AddonName}}: {{.AddonStatus}}\n"
	DefaultConfigViewFormat   = "- {{.ConfigKey}}: {{.ConfigValue}}\n"
	GithubMinikubeReleasesURL = "https://storage.googleapis.com/minikube/releases.json"
	KubernetesVersionGCSURL   = "https://storage.googleapis.com/minikube/k8s_releases.json"
)

var DefaultIsoUrl = fmt.Sprintf("https://storage.googleapis.com/%s/minikube-%s.iso", minikubeVersion.GetIsoPath(), minikubeVersion.GetIsoVersion())
var DefaultIsoShaUrl = DefaultIsoUrl + ShaSuffix

var DefaultKubernetesVersion = version.Get().GitVersion

var ConfigFilePath = MakeMiniPath("config")
var ConfigFile = MakeMiniPath("config", "config.json")

var LocalkubeDownloadURLPrefix = "https://storage.googleapis.com/minikube/k8sReleases/"
var LocalkubeLinuxFilename = "localkube-linux-amd64"

// DockerAPIVersion is the API version implemented by Docker running in the minikube VM.
const DockerAPIVersion = "1.23"

const ReportingURL = "https://clouderrorreporting.googleapis.com/v1beta1/projects/k8s-minikube/events:report?key=AIzaSyACUwzG0dEPcl-eOgpDKnyKoUFgHdfoFuA"

const AddonsPath = "/etc/kubernetes/addons"

const (
	RemoteLocalKubeErrPath = "/var/lib/localkube/localkube.err"
	RemoteLocalKubeOutPath = "/var/lib/localkube/localkube.out"
	LocalkubePIDPath       = "/var/run/localkube.pid"
)

const (
	LocalkubeServicePath = "/usr/lib/systemd/system/localkube.service"
	LocalkubeRunning     = "active"
	LocalkubeStopped     = "inactive"
)

const (
	DefaultUfsAddress  = ":5640"
	DefaultUfsDebugLvl = 0
)
