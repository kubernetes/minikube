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
	"strconv"
	"strings"

	units "github.com/docker/go-units"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/util"
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
)

var (
	registryMirror   []string
	dockerEnv        []string
	insecureRegistry []string
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
	fmt.Println("Starting local Kubernetes cluster...")
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	config := cluster.MachineConfig{
		MinikubeISO:      viper.GetString(isoURL),
		Memory:           viper.GetInt(memory),
		CPUs:             viper.GetInt(cpus),
		DiskSize:         calculateDiskSizeInMB(viper.GetString(humanReadableDiskSize)),
		VMDriver:         viper.GetString(vmDriver),
		DockerEnv:        dockerEnv,
		InsecureRegistry: insecureRegistry,
		RegistryMirror:   registryMirror,
		HostOnlyCIDR:     viper.GetString(hostOnlyCIDR),
	}

	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(api, config)
		if err != nil {
			glog.Errorf("Error starting host: %s. Retrying.\n", err)
		}
		return err
	}
	err := util.Retry(3, start)
	if err != nil {
		glog.Errorln("Error starting host: ", err)
		util.MaybeReportErrorAndExit(err)
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		glog.Errorln("Error starting host: ", err)
		util.MaybeReportErrorAndExit(err)
	}
	kubernetesConfig := cluster.KubernetesConfig{
		KubernetesVersion: viper.GetString(kubernetesVersion),
		NodeIP:            ip,
		ContainerRuntime:  viper.GetString(containerRuntime),
		NetworkPlugin:     viper.GetString(networkPlugin),
	}
	if err := cluster.UpdateCluster(host, host.Driver, kubernetesConfig); err != nil {
		glog.Errorln("Error updating cluster: ", err)
		util.MaybeReportErrorAndExit(err)
	}

	if err := cluster.SetupCerts(host.Driver); err != nil {
		glog.Errorln("Error configuring authentication: ", err)
		util.MaybeReportErrorAndExit(err)
	}

	if err := cluster.StartCluster(host, kubernetesConfig); err != nil {
		glog.Errorln("Error starting cluster: ", err)
		util.MaybeReportErrorAndExit(err)
	}

	kubeHost, err := host.Driver.GetURL()
	if err != nil {
		glog.Errorln("Error connecting to cluster: ", err)
	}
	kubeHost = strings.Replace(kubeHost, "tcp://", "https://", -1)
	kubeHost = strings.Replace(kubeHost, ":2376", ":"+strconv.Itoa(constants.APIServerPort), -1)

	// setup kubeconfig
	name := constants.MinikubeContext
	certAuth := constants.MakeMiniPath("ca.crt")
	clientCert := constants.MakeMiniPath("apiserver.crt")
	clientKey := constants.MakeMiniPath("apiserver.key")
	if err := setupKubeconfig(name, kubeHost, certAuth, clientCert, clientKey); err != nil {
		glog.Errorln("Error setting up kubeconfig: ", err)
		util.MaybeReportErrorAndExit(err)
	}
	fmt.Println("Kubectl is now configured to use the cluster.")
}

func calculateDiskSizeInMB(humanReadableDiskSize string) int {
	diskSize, err := units.FromHumanSize(humanReadableDiskSize)
	if err != nil {
		glog.Errorf("Invalid disk size: %s", err)
	}
	return int(diskSize / units.MB)
}

// setupKubeconfig reads config from disk, adds the minikube settings, and writes it back.
// activeContext is true when minikube is the CurrentContext
// If no CurrentContext is set, the given name will be used.
func setupKubeconfig(name, server, certAuth, cliCert, cliKey string) error {
	configFile := constants.KubeconfigPath

	// read existing config or create new if does not exist
	config, err := kubeconfig.ReadConfigOrNew(configFile)
	if err != nil {
		return err
	}

	clusterName := name
	cluster := cfg.NewCluster()
	cluster.Server = server
	cluster.CertificateAuthority = certAuth
	config.Clusters[clusterName] = cluster

	// user
	userName := name
	user := cfg.NewAuthInfo()
	user.ClientCertificate = cliCert
	user.ClientKey = cliKey
	config.AuthInfos[userName] = user

	// context
	contextName := name
	context := cfg.NewContext()
	context.Cluster = clusterName
	context.AuthInfo = userName
	config.Contexts[contextName] = context

	// Always set current context to minikube.
	config.CurrentContext = contextName

	// write back to disk
	if err := kubeconfig.WriteConfig(config, configFile); err != nil {
		return err
	}
	return nil
}

func init() {
	startCmd.Flags().String(isoURL, constants.DefaultIsoUrl, "Location of the minikube iso")
	startCmd.Flags().String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))
	startCmd.Flags().Int(memory, constants.DefaultMemory, "Amount of RAM allocated to the minikube VM")
	startCmd.Flags().Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the minikube VM")
	startCmd.Flags().String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startCmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minikube VM (only supported with Virtualbox driver)")
	startCmd.Flags().StringSliceVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", nil, "Insecure Docker registries to pass to the Docker daemon")
	startCmd.Flags().StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	startCmd.Flags().String(kubernetesVersion, constants.DefaultKubernetesVersion, "The kubernetes version that the minikube VM will (ex: v1.2.3) \n OR a URI which contains a localkube binary (ex: https://storage.googleapis.com/minikube/k8sReleases/v1.3.0/localkube-linux-amd64)")
	startCmd.Flags().String(containerRuntime, "", "The container runtime to be used")
	startCmd.Flags().String(networkPlugin, "", "The name of the network plugin")
	viper.BindPFlags(startCmd.Flags())
	RootCmd.AddCommand(startCmd)
}
