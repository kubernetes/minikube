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
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"k8s.io/minikube/pkg/minikube/drivers/none"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	cmdutil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/kubeadm"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/command"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/proxy"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/version"
)

const (
	isoURL                = "iso-url"
	memory                = "memory"
	cpus                  = "cpus"
	humanReadableDiskSize = "disk-size"
	vmDriver              = "vm-driver"
	nfsSharesRoot         = "nfs-shares-root"
	nfsShare              = "nfs-share"
	kubernetesVersion     = "kubernetes-version"
	hostOnlyCIDR          = "host-only-cidr"
	containerRuntime      = "container-runtime"
	criSocket             = "cri-socket"
	networkPlugin         = "network-plugin"
	enableDefaultCNI      = "enable-default-cni"
	hypervVirtualSwitch   = "hyperv-virtual-switch"
	kvmNetwork            = "kvm-network"
	kvmQemuURI            = "kvm-qemu-uri"
	kvmGPU                = "kvm-gpu"
	kvmHidden             = "kvm-hidden"
	keepContext           = "keep-context"
	createMount           = "mount"
	featureGates          = "feature-gates"
	apiServerName         = "apiserver-name"
	apiServerPort         = "apiserver-port"
	dnsDomain             = "dns-domain"
	serviceCIDR           = "service-cluster-ip-range"
	imageRepository       = "image-repository"
	imageMirrorCountry    = "image-mirror-country"
	mountString           = "mount-string"
	disableDriverMounts   = "disable-driver-mounts"
	cacheImages           = "cache-images"
	uuid                  = "uuid"
	vpnkitSock            = "hyperkit-vpnkit-sock"
	vsockPorts            = "hyperkit-vsock-ports"
	embedCerts            = "embed-certs"
	noVTXCheck            = "no-vtx-check"
	downloadOnly          = "download-only"
	dnsProxy              = "dns-proxy"
	hostDNSResolver       = "host-dns-resolver"
	waitUntilHealthy      = "wait"
)

var (
	registryMirror   []string
	dockerEnv        []string
	dockerOpt        []string
	insecureRegistry []string
	apiServerNames   []string
	apiServerIPs     []net.IP
	extraOptions     pkgutil.ExtraOptionSlice
)

func init() {
	initMinikubeFlags()
	initKubernetesFlags()
	initDriverFlags()
	initNetworkingFlags()
	if err := viper.BindPFlags(startCmd.Flags()); err != nil {
		exit.WithError("unable to bind flags", err)
	}

	RootCmd.AddCommand(startCmd)
}

// initMinikubeFlags includes commandline flags for minikube.
func initMinikubeFlags() {
	startCmd.Flags().Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the minikube VM")
	startCmd.Flags().String(memory, constants.DefaultMemorySize, "Amount of RAM allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startCmd.Flags().String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startCmd.Flags().Bool(downloadOnly, false, "If true, only download and cache files for later use - don't install or start anything.")
	startCmd.Flags().Bool(cacheImages, true, "If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --vm-driver=none.")
	startCmd.Flags().String(isoURL, constants.DefaultISOURL, "Location of the minikube iso")
	startCmd.Flags().Bool(keepContext, constants.DefaultKeepContext, "This will keep the existing kubectl context and will create a minikube context.")
	startCmd.Flags().String(containerRuntime, "docker", "The container runtime to be used (docker, crio, containerd)")
	startCmd.Flags().Bool(createMount, false, "This will start the mount daemon and automatically mount files into minikube")
	startCmd.Flags().String(mountString, constants.DefaultMountDir+":"+constants.DefaultMountEndpoint, "The argument to pass the minikube mount command on start")
	startCmd.Flags().String(criSocket, "", "The cri socket path to be used")
	startCmd.Flags().String(networkPlugin, "", "The name of the network plugin")
	startCmd.Flags().Bool(enableDefaultCNI, false, "Enable the default CNI plugin (/etc/cni/net.d/k8s.conf). Used in conjunction with \"--network-plugin=cni\"")
	startCmd.Flags().Bool(waitUntilHealthy, true, "Wait until Kubernetes core services are healthy before exiting")
}

// initKubernetesFlags inits the commandline flags for kubernetes related options
func initKubernetesFlags() {
	startCmd.Flags().String(kubernetesVersion, constants.DefaultKubernetesVersion, "The kubernetes version that the minikube VM will use (ex: v1.2.3)")
	startCmd.Flags().Var(&extraOptions, "extra-config",
		`A set of key=value pairs that describe configuration that may be passed to different components.
		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
		Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
		Valid kubeadm parameters: `+fmt.Sprintf("%s, %s", strings.Join(kubeadm.KubeadmExtraArgsWhitelist[kubeadm.KubeadmCmdParam], ", "), strings.Join(kubeadm.KubeadmExtraArgsWhitelist[kubeadm.KubeadmConfigParam], ",")))
	startCmd.Flags().String(featureGates, "", "A set of key=value pairs that describe feature gates for alpha/experimental features.")
	startCmd.Flags().String(dnsDomain, constants.ClusterDNSDomain, "The cluster dns domain name used in the kubernetes cluster")
	startCmd.Flags().Int(apiServerPort, pkgutil.APIServerPort, "The apiserver listening port")
	startCmd.Flags().String(apiServerName, constants.APIServerName, "The apiserver name which is used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().StringArrayVar(&apiServerNames, "apiserver-names", nil, "A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().IPSliceVar(&apiServerIPs, "apiserver-ips", nil, "A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")

}

// initDriverFlags inits the commandline flags for vm drivers
func initDriverFlags() {
	startCmd.Flags().String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))

	// kvm2
	startCmd.Flags().String(kvmNetwork, "default", "The KVM network name. (only supported with KVM driver)")
	startCmd.Flags().String(kvmQemuURI, "qemu:///system", "The KVM QEMU connection URI. (works only with kvm2 driver on linux)")
	startCmd.Flags().Bool(kvmGPU, false, "Enable experimental NVIDIA GPU support in minikube")
	startCmd.Flags().Bool(kvmHidden, false, "Hide the hypervisor signature from the guest in minikube")

	// virtualbox
	startCmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minikube VM (only supported with Virtualbox driver)")
	startCmd.Flags().Bool(dnsProxy, false, "Enable proxy for NAT DNS requests (virtualbox)")
	startCmd.Flags().Bool(hostDNSResolver, true, "Enable host resolver for NAT DNS requests (virtualbox)")
	startCmd.Flags().Bool(noVTXCheck, false, "Disable checking for the availability of hardware virtualization before the vm is started (virtualbox)")

	// hyperkit
	startCmd.Flags().StringSlice(vsockPorts, []string{}, "List of guest VSock ports that should be exposed as sockets on the host (Only supported on with hyperkit now).")
	startCmd.Flags().String(uuid, "", "Provide VM UUID to restore MAC address (only supported with Hyperkit driver).")
	startCmd.Flags().String(vpnkitSock, "", "Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock.")
	startCmd.Flags().StringSlice(nfsShare, []string{}, "Local folders to share with Guest via NFS mounts (Only supported on with hyperkit now)")
	startCmd.Flags().String(nfsSharesRoot, "/nfsshares", "Where to root the NFS Shares (defaults to /nfsshares, only supported with hyperkit now)")

	// hyperv
	startCmd.Flags().String(hypervVirtualSwitch, "", "The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)")
}

// initNetworkingFlags inits the commandline flags for connectivity related flags for start
func initNetworkingFlags() {
	startCmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", nil, "Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.")
	startCmd.Flags().StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	startCmd.Flags().String(imageRepository, "", "Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to \"auto\" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers")
	startCmd.Flags().String(imageMirrorCountry, "", "Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn")
	startCmd.Flags().String(serviceCIDR, pkgutil.DefaultServiceCIDR, "The CIDR to be used for service cluster IPs.")
	startCmd.Flags().StringArrayVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().StringArrayVar(&dockerOpt, "docker-opt", nil, "Specify arbitrary flags to pass to the Docker daemon. (format: key=value)")
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a local kubernetes cluster",
	Long:  "Starts a local kubernetes cluster",
	Run:   runStart,
}

// runStart handles the executes the flow of "minikube start"
func runStart(cmd *cobra.Command, args []string) {
	out.T(out.Happy, "minikube {{.version}} on {{.os}} ({{.arch}})", out.V{"version": version.GetVersion(), "os": runtime.GOOS, "arch": runtime.GOARCH})

	vmDriver := viper.GetString(vmDriver)
	if err := cmdcfg.IsValidDriver(runtime.GOOS, vmDriver); err != nil {
		exit.WithCodeT(
			exit.Failure,
			"The driver '{{.driver}}' is not supported on {{.os}}",
			out.V{"driver": vmDriver, "os": runtime.GOOS},
		)
	}
	validateConfig()
	validateUser()
	validateDriverVersion(viper.GetString(vmDriver))

	k8sVersion, isUpgrade := getKubernetesVersion()
	config, err := generateConfig(cmd, k8sVersion)
	if err != nil {
		exit.WithError("Failed to generate config", err)
	}

	// For non-"none", the ISO is required to boot, so block until it is downloaded
	downloadISO(config)

	// With "none", images are persistently stored in Docker, so internal caching isn't necessary.
	skipCache(&config)

	// Now that the ISO is downloaded, pull images in the background while the VM boots.
	var cacheGroup errgroup.Group
	beginCacheImages(&cacheGroup, config.KubernetesConfig.ImageRepository, k8sVersion)

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, saveConfig must be called before startHost, and again afterwards when we know the IP.
	if err := saveConfig(&config); err != nil {
		exit.WithError("Failed to save config", err)
	}

	// exits here in case of --download-only option.
	handleDownloadOnly(&cacheGroup, k8sVersion)
	mRunner, preExists, machineAPI, host := startMachine(&config)
	defer machineAPI.Close()
	// configure the runtime (docker, containerd, crio)
	cr := configureRuntimes(mRunner)
	showVersionInfo(k8sVersion, cr)
	waitCacheImages(&cacheGroup)

	// setup kube adm and certs and return bootstrapperx
	bs := setupKubeAdm(machineAPI, config.KubernetesConfig)
	// The kube config must be update must come before bootstrapping, otherwise health checks may use a stale IP
	kubeconfig := updateKubeConfig(host, &config)
	// pull images or restart cluster
	bootstrapCluster(bs, cr, mRunner, config.KubernetesConfig, preExists, isUpgrade)
	configureMounts()
	if err = loadCachedImagesInConfigFile(); err != nil {
		out.T(out.FailureType, "Unable to load cached images from config file.")
	}
	// special ops for none driver, like change minikube directory.
	prepareNone(viper.GetString(vmDriver))
	if viper.GetBool(waitUntilHealthy) {
		if err := bs.WaitCluster(config.KubernetesConfig); err != nil {
			exit.WithError("Wait failed", err)
		}
	}
	showKubectlConnectInfo(kubeconfig)

}

func handleDownloadOnly(cacheGroup *errgroup.Group, k8sVersion string) {
	// If --download-only, complete the remaining downloads and exit.
	if !viper.GetBool(downloadOnly) {
		return
	}
	if err := doCacheBinaries(k8sVersion); err != nil {
		exit.WithError("Failed to cache binaries", err)
	}
	waitCacheImages(cacheGroup)
	if err := CacheImagesInConfigFile(); err != nil {
		exit.WithError("Failed to cache images", err)
	}
	out.T(out.Check, "Download complete!")
	os.Exit(0)

}

func startMachine(config *cfg.Config) (runner command.Runner, preExists bool, machineAPI libmachine.API, host *host.Host) {
	m, err := machine.NewAPIClient()
	if err != nil {
		exit.WithError("Failed to get machine client", err)
	}
	host, preExists = startHost(m, config.MachineConfig)

	ip := validateNetwork(host)
	// Bypass proxy for minikube's vm host ip
	err = proxy.ExcludeIP(ip)
	if err != nil {
		out.ErrT(out.FailureType, "Failed to set NO_PROXY Env. Please use `export NO_PROXY=$NO_PROXY,{{.ip}}`.", out.V{"ip": ip})
	}
	// Save IP to configuration file for subsequent use
	config.KubernetesConfig.NodeIP = ip
	if err := saveConfig(config); err != nil {
		exit.WithError("Failed to save config", err)
	}
	runner, err = machine.CommandRunner(host)
	if err != nil {
		exit.WithError("Failed to get command runner", err)
	}

	return runner, preExists, m, host
}

func getKubernetesVersion() (k8sVersion string, isUpgrade bool) {
	oldConfig, err := cfg.Load()
	if err != nil && !os.IsNotExist(err) {
		exit.WithCodeT(exit.Data, "Unable to load config: {{.error}}", out.V{"error": err})
	}
	return validateKubernetesVersions(oldConfig)
}

func downloadISO(config cfg.Config) {
	if viper.GetString(vmDriver) != constants.DriverNone {
		if err := cluster.CacheISO(config.MachineConfig); err != nil {
			exit.WithError("Failed to cache ISO", err)
		}
	}
}

func skipCache(config *cfg.Config) {
	if viper.GetString(vmDriver) == constants.DriverNone {
		viper.Set(cacheImages, false)
		config.KubernetesConfig.ShouldLoadCachedImages = false
	}
}

func showVersionInfo(k8sVersion string, cr cruntime.Manager) {
	version, _ := cr.Version()
	out.T(cr.Style(), "Configuring environment for Kubernetes {{.k8sVersion}} on {{.runtime}} {{.runtimeVersion}}", out.V{"k8sVersion": k8sVersion, "runtime": cr.Name(), "runtimeVersion": version})
	for _, v := range dockerOpt {
		out.T(out.Option, "opt {{.docker_option}}", out.V{"docker_option": v})
	}
	for _, v := range dockerEnv {
		out.T(out.Option, "env {{.docker_env}}", out.V{"docker_env": v})
	}
}

func showKubectlConnectInfo(kubeconfig *pkgutil.KubeConfigSetup) {
	if kubeconfig.KeepContext {
		out.T(out.Kubectl, "To connect to this cluster, use: kubectl --context={{.name}}", out.V{"name": kubeconfig.ClusterName})
	} else {
		if !viper.GetBool(waitUntilHealthy) {
			out.T(out.Ready, "kubectl has been configured configured to use {{.name}}", out.V{"name": cfg.GetMachineName()})
		} else {
			out.T(out.Ready, "Done! kubectl is now configured to use {{.name}}", out.V{"name": cfg.GetMachineName()})
		}
	}
	_, err := exec.LookPath("kubectl")
	if err != nil {
		out.T(out.Tip, "For best results, install kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl/")
	}
}

func selectImageRepository(mirrorCountry string, k8sVersion string) (bool, string, error) {
	var tryCountries []string
	var fallback string
	glog.Infof("selecting image repository for country %s ...", mirrorCountry)

	if mirrorCountry != "" {
		localRepos, ok := constants.ImageRepositories[mirrorCountry]
		if !ok || len(localRepos) == 0 {
			return false, "", fmt.Errorf("invalid image mirror country code: %s", mirrorCountry)
		}

		tryCountries = append(tryCountries, mirrorCountry)

		// we'll use the first repository as fallback
		// when none of the mirrors in the given location is available
		fallback = localRepos[0]

	} else {
		// always make sure global is preferred
		tryCountries = append(tryCountries, "global")
		for k := range constants.ImageRepositories {
			if strings.ToLower(k) != "global" {
				tryCountries = append(tryCountries, k)
			}
		}
	}

	checkRepository := func(repo string) error {
		podInfraContainerImage, _ := constants.GetKubeadmCachedImages(repo, k8sVersion)

		ref, err := name.ParseReference(podInfraContainerImage, name.WeakValidation)
		if err != nil {
			return err
		}

		_, err = remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
		return err
	}

	for _, code := range tryCountries {
		localRepos := constants.ImageRepositories[code]
		for _, repo := range localRepos {
			err := checkRepository(repo)
			if err == nil {
				return true, repo, nil
			}
		}
	}

	return false, fallback, nil
}

// validerUser validates minikube is run by the recommended user (privileged or regular)
func validateUser() {
	u, err := user.Current()
	d := viper.GetString(vmDriver)
	// Check if minikube needs to run with sudo or not.
	if err == nil {
		if d == constants.DriverNone && u.Name != "root" {
			exit.UsageT(`Please run with sudo. the vm-driver "{{.driver_name}}" requires sudo.`, out.V{"driver_name": constants.DriverNone})
		} else if u.Name == "root" && !(d == constants.DriverHyperv || d == constants.DriverNone) {
			out.T(out.WarningType, "Please don't run minikube as root or with 'sudo' privileges. It isn't necessary with {{.driver}} driver.", out.V{"driver": d})
		}

	} else {
		glog.Errorf("Error getting the current user: %v", err)
	}

}

// validateConfig validates the supplied configuration against known bad combinations
func validateConfig() {
	diskSizeMB := pkgutil.CalculateSizeInMB(viper.GetString(humanReadableDiskSize))
	if diskSizeMB < pkgutil.CalculateSizeInMB(constants.MinimumDiskSize) {
		exit.WithCodeT(exit.Config, "Requested disk size {{.size_in_mb}} is less than minimum of {{.size_in_mb2}}", out.V{"size_in_mb": diskSizeMB, "size_in_mb2": pkgutil.CalculateSizeInMB(constants.MinimumDiskSize)})
	}

	err := autoSetOptions(viper.GetString(vmDriver))
	if err != nil {
		glog.Errorf("Error autoSetOptions : %v", err)
	}

	memorySizeMB := pkgutil.CalculateSizeInMB(viper.GetString(memory))
	if memorySizeMB < pkgutil.CalculateSizeInMB(constants.MinimumMemorySize) {
		exit.UsageT("Requested memory allocation {{.size_in_mb}} is less than the minimum allowed of {{.size_in_mb2}}", out.V{"size_in_mb": memorySizeMB, "size_in_mb2": pkgutil.CalculateSizeInMB(constants.MinimumMemorySize)})
	}
	if memorySizeMB < pkgutil.CalculateSizeInMB(constants.DefaultMemorySize) {
		out.T(out.Notice, "Requested memory allocation ({{.memory}}MB) is less than the default memory allocation of {{.default_memorysize}}MB. Beware that minikube might not work correctly or crash unexpectedly.",
			out.V{"memory": memorySizeMB, "default_memorysize": pkgutil.CalculateSizeInMB(constants.DefaultMemorySize)})
	}

	// check that kubeadm extra args contain only whitelisted parameters
	for param := range extraOptions.AsMap().Get(kubeadm.Kubeadm) {
		if !pkgutil.ContainsString(kubeadm.KubeadmExtraArgsWhitelist[kubeadm.KubeadmCmdParam], param) &&
			!pkgutil.ContainsString(kubeadm.KubeadmExtraArgsWhitelist[kubeadm.KubeadmConfigParam], param) {
			exit.UsageT("Sorry, the kubeadm.{{.parameter_name}} parameter is currently not supported by --extra-config", out.V{"parameter_name": param})
		}
	}

	validateRegistryMirror()
}

// This function validates if the --registry-mirror
// args match the format of http://localhost
func validateRegistryMirror() {

	if len(registryMirror) > 0 {
		for _, loc := range registryMirror {
			URL, err := url.Parse(loc)
			if err != nil {
				glog.Errorln("Error Parsing URL: ", err)
			}
			if (URL.Scheme != "http" && URL.Scheme != "https") || URL.Path != "" {
				exit.UsageT("Sorry, url provided with --registry-mirror flag is invalid {{.url}}", out.V{"url": loc})
			}

		}
	}
}

// doCacheBinaries caches Kubernetes binaries in the foreground
func doCacheBinaries(k8sVersion string) error {
	return machine.CacheBinariesForBootstrapper(k8sVersion, viper.GetString(cmdcfg.Bootstrapper))
}

// beginCacheImages caches Docker images in the background
func beginCacheImages(g *errgroup.Group, imageRepository string, k8sVersion string) {
	if !viper.GetBool(cacheImages) {
		return
	}

	g.Go(func() error {
		return machine.CacheImagesForBootstrapper(imageRepository, k8sVersion, viper.GetString(cmdcfg.Bootstrapper))
	})
}

// waitCacheImages blocks until the image cache jobs complete
func waitCacheImages(g *errgroup.Group) {
	if !viper.GetBool(cacheImages) {
		return
	}
	if err := g.Wait(); err != nil {
		glog.Errorln("Error caching images: ", err)
	}
}

// generateConfig generates cfg.Config based on flags and supplied arguments
func generateConfig(cmd *cobra.Command, k8sVersion string) (cfg.Config, error) {
	r, err := cruntime.New(cruntime.Config{Type: viper.GetString(containerRuntime)})
	if err != nil {
		return cfg.Config{}, err
	}

	// Pick good default values for --network-plugin and --enable-default-cni based on runtime.
	selectedEnableDefaultCNI := viper.GetBool(enableDefaultCNI)
	selectedNetworkPlugin := viper.GetString(networkPlugin)
	if r.DefaultCNI() && !cmd.Flags().Changed(networkPlugin) {
		selectedNetworkPlugin = "cni"
		if !cmd.Flags().Changed(enableDefaultCNI) {
			selectedEnableDefaultCNI = true
		}
	}

	// Feed Docker our host proxy environment by default, so that it can pull images
	if _, ok := r.(*cruntime.Docker); ok {
		if !cmd.Flags().Changed("docker-env") {
			for _, k := range proxy.EnvVars {
				if v := os.Getenv(k); v != "" {
					// convert https_proxy to HTTPS_PROXY for linux
					// TODO (@medyagh): if user has both http_proxy & HTTPS_PROXY set merge them.
					k = strings.ToUpper(k)
					dockerEnv = append(dockerEnv, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
	}

	repository := viper.GetString(imageRepository)
	mirrorCountry := strings.ToLower(viper.GetString(imageMirrorCountry))
	if strings.ToLower(repository) == "auto" || mirrorCountry != "" {
		found, autoSelectedRepository, err := selectImageRepository(mirrorCountry, k8sVersion)
		if err != nil {
			exit.WithError("Failed to check main repository and mirrors for images for images", err)
		}

		if !found {
			if autoSelectedRepository == "" {
				exit.WithCodeT(exit.Failure, "None of known repositories is accessible. Consider specifying an alternative image repository with --image-repository flag")
			} else {
				out.WarningT("None of known repositories in your location is accessible. Use {{.image_repository_name}} as fallback.", out.V{"image_repository_name": autoSelectedRepository})
			}
		}

		repository = autoSelectedRepository
	}

	if repository != "" {
		out.T(out.SuccessType, "Using image repository {{.name}}", out.V{"name": repository})
	}

	cfg := cfg.Config{
		MachineConfig: cfg.MachineConfig{
			KeepContext:         viper.GetBool(keepContext),
			MinikubeISO:         viper.GetString(isoURL),
			Memory:              pkgutil.CalculateSizeInMB(viper.GetString(memory)),
			CPUs:                viper.GetInt(cpus),
			DiskSize:            pkgutil.CalculateSizeInMB(viper.GetString(humanReadableDiskSize)),
			VMDriver:            viper.GetString(vmDriver),
			ContainerRuntime:    viper.GetString(containerRuntime),
			HyperkitVpnKitSock:  viper.GetString(vpnkitSock),
			HyperkitVSockPorts:  viper.GetStringSlice(vsockPorts),
			NFSShare:            viper.GetStringSlice(nfsShare),
			NFSSharesRoot:       viper.GetString(nfsSharesRoot),
			DockerEnv:           dockerEnv,
			DockerOpt:           dockerOpt,
			InsecureRegistry:    insecureRegistry,
			RegistryMirror:      registryMirror,
			HostOnlyCIDR:        viper.GetString(hostOnlyCIDR),
			HypervVirtualSwitch: viper.GetString(hypervVirtualSwitch),
			KVMNetwork:          viper.GetString(kvmNetwork),
			KVMQemuURI:          viper.GetString(kvmQemuURI),
			KVMGPU:              viper.GetBool(kvmGPU),
			KVMHidden:           viper.GetBool(kvmHidden),
			Downloader:          pkgutil.DefaultDownloader{},
			DisableDriverMounts: viper.GetBool(disableDriverMounts),
			UUID:                viper.GetString(uuid),
			NoVTXCheck:          viper.GetBool(noVTXCheck),
			DNSProxy:            viper.GetBool(dnsProxy),
			HostDNSResolver:     viper.GetBool(hostDNSResolver),
		},
		KubernetesConfig: cfg.KubernetesConfig{
			KubernetesVersion:      k8sVersion,
			NodePort:               viper.GetInt(apiServerPort),
			NodeName:               constants.DefaultNodeName,
			APIServerName:          viper.GetString(apiServerName),
			APIServerNames:         apiServerNames,
			APIServerIPs:           apiServerIPs,
			DNSDomain:              viper.GetString(dnsDomain),
			FeatureGates:           viper.GetString(featureGates),
			ContainerRuntime:       viper.GetString(containerRuntime),
			CRISocket:              viper.GetString(criSocket),
			NetworkPlugin:          selectedNetworkPlugin,
			ServiceCIDR:            viper.GetString(serviceCIDR),
			ImageRepository:        repository,
			ExtraOptions:           extraOptions,
			ShouldLoadCachedImages: viper.GetBool(cacheImages),
			EnableDefaultCNI:       selectedEnableDefaultCNI,
		},
	}
	return cfg, nil
}

// autoSetOptions sets the options needed for specific vm-driver automatically.
func autoSetOptions(vmDriver string) error {
	//  options for none driver
	if vmDriver == constants.DriverNone {
		if o := none.AutoOptions(); o != "" {
			return extraOptions.Set(o)
		}
	}
	return nil
}

// prepareNone prepares the user and host for the joy of the "none" driver
func prepareNone(vmDriver string) {
	if vmDriver != constants.DriverNone {
		return
	}
	out.T(out.StartingNone, "Configuring local host environment ...")
	if viper.GetBool(cfg.WantNoneDriverWarning) {
		out.T(out.Empty, "")
		out.WarningT("The 'none' driver provides limited isolation and may reduce system security and reliability.")
		out.WarningT("For more information, see:")
		out.T(out.URL, "https://github.com/kubernetes/minikube/blob/master/docs/vmdriver-none.md")
		out.T(out.Empty, "")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		home := os.Getenv("HOME")
		out.WarningT("kubectl and minikube configuration will be stored in {{.home_folder}}", out.V{"home_folder": home})
		out.WarningT("To use kubectl or minikube commands as your own user, you may")
		out.WarningT("need to relocate them. For example, to overwrite your own settings:")

		out.T(out.Empty, "")
		out.T(out.Command, "sudo mv {{.home_folder}}/.kube {{.home_folder}}/.minikube $HOME", out.V{"home_folder": home})
		out.T(out.Command, "sudo chown -R $USER $HOME/.kube $HOME/.minikube")
		out.T(out.Empty, "")

		out.T(out.Tip, "This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true")
	}

	if err := pkgutil.MaybeChownDirRecursiveToMinikubeUser(constants.GetMinipath()); err != nil {
		exit.WithCodeT(exit.Permissions, "Failed to chown {{.minikube_dir_path}}: {{.error}}", out.V{"minikube_dir_path": constants.GetMinipath(), "error": err})
	}
}

// startHost starts a new minikube host using a VM or None
func startHost(api libmachine.API, mc cfg.MachineConfig) (*host.Host, bool) {
	exists, err := api.Exists(cfg.GetMachineName())
	if err != nil {
		exit.WithError("Failed to check if machine exists", err)
	}

	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(api, mc)
		if err != nil {
			glog.Errorf("StartHost: %v", err)
		}
		return err
	}
	if err = pkgutil.RetryAfter(3, start, 2*time.Second); err != nil {
		exit.WithError("Unable to start VM", err)
	}
	return host, exists
}

// validateNetwork tries to catch network problems as soon as possible
func validateNetwork(h *host.Host) string {
	ip, err := h.Driver.GetIP()
	if err != nil {
		exit.WithError("Unable to get VM IP address", err)
	}

	optSeen := false
	warnedOnce := false
	for _, k := range proxy.EnvVars {
		if v := os.Getenv(k); v != "" {
			if !optSeen {
				out.T(out.Internet, "Found network options:")
				optSeen = true
			}
			out.T(out.Option, "{{.key}}={{.value}}", out.V{"key": k, "value": v})
			ipExcluded := proxy.IsIPExcluded(ip) // Skip warning if minikube ip is already in NO_PROXY
			k = strings.ToUpper(k)               // for http_proxy & https_proxy
			if (k == "HTTP_PROXY" || k == "HTTPS_PROXY") && !ipExcluded && !warnedOnce {
				out.WarningT("You appear to be using a proxy, but your NO_PROXY environment does not include the minikube IP ({{.ip_address}}). Please see https://github.com/kubernetes/minikube/blob/master/docs/http_proxy.md for more details", out.V{"ip_address": ip})
				warnedOnce = true
			}
		}
	}

	// Here is where we should be checking connectivity to/from the VM
	return ip
}

// validateKubernetesVersions ensures that the requested version is reasonable
func validateKubernetesVersions(old *cfg.Config) (string, bool) {
	rawVersion := viper.GetString(kubernetesVersion)
	isUpgrade := false
	if rawVersion == "" {
		rawVersion = constants.DefaultKubernetesVersion
	}

	nvs, err := semver.Make(strings.TrimPrefix(rawVersion, version.VersionPrefix))
	if err != nil {
		exit.WithCodeT(exit.Data, `Unable to parse "{{.kubenretes_version}}": {{.error}}`, out.V{"kubenretes_version": rawVersion, "error": err})
	}
	nv := version.VersionPrefix + nvs.String()

	if old == nil || old.KubernetesConfig.KubernetesVersion == "" {
		return nv, isUpgrade
	}

	ovs, err := semver.Make(strings.TrimPrefix(old.KubernetesConfig.KubernetesVersion, version.VersionPrefix))
	if err != nil {
		glog.Errorf("Error parsing old version %q: %v", old.KubernetesConfig.KubernetesVersion, err)
	}

	if nvs.LT(ovs) {
		nv = version.VersionPrefix + ovs.String()
		out.ErrT(out.Conflict, "Kubernetes downgrade is not supported, will continue to use {{.version}}", out.V{"version": nv})
		return nv, isUpgrade
	}
	if nvs.GT(ovs) {
		out.T(out.ThumbsUp, "minikube will upgrade the local cluster from Kubernetes {{.old}} to {{.new}}", out.V{"old": ovs, "new": nvs})
		isUpgrade = true
	}
	return nv, isUpgrade
}

// setupKubeAdm adds any requested files into the VM before Kubernetes is started
func setupKubeAdm(mAPI libmachine.API, kc cfg.KubernetesConfig) bootstrapper.Bootstrapper {
	bs, err := getClusterBootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper))
	if err != nil {
		exit.WithError("Failed to get bootstrapper", err)
	}
	for _, eo := range extraOptions {
		out.T(out.Option, "{{.extra_option_component_name}}.{{.key}}={{.value}}", out.V{"extra_option_component_name": eo.Component, "key": eo.Key, "value": eo.Value})
	}
	// Loads cached images, generates config files, download binaries
	if err := bs.UpdateCluster(kc); err != nil {
		exit.WithError("Failed to update cluster", err)
	}
	if err := bs.SetupCerts(kc); err != nil {
		exit.WithError("Failed to setup certs", err)
	}
	return bs
}

// updateKubeConfig sets up kubectl
func updateKubeConfig(h *host.Host, c *cfg.Config) *pkgutil.KubeConfigSetup {
	addr, err := h.Driver.GetURL()
	if err != nil {
		exit.WithError("Failed to get driver URL", err)
	}
	addr = strings.Replace(addr, "tcp://", "https://", -1)
	addr = strings.Replace(addr, ":2376", ":"+strconv.Itoa(c.KubernetesConfig.NodePort), -1)
	if c.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.Replace(addr, c.KubernetesConfig.NodeIP, c.KubernetesConfig.APIServerName, -1)
	}

	kcs := &pkgutil.KubeConfigSetup{
		ClusterName:          cfg.GetMachineName(),
		ClusterServerAddress: addr,
		ClientCertificate:    constants.MakeMiniPath("client.crt"),
		ClientKey:            constants.MakeMiniPath("client.key"),
		CertificateAuthority: constants.MakeMiniPath("ca.crt"),
		KeepContext:          viper.GetBool(keepContext),
		EmbedCerts:           viper.GetBool(embedCerts),
	}
	kcs.SetKubeConfigFile(cmdutil.GetKubeConfigPath())
	if err := pkgutil.SetupKubeConfig(kcs); err != nil {
		exit.WithError("Failed to setup kubeconfig", err)
	}
	return kcs
}

// configureRuntimes does what needs to happen to get a runtime going.
func configureRuntimes(runner cruntime.CommandRunner) cruntime.Manager {
	config := cruntime.Config{Type: viper.GetString(containerRuntime), Runner: runner}
	cr, err := cruntime.New(config)
	if err != nil {
		exit.WithError("Failed runtime", err)
	}

	disableOthers := true
	if viper.GetString(vmDriver) == constants.DriverNone {
		disableOthers = false
	}
	err = cr.Enable(disableOthers)
	if err != nil {
		exit.WithError("Failed to enable container runtime", err)
	}

	return cr
}

// bootstrapCluster starts Kubernetes using the chosen bootstrapper
func bootstrapCluster(bs bootstrapper.Bootstrapper, r cruntime.Manager, runner command.Runner, kc cfg.KubernetesConfig, preexisting bool, isUpgrade bool) {
	// hum. bootstrapper.Bootstrapper should probably have a Name function.
	bsName := viper.GetString(cmdcfg.Bootstrapper)

	if isUpgrade || !preexisting {
		out.T(out.Pulling, "Pulling images ...")
		if err := bs.PullImages(kc); err != nil {
			out.T(out.FailureType, "Unable to pull images, which may be OK: {{.error}}", out.V{"error": err})
		}
	}

	if preexisting {
		out.T(out.Restarting, "Relaunching Kubernetes {{.version}} using {{.bootstrapper}} ... ", out.V{"version": kc.KubernetesVersion, "bootstrapper": bsName})
		if err := bs.RestartCluster(kc); err != nil {
			exit.WithLogEntries("Error restarting cluster", err, logs.FindProblems(r, bs, runner))
		}
		return
	}

	out.T(out.Launch, "Launching Kubernetes ... ")
	if err := bs.StartCluster(kc); err != nil {
		exit.WithLogEntries("Error starting cluster", err, logs.FindProblems(r, bs, runner))
	}
}

// configureMounts configures any requested filesystem mounts
func configureMounts() {
	if !viper.GetBool(createMount) {
		return
	}

	out.T(out.Mounting, "Creating mount {{.name}} ...", out.V{"name": viper.GetString(mountString)})
	path := os.Args[0]
	mountDebugVal := 0
	if glog.V(8) {
		mountDebugVal = 1
	}
	mountCmd := exec.Command(path, "mount", fmt.Sprintf("--v=%d", mountDebugVal), viper.GetString(mountString))
	mountCmd.Env = append(os.Environ(), constants.IsMinikubeChildProcess+"=true")
	if glog.V(8) {
		mountCmd.Stdout = os.Stdout
		mountCmd.Stderr = os.Stderr
	}
	if err := mountCmd.Start(); err != nil {
		exit.WithError("Error starting mount", err)
	}
	if err := ioutil.WriteFile(filepath.Join(constants.GetMinipath(), constants.MountProcessFileName), []byte(strconv.Itoa(mountCmd.Process.Pid)), 0644); err != nil {
		exit.WithError("Error writing mount pid", err)
	}
}

// saveConfig saves profile cluster configuration in $MINIKUBE_HOME/profiles/<profilename>/config.json
func saveConfig(clusterConfig *cfg.Config) error {
	data, err := json.MarshalIndent(clusterConfig, "", "    ")
	if err != nil {
		return err
	}
	glog.Infof("Saving config:\n%s", data)
	path := constants.GetProfileFile(viper.GetString(cfg.MachineProfile))
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	// If no config file exists, don't worry about swapping paths
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := ioutil.WriteFile(path, data, 0600); err != nil {
			return err
		}
		return nil
	}

	tf, err := ioutil.TempFile(filepath.Dir(path), "config.json.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tf.Name())

	if err = ioutil.WriteFile(tf.Name(), data, 0600); err != nil {
		return err
	}

	if err = tf.Close(); err != nil {
		return err
	}

	if err = os.Remove(path); err != nil {
		return err
	}

	if err = os.Rename(tf.Name(), path); err != nil {
		return err
	}
	return nil
}

func validateDriverVersion(vmDriver string) {
	var (
		driverExecutable    string
		driverDocumentation string
	)

	switch vmDriver {
	case constants.DriverKvm2:
		driverExecutable = fmt.Sprintf("docker-machine-driver-%s", constants.DriverKvm2)
		driverDocumentation = fmt.Sprintf("%s#%s", constants.DriverDocumentation, "kvm2-upgrade")
	case constants.DriverHyperkit:
		driverExecutable = fmt.Sprintf("docker-machine-driver-%s", constants.DriverHyperkit)
		driverDocumentation = fmt.Sprintf("%s#%s", constants.DriverDocumentation, "hyperkit-upgrade")
	default: // driver doesn't support version
		return
	}

	cmd := exec.Command(driverExecutable, "version")
	output, err := cmd.Output()

	// we don't want to fail if an error was returned,
	// libmachine has a nice message for the user if the driver isn't present
	if err != nil {
		out.WarningT("Error checking driver version: {{.error}}", out.V{"error": err})
		return
	}

	v := extractVMDriverVersion(string(output))

	// if the driver doesn't have return any version, it is really old, we force a upgrade.
	if len(v) == 0 {
		exit.WithCodeT(
			exit.Failure,
			"Please upgrade the '{{.driver_executable}}'. {{.documentation_url}}",
			out.V{"driver_executable": driverExecutable, "documentation_url": driverDocumentation},
		)
	}

	vmDriverVersion, err := semver.Make(v)
	if err != nil {
		out.WarningT("Error parsing vmDriver version: {{.error}}", out.V{"error": err})
		return
	}

	minikubeVersion, err := version.GetSemverVersion()
	if err != nil {
		out.WarningT("Error parsing minukube version: {{.error}}", out.V{"error": err})
		return
	}

	if vmDriverVersion.LT(minikubeVersion) {
		out.WarningT(
			"There's a new version for '{{.driver_executable}}'. Please consider upgrading. {{.documentation_url}}",
			out.V{"driver_executable": driverExecutable, "documentation_url": driverDocumentation},
		)
	}
}

// extractVMDriverVersion extracts the driver version.
// KVM and Hyperkit drivers support the 'version' command, that display the information as:
// version: vX.X.X
// commit: XXXX
// This method returns the version 'vX.X.X' or empty if the version isn't found.
func extractVMDriverVersion(s string) string {
	versionRegex := regexp.MustCompile(`version:(.*)`)
	matches := versionRegex.FindStringSubmatch(s)

	if len(matches) != 2 {
		return ""
	}

	v := strings.TrimSpace(matches[1])
	return strings.TrimPrefix(v, version.VersionPrefix)
}
