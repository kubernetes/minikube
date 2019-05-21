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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

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
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/machine"
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
	xhyveDiskDriver       = "xhyve-disk-driver"
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
	gpu                   = "gpu"
	hidden                = "hidden"
	embedCerts            = "embed-certs"
	noVTXCheck            = "no-vtx-check"
	downloadOnly          = "download-only"
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
	startCmd.Flags().Bool(keepContext, constants.DefaultKeepContext, "This will keep the existing kubectl context and will create a minikube context.")
	startCmd.Flags().Bool(createMount, false, "This will start the mount daemon and automatically mount files into minikube")
	startCmd.Flags().String(mountString, constants.DefaultMountDir+":"+constants.DefaultMountEndpoint, "The argument to pass the minikube mount command on start")
	startCmd.Flags().Bool(disableDriverMounts, false, "Disables the filesystem mounts provided by the hypervisors (vboxfs, xhyve-9p)")
	startCmd.Flags().String(isoURL, constants.DefaultISOURL, "Location of the minikube iso")
	startCmd.Flags().String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))
	startCmd.Flags().Int(memory, constants.DefaultMemory, "Amount of RAM allocated to the minikube VM in MB")
	startCmd.Flags().Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the minikube VM")
	startCmd.Flags().String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startCmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minikube VM (only supported with Virtualbox driver)")
	startCmd.Flags().String(hypervVirtualSwitch, "", "The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)")
	startCmd.Flags().String(kvmNetwork, "default", "The KVM network name. (only supported with KVM driver)")
	startCmd.Flags().String(xhyveDiskDriver, "ahci-hd", "The disk driver to use [ahci-hd|virtio-blk] (only supported with xhyve driver)")
	startCmd.Flags().StringSlice(nfsShare, []string{}, "Local folders to share with Guest via NFS mounts (Only supported on with hyperkit now)")
	startCmd.Flags().String(nfsSharesRoot, "/nfsshares", "Where to root the NFS Shares (defaults to /nfsshares, only supported with hyperkit now)")
	startCmd.Flags().StringArrayVar(&dockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().StringArrayVar(&dockerOpt, "docker-opt", nil, "Specify arbitrary flags to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().Int(apiServerPort, pkgutil.APIServerPort, "The apiserver listening port")
	startCmd.Flags().String(apiServerName, constants.APIServerName, "The apiserver name which is used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().StringArrayVar(&apiServerNames, "apiserver-names", nil, "A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().IPSliceVar(&apiServerIPs, "apiserver-ips", nil, "A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().String(dnsDomain, constants.ClusterDNSDomain, "The cluster dns domain name used in the kubernetes cluster")
	startCmd.Flags().String(serviceCIDR, pkgutil.DefaultServiceCIDR, "The CIDR to be used for service cluster IPs.")
	startCmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", nil, "Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.")
	startCmd.Flags().StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	startCmd.Flags().String(imageRepository, "", "Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to \"auto\" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers")
	startCmd.Flags().String(imageMirrorCountry, "", "Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn")
	startCmd.Flags().String(containerRuntime, "docker", "The container runtime to be used (docker, crio, containerd)")
	startCmd.Flags().String(criSocket, "", "The cri socket path to be used")
	startCmd.Flags().String(kubernetesVersion, constants.DefaultKubernetesVersion, "The kubernetes version that the minikube VM will use (ex: v1.2.3)")
	startCmd.Flags().String(networkPlugin, "", "The name of the network plugin")
	startCmd.Flags().Bool(enableDefaultCNI, false, "Enable the default CNI plugin (/etc/cni/net.d/k8s.conf). Used in conjunction with \"--network-plugin=cni\"")
	startCmd.Flags().String(featureGates, "", "A set of key=value pairs that describe feature gates for alpha/experimental features.")
	startCmd.Flags().Bool(downloadOnly, false, "If true, only download and cache files for later use - don't install or start anything.")
	startCmd.Flags().Bool(cacheImages, true, "If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --vm-driver=none.")
	startCmd.Flags().Var(&extraOptions, "extra-config",
		`A set of key=value pairs that describe configuration that may be passed to different components.
		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
		Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
		Valid kubeadm parameters: `+fmt.Sprintf("%s, %s", strings.Join(kubeadm.KubeadmExtraArgsWhitelist[kubeadm.KubeadmCmdParam], ", "), strings.Join(kubeadm.KubeadmExtraArgsWhitelist[kubeadm.KubeadmConfigParam], ",")))
	startCmd.Flags().String(uuid, "", "Provide VM UUID to restore MAC address (only supported with Hyperkit driver).")
	startCmd.Flags().String(vpnkitSock, "", "Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock.")
	startCmd.Flags().StringSlice(vsockPorts, []string{}, "List of guest VSock ports that should be exposed as sockets on the host (Only supported on with hyperkit now).")
	startCmd.Flags().Bool(gpu, false, "Enable experimental NVIDIA GPU support in minikube (works only with kvm2 driver on Linux)")
	startCmd.Flags().Bool(hidden, false, "Hide the hypervisor signature from the guest in minikube (works only with kvm2 driver on Linux)")
	startCmd.Flags().Bool(noVTXCheck, false, "Disable checking for the availability of hardware virtualization before the vm is started (virtualbox)")
	if err := viper.BindPFlags(startCmd.Flags()); err != nil {
		exit.WithError("unable to bind flags", err)
	}
	RootCmd.AddCommand(startCmd)
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a local kubernetes cluster",
	Long: `Starts a local kubernetes cluster using VM. This command
assumes you have already installed one of the VM drivers: virtualbox/parallels/vmwarefusion/kvm/xhyve/hyperv.`,
	Run: runStart,
}

// runStart handles the executes the flow of "minikube start"
func runStart(cmd *cobra.Command, args []string) {
	console.OutStyle("happy", "minikube %s on %s (%s)", version.GetVersion(), runtime.GOOS, runtime.GOARCH)
	validateConfig()

	oldConfig, err := cfg.Load()
	if err != nil && !os.IsNotExist(err) {
		exit.WithCode(exit.Data, "Unable to load config: %v", err)
	}
	k8sVersion, isUpgrade := validateKubernetesVersions(oldConfig)
	config, err := generateConfig(cmd, k8sVersion)
	if err != nil {
		exit.WithError("Failed to generate config", err)
	}

	// For non-"none", the ISO is required to boot, so block until it is downloaded
	if viper.GetString(vmDriver) != constants.DriverNone {
		if err := cluster.CacheISO(config.MachineConfig); err != nil {
			exit.WithError("Failed to cache ISO", err)
		}
	} else {
		// With "none", images are persistently stored in Docker, so internal caching isn't necessary.
		viper.Set(cacheImages, false)
	}

	// Now that the ISO is downloaded, pull images in the background while the VM boots.
	var cacheGroup errgroup.Group
	beginCacheImages(&cacheGroup, config.KubernetesConfig.ImageRepository, k8sVersion)

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, saveConfig must be called before startHost, and again afterwards when we know the IP.
	if err := saveConfig(config); err != nil {
		exit.WithError("Failed to save config", err)
	}

	m, err := machine.NewAPIClient()
	if err != nil {
		exit.WithError("Failed to get machine client", err)
	}

	// If --download-only, complete the remaining downloads and exit.
	if viper.GetBool(downloadOnly) {
		if err := doCacheBinaries(k8sVersion); err != nil {
			exit.WithError("Failed to cache binaries", err)
		}
		waitCacheImages(&cacheGroup)
		if err := CacheImagesInConfigFile(); err != nil {
			exit.WithError("Failed to cache images", err)
		}
		console.OutStyle("check", "Download complete!")
		return
	}

	host, preexisting := startHost(m, config.MachineConfig)

	ip := validateNetwork(host)
	// Makes minikube node ip to bypass http(s) proxy. since it is local traffic.
	err = proxy.ExcludeIP(ip)
	if err != nil {
		console.ErrStyle("Failed to set NO_PROXY Env. please Use `export NO_PROXY=$NO_PROXY,%s`.", ip)
	}
	// Save IP to configuration file for subsequent use
	config.KubernetesConfig.NodeIP = ip
	if err := saveConfig(config); err != nil {
		exit.WithError("Failed to save config", err)
	}
	runner, err := machine.CommandRunner(host)
	if err != nil {
		exit.WithError("Failed to get command runner", err)
	}

	cr := configureRuntimes(runner)
	version, _ := cr.Version()
	console.OutStyle(cr.Name(), "Configuring environment for Kubernetes %s on %s %s", k8sVersion, cr.Name(), version)
	for _, v := range dockerOpt {
		console.OutStyle("option", "opt %s", v)
	}
	for _, v := range dockerEnv {
		console.OutStyle("option", "env %s", v)
	}

	// prepareHostEnvironment uses the downloaded images, so we need to wait for background task completion.
	waitCacheImages(&cacheGroup)

	bs := prepareHostEnvironment(m, config.KubernetesConfig)

	// The kube config must be update must come before bootstrapping, otherwise health checks may use a stale IP
	kubeconfig := updateKubeConfig(host, &config)
	bootstrapCluster(bs, cr, runner, config.KubernetesConfig, preexisting, isUpgrade)
	configureMounts()
	if err = LoadCachedImagesInConfigFile(); err != nil {
		console.Failure("Unable to load cached images from config file.")
	}

	if config.MachineConfig.VMDriver == constants.DriverNone {
		console.OutStyle("starting-none", "Configuring local host environment ...")
		prepareNone()
	}

	if err := bs.WaitCluster(config.KubernetesConfig); err != nil {
		exit.WithError("Wait failed", err)
	}
	showKubectlConnectInfo(kubeconfig)

}

func showKubectlConnectInfo(kubeconfig *pkgutil.KubeConfigSetup) {
	if kubeconfig.KeepContext {
		console.OutStyle("kubectl", "To connect to this cluster, use: kubectl --context=%s", kubeconfig.ClusterName)
	} else {
		console.OutStyle("ready", "Done! kubectl is now configured to use %q", cfg.GetMachineName())
	}
	_, err := exec.LookPath("kubectl")
	if err != nil {
		console.OutStyle("tip", "For best results, install kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl/")
	}
}

func selectImageRepository(mirrorCountry string, k8sVersion string) (bool, string, error) {
	var tryCountries []string
	var fallback string

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

// validateConfig validates the supplied configuration against known bad combinations
func validateConfig() {
	diskSizeMB := pkgutil.CalculateDiskSizeInMB(viper.GetString(humanReadableDiskSize))
	if diskSizeMB < constants.MinimumDiskSizeMB {
		exit.WithCode(exit.Config, "Requested disk size (%dMB) is less than minimum of %dMB", diskSizeMB, constants.MinimumDiskSizeMB)
	}

	if viper.GetBool(gpu) && viper.GetString(vmDriver) != "kvm2" {
		exit.Usage("Sorry, the --gpu feature is currently only supported with --vm-driver=kvm2")
	}
	if viper.GetBool(hidden) && viper.GetString(vmDriver) != "kvm2" {
		exit.Usage("Sorry, the --hidden feature is currently only supported with --vm-driver=kvm2")
	}

	// check that kubeadm extra args contain only whitelisted parameters
	for param := range extraOptions.AsMap().Get(kubeadm.Kubeadm) {
		if !pkgutil.ContainsString(kubeadm.KubeadmExtraArgsWhitelist[kubeadm.KubeadmCmdParam], param) &&
			!pkgutil.ContainsString(kubeadm.KubeadmExtraArgsWhitelist[kubeadm.KubeadmConfigParam], param) {
			exit.Usage("Sorry, the kubeadm.%s parameter is currently not supported by --extra-config", param)
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
					dockerEnv = append(dockerEnv, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
	}

	repository := viper.GetString(imageRepository)
	mirrorCountry := strings.ToLower(viper.GetString(imageMirrorCountry))
	if strings.ToLower(repository) == "auto" || mirrorCountry != "" {
		console.OutStyle("connectivity", "checking main repository and mirrors for images")
		found, autoSelectedRepository, err := selectImageRepository(mirrorCountry, k8sVersion)
		if err != nil {
			exit.WithError("Failed to check main repository and mirrors for images for images", err)
		}

		if !found {
			if autoSelectedRepository == "" {
				exit.WithCode(exit.Failure, "None of known repositories is accessible. Consider specifying an alternative image repository with --image-repository flag")
			} else {
				console.Warning("None of known repositories in your location is accessible. Use %s as fallback.", autoSelectedRepository)
			}
		}

		repository = autoSelectedRepository
	}

	if repository != "" {
		console.OutStyle("success", "using image repository %s", repository)
	}

	cfg := cfg.Config{
		MachineConfig: cfg.MachineConfig{
			MinikubeISO:         viper.GetString(isoURL),
			Memory:              viper.GetInt(memory),
			CPUs:                viper.GetInt(cpus),
			DiskSize:            pkgutil.CalculateDiskSizeInMB(viper.GetString(humanReadableDiskSize)),
			VMDriver:            viper.GetString(vmDriver),
			ContainerRuntime:    viper.GetString(containerRuntime),
			HyperkitVpnKitSock:  viper.GetString(vpnkitSock),
			HyperkitVSockPorts:  viper.GetStringSlice(vsockPorts),
			XhyveDiskDriver:     viper.GetString(xhyveDiskDriver),
			NFSShare:            viper.GetStringSlice(nfsShare),
			NFSSharesRoot:       viper.GetString(nfsSharesRoot),
			DockerEnv:           dockerEnv,
			DockerOpt:           dockerOpt,
			InsecureRegistry:    insecureRegistry,
			RegistryMirror:      registryMirror,
			HostOnlyCIDR:        viper.GetString(hostOnlyCIDR),
			HypervVirtualSwitch: viper.GetString(hypervVirtualSwitch),
			KvmNetwork:          viper.GetString(kvmNetwork),
			Downloader:          pkgutil.DefaultDownloader{},
			DisableDriverMounts: viper.GetBool(disableDriverMounts),
			UUID:                viper.GetString(uuid),
			GPU:                 viper.GetBool(gpu),
			Hidden:              viper.GetBool(hidden),
			NoVTXCheck:          viper.GetBool(noVTXCheck),
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

// prepareNone prepares the user and host for the joy of the "none" driver
func prepareNone() {
	if viper.GetBool(cfg.WantNoneDriverWarning) {
		console.OutLn("")
		console.Warning("The 'none' driver provides limited isolation and may reduce system security and reliability.")
		console.Warning("For more information, see:")
		console.OutStyle("url", "https://github.com/kubernetes/minikube/blob/master/docs/vmdriver-none.md")
		console.OutLn("")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		home := os.Getenv("HOME")
		console.Warning("kubectl and minikube configuration will be stored in %s", home)
		console.Warning("To use kubectl or minikube commands as your own user, you may")
		console.Warning("need to relocate them. For example, to overwrite your own settings:")

		console.OutLn("")
		console.OutStyle("command", "sudo mv %s/.kube %s/.minikube $HOME", home, home)
		console.OutStyle("command", "sudo chown -R $USER $HOME/.kube $HOME/.minikube")
		console.OutLn("")

		console.OutStyle("tip", "This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true")
	}

	if err := pkgutil.MaybeChownDirRecursiveToMinikubeUser(constants.GetMinipath()); err != nil {
		exit.WithCode(exit.Permissions, "Failed to chown %s: %v", constants.GetMinipath(), err)
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
				console.OutStyle("internet", "Found network options:")
				optSeen = true
			}
			console.OutStyle("option", "%s=%s", k, v)
			ipExcluded := proxy.IsIPExcluded(ip) // Skip warning if minikube ip is already in NO_PROXY
			if (k == "HTTP_PROXY" || k == "HTTPS_PROXY") && !ipExcluded && !warnedOnce {
				console.Warning("You appear to be using a proxy, but your NO_PROXY environment does not include the minikube IP (%s). Please see https://github.com/kubernetes/minikube/blob/master/docs/http_proxy.md for more details", ip)
				warnedOnce = true
			}
		}
	}

	// Here is where we should be checking connectivity to/from the VM
	return ip
}

// validateKubernetesVersions ensures that the requested version is reasonable
func validateKubernetesVersions(old *cfg.Config) (string, bool) {
	nv := viper.GetString(kubernetesVersion)
	isUpgrade := false
	if nv == "" {
		nv = constants.DefaultKubernetesVersion
	}
	nvs, err := semver.Make(strings.TrimPrefix(nv, version.VersionPrefix))
	if err != nil {
		exit.WithCode(exit.Data, "Unable to parse %q: %v", nv, err)
	}

	if old == nil || old.KubernetesConfig.KubernetesVersion == "" {
		return nv, isUpgrade
	}

	ovs, err := semver.Make(strings.TrimPrefix(old.KubernetesConfig.KubernetesVersion, version.VersionPrefix))
	if err != nil {
		glog.Errorf("Error parsing old version %q: %v", old.KubernetesConfig.KubernetesVersion, err)
	}

	if nvs.LT(ovs) {
		nv = version.VersionPrefix + ovs.String()
		console.ErrStyle("conflict", "Kubernetes downgrade is not supported, will continue to use %v", nv)
		return nv, isUpgrade
	}
	if nvs.GT(ovs) {
		console.OutStyle("thumbs-up", "minikube will upgrade the local cluster from Kubernetes %s to %s", ovs, nvs)
		isUpgrade = true
	}
	return nv, isUpgrade
}

// prepareHostEnvironment adds any requested files into the VM before Kubernetes is started
func prepareHostEnvironment(api libmachine.API, kc cfg.KubernetesConfig) bootstrapper.Bootstrapper {
	bs, err := GetClusterBootstrapper(api, viper.GetString(cmdcfg.Bootstrapper))
	if err != nil {
		exit.WithError("Failed to get bootstrapper", err)
	}
	for _, eo := range extraOptions {
		console.OutStyle("option", "%s.%s=%s", eo.Component, eo.Key, eo.Value)
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
		exit.WithError(fmt.Sprintf("Failed runtime for %+v", config), err)
	}

	err = cr.Enable()
	if err != nil {
		exit.WithError("Failed to enable container runtime", err)
	}

	return cr
}

// bootstrapCluster starts Kubernetes using the chosen bootstrapper
func bootstrapCluster(bs bootstrapper.Bootstrapper, r cruntime.Manager, runner bootstrapper.CommandRunner, kc cfg.KubernetesConfig, preexisting bool, isUpgrade bool) {
	// hum. bootstrapper.Bootstrapper should probably have a Name function.
	bsName := viper.GetString(cmdcfg.Bootstrapper)

	if isUpgrade || !preexisting {
		console.OutStyle("pulling", "Pulling images ...")
		if err := bs.PullImages(kc); err != nil {
			console.OutStyle("failure", "Unable to pull images, which may be OK: %v", err)
		}
	}

	if preexisting {
		console.OutStyle("restarting", "Relaunching Kubernetes %s using %s ... ", kc.KubernetesVersion, bsName)
		if err := bs.RestartCluster(kc); err != nil {
			exit.WithLogEntries("Error restarting cluster", err, logs.FindProblems(r, bs, runner))
		}
		return
	}

	console.OutStyle("launch", "Launching Kubernetes ... ")
	if err := bs.StartCluster(kc); err != nil {
		exit.WithLogEntries("Error starting cluster", err, logs.FindProblems(r, bs, runner))
	}
}

// configureMounts configures any requested filesystem mounts
func configureMounts() {
	if !viper.GetBool(createMount) {
		return
	}

	console.OutStyle("mounting", "Creating mount %s ...", viper.GetString(mountString))
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
func saveConfig(clusterConfig cfg.Config) error {
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
