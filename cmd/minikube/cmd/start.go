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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	units "github.com/docker/go-units"
	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/kubernetes_versions"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/util"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/kubeconfig"
	"k8s.io/minikube/pkg/version"
)

const (
	isoURL                = "iso-url"
	memory                = "memory"
	cpus                  = "cpus"
	humanReadableDiskSize = "disk-size"
	vmDriver              = "vm-driver"
	xhyveDiskDriver       = "xhyve-disk-driver"
	kubernetesVersion     = "kubernetes-version"
	hostOnlyCIDR          = "host-only-cidr"
	containerRuntime      = "container-runtime"
	networkPlugin         = "network-plugin"
	hypervVirtualSwitch   = "hyperv-virtual-switch"
	kvmNetwork            = "kvm-network"
	keepContext           = "keep-context"
	createMount           = "mount"
	featureGates          = "feature-gates"
	apiServerName         = "apiserver-name"
	dnsDomain             = "dns-domain"
	mountString           = "mount-string"
	disableDriverMounts   = "disable-driver-mounts"
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
	Short: "Starts a local kubernetes cluster",
	Long: `Starts a local kubernetes cluster using VM. This command
assumes you have already installed one of the VM drivers: virtualbox/vmwarefusion/kvm/xhyve/hyperv.`,
	Run: runStart,
}

func runStart(cmd *cobra.Command, args []string) {
	api, err := machine.NewAPIClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
		os.Exit(1)
	}
	defer api.Close()

	// Load default minikube config
	minikubeConfig := getDefaultConfig()

	// Override with global config
	globalConfig, err := cfg.ReadConfig()
	if err != nil {
		glog.Errorln("Error reading global config:", err)
	}

	for k, v := range globalConfig {
		minikubeConfig[k] = v
	}

	// Override with profile config
	profileConfig, err := ReadProfileConfig(viper.GetString(cfg.MachineProfile))
	if err != nil {
		glog.Errorln("Error reading profile config:", err)
	}

	for k, v := range profileConfig {
		minikubeConfig[k] = v
	}

	// Override environment variables

	// Override with flags

	// Generate Machine and Kubernetes configs
	machineConfig := loadMachineConfig(cluster.MachineConfig{}, minikubeConfig)
	kubernetesConfig := loadKubernetesConfig(cluster.KubernetesConfig{}, minikubeConfig)

	if machineConfig.DiskSize < constants.MinimumDiskSizeMB {
		err := fmt.Errorf("Disk Size %dMB (%s) is too small, the minimum disk size is %dMB", machineConfig.DiskSize, minikubeConfig[humanReadableDiskSize].(string), constants.MinimumDiskSizeMB)
		glog.Errorln("Error parsing disk size:", err)
		os.Exit(1)
	}

	if kubernetesConfig.KubernetesVersion != constants.DefaultKubernetesVersion {
		validateK8sVersion(kubernetesConfig.KubernetesVersion)
	}

	machineConfig.Downloader = pkgutil.DefaultDownloader{}
	machineConfig.DockerEnv = dockerEnv
	machineConfig.DockerOpt = dockerOpt
	machineConfig.InsecureRegistry = insecureRegistry
	machineConfig.RegistryMirror = registryMirror

	fmt.Printf("Starting local Kubernetes %s cluster...\n", kubernetesConfig.KubernetesVersion)
	fmt.Println("Starting VM...")
	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(api, machineConfig)
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

	fmt.Println("Getting VM IP address...")
	ip, err := host.Driver.GetIP()
	if err != nil {
		glog.Errorln("Error getting VM IP address: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	kubernetesConfig.NodeIP = ip

	selectedKubernetesVersion := kubernetesConfig.KubernetesVersion

	// Only if kubernetes version is found in profile config, perform version checks
	if _, ok := profileConfig[kubernetesVersion]; ok {
		oldKubernetesVersion, err := semver.Make(strings.TrimPrefix(profileConfig[kubernetesVersion].(string), version.VersionPrefix))
		if err != nil {
			glog.Errorln("Error parsing version semver: ", err)
		}

		newKubernetesVersion, err := semver.Make(strings.TrimPrefix(kubernetesConfig.KubernetesVersion, version.VersionPrefix))
		if err != nil {
			glog.Errorln("Error parsing version semver: ", err)
		}

		// Check if it's an attempt to downgrade version. Avoid version downgrad.
		if newKubernetesVersion.LT(oldKubernetesVersion) {
			selectedKubernetesVersion = version.VersionPrefix + oldKubernetesVersion.String()
			fmt.Println("Kubernetes version downgrade is not supported. Using version:", selectedKubernetesVersion)
		}
	}

	kubernetesConfig.KubernetesVersion = selectedKubernetesVersion
	profileConfig[kubernetesVersion] = selectedKubernetesVersion

	// Write the updated profile to config file
	WriteProfileConfig(profileConfig, viper.GetString(cfg.MachineProfile))

	fmt.Println("Moving files into cluster...")
	if err := cluster.UpdateCluster(host.Driver, kubernetesConfig); err != nil {
		glog.Errorln("Error updating cluster: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	fmt.Println("Setting up certs...")
	if err := cluster.SetupCerts(host.Driver, kubernetesConfig.APIServerName, kubernetesConfig.DNSDomain); err != nil {
		glog.Errorln("Error configuring authentication: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	fmt.Println("Starting cluster components...")

	if err := cluster.StartCluster(api, kubernetesConfig); err != nil {
		glog.Errorln("Error starting cluster: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	fmt.Println("Connecting to cluster...")
	kubeHost, err := host.Driver.GetURL()
	if err != nil {
		glog.Errorln("Error connecting to cluster: ", err)
	}
	kubeHost = strings.Replace(kubeHost, "tcp://", "https://", -1)
	kubeHost = strings.Replace(kubeHost, ":2376", ":"+strconv.Itoa(pkgutil.APIServerPort), -1)

	fmt.Println("Setting up kubeconfig...")
	// setup kubeconfig

	kubeConfigFile := cmdUtil.GetKubeConfigPath()

	kubeCfgSetup := &kubeconfig.KubeConfigSetup{
		ClusterName:          cfg.GetMachineName(),
		ClusterServerAddress: kubeHost,
		ClientCertificate:    constants.MakeMiniPath("apiserver.crt"),
		ClientKey:            constants.MakeMiniPath("apiserver.key"),
		CertificateAuthority: constants.MakeMiniPath("ca.crt"),
		KeepContext:          minikubeConfig["keep-context"].(bool),
	}
	kubeCfgSetup.SetKubeConfigFile(kubeConfigFile)

	if err := kubeconfig.SetupKubeConfig(kubeCfgSetup); err != nil {
		glog.Errorln("Error setting up kubeconfig: ", err)
		cmdUtil.MaybeReportErrorAndExit(err)
	}

	// start 9p server mount
	if minikubeConfig[createMount].(bool) {
		mntString := minikubeConfig[mountString].(string)
		fmt.Printf("Setting up hostmount on %s...\n", mntString)

		path := os.Args[0]
		mountDebugVal := 0
		if glog.V(8) {
			mountDebugVal = 1
		}
		mountCmd := exec.Command(path, "mount", fmt.Sprintf("--v=%d", mountDebugVal), mntString)
		mountCmd.Env = append(os.Environ(), constants.IsMinikubeChildProcess+"=true")
		if glog.V(8) {
			mountCmd.Stdout = os.Stdout
			mountCmd.Stderr = os.Stderr
		}
		err = mountCmd.Start()
		if err != nil {
			glog.Errorf("Error running command minikube mount %s", err)
			cmdUtil.MaybeReportErrorAndExit(err)
		}
		err = ioutil.WriteFile(filepath.Join(constants.GetMinipath(), constants.MountProcessFileName), []byte(strconv.Itoa(mountCmd.Process.Pid)), 0644)
		if err != nil {
			glog.Errorf("Error writing mount process pid to file: %s", err)
			cmdUtil.MaybeReportErrorAndExit(err)
		}
	}

	if kubeCfgSetup.KeepContext {
		fmt.Printf("The local Kubernetes cluster has started. The kubectl context has not been altered, kubectl will require \"--context=%s\" to use the local Kubernetes cluster.\n",
			kubeCfgSetup.ClusterName)
	} else {
		fmt.Println("Kubectl is now configured to use the cluster.")
	}

	if machineConfig.VMDriver == "none" {
		fmt.Println(`===================
WARNING: IT IS RECOMMENDED NOT TO RUN THE NONE DRIVER ON PERSONAL WORKSTATIONS
	The 'none' driver will run an insecure kubernetes apiserver as root that may leave the host vulnerable to CSRF attacks
`)

		if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
			fmt.Println(`When using the none driver, the kubectl config and credentials generated will be root owned and will appear in the root home directory.
You will need to move the files to the appropriate location and then set the correct permissions.  An example of this is below:
	sudo mv /root/.kube $HOME/.kube # this will overwrite any config you have.  You may have to append the file contents manually
	sudo chown -R $USER $HOME/.kube
	sudo chgrp -R $USER $HOME/.kube
	
    sudo mv /root/.minikube $HOME/.minikube # this will overwrite any config you have.  You may have to append the file contents manually
	sudo chown -R $USER $HOME/.minikube
	sudo chgrp -R $USER $HOME/.minikube 
This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true`)
		}
		if err := util.MaybeChownDirRecursiveToMinikubeUser(constants.GetMinipath()); err != nil {
			glog.Errorf("Error recursively changing ownership of directory %s: %s",
				constants.GetMinipath(), err)
			cmdUtil.MaybeReportErrorAndExit(err)
		}
	}
}

func validateK8sVersion(version string) {
	validVersion, err := kubernetes_versions.IsValidLocalkubeVersion(version, constants.KubernetesVersionGCSURL)
	if err != nil {
		glog.Errorln("Error getting valid kubernetes versions", err)
		os.Exit(1)
	}
	if !validVersion {
		fmt.Println("Invalid Kubernetes version.")
		kubernetes_versions.PrintKubernetesVersionsFromGCS(os.Stdout)
		os.Exit(1)
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
	startCmd.Flags().Bool(createMount, false, "This will start the mount daemon and automatically mount files into minikube")
	startCmd.Flags().String(mountString, constants.DefaultMountDir+":"+constants.DefaultMountEndpoint, "The argument to pass the minikube mount command on start")
	startCmd.Flags().Bool(disableDriverMounts, false, "Disables the filesystem mounts provided by the hypervisors (vboxfs, xhyve-9p)")
	startCmd.Flags().String(isoURL, constants.DefaultIsoUrl, "Location of the minikube iso")
	startCmd.Flags().String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))
	startCmd.Flags().Int(memory, constants.DefaultMemory, "Amount of RAM allocated to the minikube VM")
	startCmd.Flags().Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the minikube VM")
	startCmd.Flags().String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startCmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minikube VM (only supported with Virtualbox driver)")
	startCmd.Flags().String(hypervVirtualSwitch, "", "The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)")
	startCmd.Flags().String(kvmNetwork, "default", "The KVM network name. (only supported with KVM driver)")
	startCmd.Flags().String(xhyveDiskDriver, "ahci-hd", "The disk driver to use [ahci-hd|virtio-blk] (only supported with xhyve driver)")
	startCmd.Flags().StringArrayVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().StringArrayVar(&dockerOpt, "docker-opt", nil, "Specify arbitrary flags to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().String(apiServerName, constants.APIServerName, "The apiserver name which is used in the generated certificate for localkube/kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().String(dnsDomain, constants.ClusterDNSDomain, "The cluster dns domain name used in the kubernetes cluster")
	startCmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", []string{pkgutil.DefaultInsecureRegistry}, "Insecure Docker registries to pass to the Docker daemon")
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

// loadMachineConfig loads only the configs defined in config in the provided machine config object
func loadMachineConfig(machineConfig cluster.MachineConfig, config cfg.MinikubeConfig) cluster.MachineConfig {
	// Iterate through the config and load the defined configs
	for prop, val := range config {
		switch prop {
		case isoURL:
			machineConfig.MinikubeISO = val.(string)
		case memory:
			machineConfig.Memory = int(val.(float64))
		case cpus:
			machineConfig.CPUs = val.(int)
		case humanReadableDiskSize:
			machineConfig.DiskSize = calculateDiskSizeInMB(val.(string))
		case vmDriver:
			machineConfig.VMDriver = val.(string)
		case xhyveDiskDriver:
			machineConfig.XhyveDiskDriver = val.(string)
		case "docker-env":
			// machineConfig.DockerEnv = val.([]string)
		case "docker-opt":
			// machineConfig.DockerOpt = val.([]string)
		case "insecure-registry":
			// machineConfig.InsecureRegistry = val.([]string)
		case "registry-mirror":
			// machineConfig.RegistryMirror = val.([]string)
		case hostOnlyCIDR:
			machineConfig.HostOnlyCIDR = val.(string)
		case hypervVirtualSwitch:
			machineConfig.HypervVirtualSwitch = val.(string)
		case kvmNetwork:
			machineConfig.KvmNetwork = val.(string)
		// case "":
		// machineConfig.Downloader =
		case disableDriverMounts:
			machineConfig.DisableDriverMounts = val.(bool)
		default:
			// unknown config
		}
	}

	return machineConfig
}

// loadKubernetesConfig loads only the configs defined in config in the provided kubernetes config object
func loadKubernetesConfig(kubernetesConfig cluster.KubernetesConfig, config cfg.MinikubeConfig) cluster.KubernetesConfig {
	for prop, val := range config {
		switch prop {
		case kubernetesVersion:
			kubernetesConfig.KubernetesVersion = val.(string)
		// case "":
		//  kubernetesConfig.NodeIP =
		case apiServerName:
			kubernetesConfig.APIServerName = val.(string)
		case dnsDomain:
			kubernetesConfig.DNSDomain = val.(string)
		case featureGates:
			kubernetesConfig.FeatureGates = val.(string)
		case containerRuntime:
			kubernetesConfig.ContainerRuntime = val.(string)
		case networkPlugin:
			kubernetesConfig.NetworkPlugin = val.(string)
		case "extra-config":
			kubernetesConfig.ExtraOptions = val.(util.ExtraOptionSlice)
		default:
			// unknown config
		}
	}

	return kubernetesConfig
}

// ReadProfileConfig reads in the JSON minikube profile config
func ReadProfileConfig(profile string) (cfg.MinikubeConfig, error) {
	profileConfigFile := constants.GetProfileFile(profile)
	f, err := os.Open(profileConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("Could not open file %s: %s", profileConfigFile, err)
	}
	m, err := decode(f)
	if err != nil {
		return nil, fmt.Errorf("Could not decode config %s: %s", profileConfigFile, err)
	}

	return m, nil
}

func decode(r io.Reader) (cfg.MinikubeConfig, error) {
	var data cfg.MinikubeConfig
	err := json.NewDecoder(r).Decode(&data)
	return data, err
}

// WriteProfileConfig writes a minikube profile config.
func WriteProfileConfig(m cfg.MinikubeConfig, profile string) error {
	profileConfigFile := constants.GetProfileFile(profile)
	f, err := os.Create(profileConfigFile)
	if err != nil {
		return fmt.Errorf("Coult not open file %s: %s", profileConfigFile, err)
	}
	defer f.Close()
	err = encode(f, m)
	if err != nil {
		return fmt.Errorf("Error encoding config %s: %s", profileConfigFile, err)
	}
	return nil
}

func encode(w io.Writer, m cfg.MinikubeConfig) error {
	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}

func getDefaultConfig() cfg.MinikubeConfig {
	return cfg.MinikubeConfig{
		keepContext:         constants.DefaultKeepContext,
		createMount:         false,
		mountString:         constants.DefaultMountDir + ":" + constants.DefaultMountEndpoint,
		disableDriverMounts: false,
		isoURL:              constants.DefaultIsoUrl,
		vmDriver:            constants.DefaultVMDriver,
		memory:              constants.DefaultMemory,
		cpus:                constants.DefaultCPUS,
		humanReadableDiskSize: constants.DefaultDiskSize,
		hostOnlyCIDR:          "192.168.99.1/24",
		hypervVirtualSwitch:   "",
		kvmNetwork:            "default",
		xhyveDiskDriver:       "ahci-hd",
		"docker-env":          nil,
		"docker-opt":          nil,
		apiServerName:         constants.APIServerName,
		dnsDomain:             constants.ClusterDNSDomain,
		"insecure-registry":   []string{pkgutil.DefaultInsecureRegistry},
		"registry-mirror":     nil,
		kubernetesVersion:     constants.DefaultKubernetesVersion,
		containerRuntime:      "",
		networkPlugin:         "",
		featureGates:          "",

		cfg.WantUpdateNotification:    true,
		cfg.ReminderWaitPeriodInHours: true,
		cfg.WantReportError:           true,
		cfg.WantReportErrorPrompt:     true,
		cfg.WantKubectlDownloadMsg:    true,

		cfg.Dashboard:           true,
		cfg.AddonManager:        false,
		cfg.DefaultStorageclass: false,
		cfg.KubeDNS:             false,
		cfg.Heapster:            false,
		cfg.Ingress:             false,
		cfg.Registry:            false,
		cfg.RegistryCreds:       false,
	}
}
