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

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	units "github.com/docker/go-units"
	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/util"
	pkgutil "k8s.io/minikube/pkg/util"
)

const (
	isoURL                = "iso-url"
	memory                = "memory"
	cpus                  = "cpus"
	humanReadableDiskSize = "disk-size"
	vmDriver              = "vm-driver"
	kubernetesVersion     = "kubernetes-version"
	hostOnlyCIDR          = "host-only-cidr"
	containerRuntime      = "container-runtime"
	networkPlugin         = "network-plugin"
	hypervVirtualSwitch   = "hyperv-virtual-switch"
	kvmNetwork            = "kvm-network"
	keepContext           = "keep-context"
	featureGates          = "feature-gates"
	apiServerName         = "apiserver-name"
	dnsDomain             = "dns-domain"
)

var (
	registryMirror   []string
	dockerEnv        []string
	dockerOpt        []string
	insecureRegistry []string
	extraOptions     util.ExtraOptionSlice
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a local kubernetes cluster.",
	Long: `Starts a local kubernetes cluster using Virtualbox. This command
assumes you already have Virtualbox installed.`,
	Run: runStart,
}

func runStart(cmd *cobra.Command, args []string) {
	fmt.Printf("Starting local Kubernetes %s cluster...\n", viper.GetString(kubernetesVersion))
	api, err := machine.NewAPIClient(clientType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
		os.Exit(1)
	}
	defer api.Close()

	diskSize := viper.GetString(humanReadableDiskSize)
	diskSizeMB := calculateDiskSizeInMB(diskSize)

	if diskSizeMB < constants.MinimumDiskSizeMB {
		err := fmt.Errorf("Disk Size %dMB (%s) is too small, the minimum disk size is %dMB", diskSizeMB, diskSize, constants.MinimumDiskSizeMB)
		glog.Errorln("Error parsing disk size:", err)
		os.Exit(1)
	}

	config := cluster.MachineConfig{
		MinikubeISO:         viper.GetString(isoURL),
		Memory:              viper.GetInt(memory),
		CPUs:                viper.GetInt(cpus),
		DiskSize:            diskSizeMB,
		VMDriver:            viper.GetString(vmDriver),
		DockerEnv:           dockerEnv,
		DockerOpt:           dockerOpt,
		InsecureRegistry:    insecureRegistry,
		RegistryMirror:      registryMirror,
		HostOnlyCIDR:        viper.GetString(hostOnlyCIDR),
		HypervVirtualSwitch: viper.GetString(hypervVirtualSwitch),
		KvmNetwork:          viper.GetString(kvmNetwork),
		Downloader:          pkgutil.DefaultDownloader{},
	}

	fmt.Println("Starting VM...")
	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(api, config)
		if err != nil {
			glog.Errorf("Error starting host: %s.\n\n Retrying.\n", err)
		}
		return err
	}
	err = util.RetryAfter(5, start, 2*time.Second)
	if err != nil {
		glog.Errorln("Error starting host: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		glog.Errorln("Error starting host: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}
	kubernetesConfig := cluster.KubernetesConfig{
		KubernetesVersion: viper.GetString(kubernetesVersion),
		NodeIP:            ip,
		APIServerName:     viper.GetString(apiServerName),
		DNSDomain:         viper.GetString(dnsDomain),
		FeatureGates:      viper.GetString(featureGates),
		ContainerRuntime:  viper.GetString(containerRuntime),
		NetworkPlugin:     viper.GetString(networkPlugin),
		ExtraOptions:      extraOptions,
	}

	fmt.Println("SSH-ing files into VM...")
	if err := cluster.UpdateCluster(host, host.Driver, kubernetesConfig); err != nil {
		glog.Errorln("Error updating cluster: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	fmt.Println("Setting up certs...")
	if err := cluster.SetupCerts(host.Driver, kubernetesConfig.APIServerName); err != nil {
		glog.Errorln("Error configuring authentication: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	fmt.Println("Starting cluster components...")
	if err := cluster.StartCluster(host, kubernetesConfig); err != nil {
		glog.Errorln("Error starting cluster: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	fmt.Println("Connecting to cluster...")
	kubeHost, err := host.Driver.GetURL()
	if err != nil {
		glog.Errorln("Error connecting to cluster: ", err)
	}
	kubeHost = strings.Replace(kubeHost, "tcp://", "https://", -1)
	kubeHost = strings.Replace(kubeHost, ":2376", ":"+strconv.Itoa(constants.APIServerPort), -1)

	fmt.Println("Setting up kubeconfig...")
	// setup kubeconfig

	kubeConfigEnv := os.Getenv(constants.KubeconfigEnvVar)
	var kubeConfigFile string
	if kubeConfigEnv == "" {
		kubeConfigFile = constants.KubeconfigPath
	} else {
		kubeConfigFile = filepath.SplitList(kubeConfigEnv)[0]
	}

	kubeCfgSetup := &kubeconfig.KubeConfigSetup{
		ClusterName:          constants.MinikubeContext,
		ClusterServerAddress: kubeHost,
		ClientCertificate:    constants.MakeMiniPath("apiserver.crt"),
		ClientKey:            constants.MakeMiniPath("apiserver.key"),
		CertificateAuthority: constants.MakeMiniPath("ca.crt"),
		KeepContext:          viper.GetBool(keepContext),
	}
	kubeCfgSetup.SetKubeConfigFile(kubeConfigFile)

	if err := kubeconfig.SetupKubeConfig(kubeCfgSetup); err != nil {
		glog.Errorln("Error setting up kubeconfig: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	if kubeCfgSetup.KeepContext {
		fmt.Printf("The local Kubernetes cluster has started. The kubectl context has not been altered, kubectl will require \"--context=%s\" to use the local Kubernetes cluster.\n", kubeCfgSetup.ClusterName)
	} else {
		fmt.Println("Kubectl is now configured to use the cluster.")
	}
}

func calculateDiskSizeInMB(humanReadableDiskSize string) int {
	diskSize, err := units.FromHumanSize(humanReadableDiskSize)
	if err != nil {
		glog.Errorf("Invalid disk size: %s", err)
	}
	return int(diskSize / units.MB)
}

func init() {
	startCmd.Flags().Bool(keepContext, constants.DefaultKeepContext, "This will keep the existing kubectl context and will create a minikube context.")
	startCmd.Flags().String(isoURL, constants.DefaultIsoUrl, "Location of the minikube iso")
	startCmd.Flags().String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))
	startCmd.Flags().Int(memory, constants.DefaultMemory, "Amount of RAM allocated to the minikube VM")
	startCmd.Flags().Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the minikube VM")
	startCmd.Flags().String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startCmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minikube VM (only supported with Virtualbox driver)")
	startCmd.Flags().String(hypervVirtualSwitch, "", "The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)")
	startCmd.Flags().String(kvmNetwork, "default", "The KVM network name. (only supported with KVM driver)")
	startCmd.Flags().StringArrayVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().StringArrayVar(&dockerOpt, "docker-opt", nil, "Specify arbitrary flags to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().String(apiServerName, constants.APIServerName, "The apiserver name which is used in the generated certificate for localkube/kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().String(dnsDomain, "", "The cluster dns domain name used in the kubernetes cluster")
	startCmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", nil, "Insecure Docker registries to pass to the Docker daemon")
	startCmd.Flags().StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	startCmd.Flags().String(kubernetesVersion, constants.DefaultKubernetesVersion, "The kubernetes version that the minikube VM will use (ex: v1.2.3) \n OR a URI which contains a localkube binary (ex: https://storage.googleapis.com/minikube/k8sReleases/v1.3.0/localkube-linux-amd64)")
	startCmd.Flags().String(containerRuntime, "", "The container runtime to be used")
	startCmd.Flags().String(networkPlugin, "", "The name of the network plugin")
	startCmd.Flags().String(featureGates, "", "A set of key=value pairs that describe feature gates for alpha/experimental features.")
	startCmd.Flags().Var(&extraOptions, "extra-config",
		`A set of key=value pairs that describe configuration that may be passed to different components.
		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
		Valid components are: kubelet, apiserver, controller-manager, etcd, proxy, scheduler.`)
	viper.BindPFlags(startCmd.Flags())
	RootCmd.AddCommand(startCmd)
}
