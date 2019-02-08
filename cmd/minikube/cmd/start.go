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
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	cmdutil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/cluster"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/machine"
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
	NFSSharesRoot         = "nfs-shares-root"
	NFSShare              = "nfs-share"
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
	mountString           = "mount-string"
	disableDriverMounts   = "disable-driver-mounts"
	cacheImages           = "cache-images"
	uuid                  = "uuid"
	vpnkitSock            = "hyperkit-vpnkit-sock"
	vsockPorts            = "hyperkit-vsock-ports"
	gpu                   = "gpu"
	embedCerts            = "embed-certs"
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
	startCmd.Flags().String(isoURL, constants.DefaultIsoUrl, "Location of the minikube iso")
	startCmd.Flags().String(vmDriver, constants.DefaultVMDriver, fmt.Sprintf("VM driver is one of: %v", constants.SupportedVMDrivers))
	startCmd.Flags().Int(memory, constants.DefaultMemory, "Amount of RAM allocated to the minikube VM in MB")
	startCmd.Flags().Int(cpus, constants.DefaultCPUS, "Number of CPUs allocated to the minikube VM")
	startCmd.Flags().String(humanReadableDiskSize, constants.DefaultDiskSize, "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g)")
	startCmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minikube VM (only supported with Virtualbox driver)")
	startCmd.Flags().String(hypervVirtualSwitch, "", "The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)")
	startCmd.Flags().String(kvmNetwork, "default", "The KVM network name. (only supported with KVM driver)")
	startCmd.Flags().String(xhyveDiskDriver, "ahci-hd", "The disk driver to use [ahci-hd|virtio-blk] (only supported with xhyve driver)")
	startCmd.Flags().StringSlice(NFSShare, []string{}, "Local folders to share with Guest via NFS mounts (Only supported on with hyperkit now)")
	startCmd.Flags().String(NFSSharesRoot, "/nfsshares", "Where to root the NFS Shares (defaults to /nfsshares, only supported with hyperkit now)")
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
	startCmd.Flags().String(containerRuntime, "docker", "The container runtime to be used (docker, crio, containerd)")
	startCmd.Flags().String(criSocket, "", "The cri socket path to be used")
	startCmd.Flags().String(kubernetesVersion, constants.DefaultKubernetesVersion, "The kubernetes version that the minikube VM will use (ex: v1.2.3)")
	startCmd.Flags().String(networkPlugin, "", "The name of the network plugin")
	startCmd.Flags().Bool(enableDefaultCNI, false, "Enable the default CNI plugin (/etc/cni/net.d/k8s.conf). Used in conjunction with \"--network-plugin=cni\"")
	startCmd.Flags().String(featureGates, "", "A set of key=value pairs that describe feature gates for alpha/experimental features.")
	// TODO(tstromberg): Flip cacheImages to true once it can be stabilized
	startCmd.Flags().Bool(cacheImages, false, "If true, cache docker images for the current bootstrapper and load them into the machine.")
	startCmd.Flags().Var(&extraOptions, "extra-config",
		`A set of key=value pairs that describe configuration that may be passed to different components.
		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
		Valid components are: kubelet, apiserver, controller-manager, etcd, proxy, scheduler.`)
	startCmd.Flags().String(uuid, "", "Provide VM UUID to restore MAC address (only supported with Hyperkit driver).")
	startCmd.Flags().String(vpnkitSock, "", "Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock.")
	startCmd.Flags().StringSlice(vsockPorts, []string{}, "List of guest VSock ports that should be exposed as sockets on the host (Only supported on with hyperkit now).")
	startCmd.Flags().Bool(gpu, false, "Enable experimental NVIDIA GPU support in minikube (works only with kvm2 driver on Linux)")
	viper.BindPFlags(startCmd.Flags())
	RootCmd.AddCommand(startCmd)
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a local kubernetes cluster",
	Long: `Starts a local kubernetes cluster using VM. This command
assumes you have already installed one of the VM drivers: virtualbox/vmwarefusion/kvm/xhyve/hyperv.`,
	Run: runStart,
}

// runStart handles the executes the flow of "minikube start"
func runStart(cmd *cobra.Command, args []string) {
	console.OutStyle("happy", "minikube %s on %s (%s)", version.GetVersion(), runtime.GOOS, runtime.GOARCH)
	validateConfig()

	oldConfig, err := cfg.Load()
	if err != nil && !os.IsNotExist(err) {
		fatalExit("Unable to load config: %v", err)
	}
	kVersion := validateKubernetesVersions(oldConfig)
	var cacheGroup errgroup.Group
	beginCacheImages(&cacheGroup, kVersion)
	config := generateConfig(kVersion)

	m, err := machine.NewAPIClient()
	if err != nil {
		reportErrAndExit("Failed to get machine client: %v", err)
	}

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot :(
	if err := saveConfig(config); err != nil {
		reportErrAndExit("Failed to save config", err)
	}

	host, preexisting := startHost(m, config.MachineConfig)

	ip := validateNetwork(host)
	// Save IP to configuration file for subsequent use
	config.KubernetesConfig.NodeIP = ip
	if err := saveConfig(config); err != nil {
		reportErrAndExit("Failed to save config", err)
	}

	configureRuntimes(host)
	bs := prepareHostEnvironment(m, config.KubernetesConfig)
	waitCacheImages(&cacheGroup)

	// The kube config must be update must come before bootstrapping, otherwise health checks may use a stale IP
	kubeconfig := updateKubeConfig(host, &config)
	bootstrapCluster(bs, config.KubernetesConfig, preexisting)
	validateCluster(bs, ip)
	configureMounts()
	if err = LoadCachedImagesInConfigFile(); err != nil {
		console.Failure("Unable to load cached images from config file.")
	}

	if kubeconfig.KeepContext {
		console.OutStyle("kubectl", "To connect to this cluster, use: kubectl --context=%s", kubeconfig.ClusterName)
	} else {
		console.OutStyle("kubectl", "kubectl is now configured to use %q", cfg.GetMachineName())
	}
	console.OutStyle("ready", "Done! Thank you for using minikube!")
}

// validateConfig validates the supplied configuration against known bad combinations
func validateConfig() {
	diskSizeMB := pkgutil.CalculateDiskSizeInMB(viper.GetString(humanReadableDiskSize))
	if diskSizeMB < constants.MinimumDiskSizeMB {
		fatalExit("Requested disk size (%dMB) is less than minimum of %dMB", diskSizeMB, constants.MinimumDiskSizeMB)
	}

	if viper.GetBool(gpu) && viper.GetString(vmDriver) != "kvm2" {
		fatalExit("Sorry, the --gpu feature is currently only supported with --vm-driver=kvm2")
	}
}

// beginCacheImages caches Docker images in the background
func beginCacheImages(g *errgroup.Group, kVersion string) {
	if !viper.GetBool(cacheImages) {
		return
	}
	console.OutStyle("caching", "Caching images in the background ...")
	g.Go(func() error {
		return machine.CacheImagesForBootstrapper(kVersion, viper.GetString(cmdcfg.Bootstrapper))
	})
}

// generateConfig generates cfg.Config based on flags and supplied arguments
func generateConfig(kVersion string) cfg.Config {
	// Write profile cluster configuration to file
	return cfg.Config{
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
			NFSShare:            viper.GetStringSlice(NFSShare),
			NFSSharesRoot:       viper.GetString(NFSSharesRoot),
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
		},
		KubernetesConfig: cfg.KubernetesConfig{
			KubernetesVersion:      kVersion,
			NodePort:               viper.GetInt(apiServerPort),
			NodeName:               constants.DefaultNodeName,
			APIServerName:          viper.GetString(apiServerName),
			APIServerNames:         apiServerNames,
			APIServerIPs:           apiServerIPs,
			DNSDomain:              viper.GetString(dnsDomain),
			FeatureGates:           viper.GetString(featureGates),
			ContainerRuntime:       viper.GetString(containerRuntime),
			CRISocket:              viper.GetString(criSocket),
			NetworkPlugin:          viper.GetString(networkPlugin),
			ServiceCIDR:            viper.GetString(serviceCIDR),
			ExtraOptions:           extraOptions,
			ShouldLoadCachedImages: viper.GetBool(cacheImages),
			EnableDefaultCNI:       viper.GetBool(enableDefaultCNI),
		},
	}
}

// prepareNone prepares the user and host for the joy of the "none" driver
func prepareNone() {
	if viper.GetBool(cfg.WantNoneDriverWarning) {
		console.ErrLn(`===================
WARNING: IT IS RECOMMENDED NOT TO RUN THE NONE DRIVER ON PERSONAL WORKSTATIONS
The 'none' driver will run an insecure kubernetes apiserver as root that may leave the host vulnerable to CSRF attacks` + "\n")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		console.Fatal(`When using the none driver, the kubectl config and credentials generated will be root owned and will appear in the root home directory.
You will need to move the files to the appropriate location and then set the correct permissions.  An example of this is below:

sudo mv /root/.kube $HOME/.kube # this will write over any previous configuration
sudo chown -R $USER:$USER $HOME/.kube

sudo mv /root/.minikube $HOME/.minikube # this will write over any previous configuration
sudo chown -R $USER:$USER $HOME/.minikube

This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true`)
	}

	if err := pkgutil.MaybeChownDirRecursiveToMinikubeUser(constants.GetMinipath()); err != nil {
		fatalExit("Failed to chown %s: %v", constants.GetMinipath(), err)
	}
}

// startHost starts a new minikube host using a VM or None
func startHost(api libmachine.API, mc cfg.MachineConfig) (*host.Host, bool) {
	exists, err := api.Exists(cfg.GetMachineName())
	if err != nil {
		reportErrAndExit("Failed to check if machine exists", err)
	}
	if mc.VMDriver == constants.DriverNone {
		console.OutStyle("starting-none", "Configuring local host environment ...")
		prepareNone()
	} else {
		if exists {
			if cfg.GetMachineName() == constants.DefaultMachineName {
				console.OutStyle("tip", "Tip: To create another VM, use 'minikube start -p <new name>' or use 'minikube delete' to delete this one.")
			}
			console.OutStyle("waiting", "Waiting for existing %q VM to start ...", cfg.GetMachineName())
		} else {
			console.OutStyle("starting-vm", "Creating %s VM (CPUs=%d, Memory=%dMB, Disk=%dMB) ...", mc.VMDriver, mc.CPUs, mc.Memory, mc.DiskSize)
		}
	}

	var host *host.Host
	start := func() (err error) {
		host, err = cluster.StartHost(api, mc)
		if err != nil {
			glog.Infof("StartHost: %v", err)
		}
		return err
	}
	if err = pkgutil.RetryAfter(3, start, 2*time.Second); err != nil {
		reportErrAndExit("Unable to start VM", err)
	}
	return host, exists
}

// validateNetwork tries to catch network problems as soon as possible
func validateNetwork(h *host.Host) string {
	ip, err := h.Driver.GetIP()
	if err != nil {
		reportErrAndExit("Unable to get VM IP address", err)
	}
	console.OutStyle("connectivity", "%q IP address is %s.", cfg.GetMachineName(), ip)

	// Here is where we should be checking connectivity to/from the VM
	for _, k := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
		if v := os.Getenv(k); v != "" {
			console.OutStyle("option", "%s=%s", k, v)
		}
	}
	return ip
}

// validateKubernetesVersions ensures that the requested version is reasonable
func validateKubernetesVersions(old *cfg.Config) string {
	nv := viper.GetString(kubernetesVersion)
	if nv == "" {
		nv = constants.DefaultKubernetesVersion
	}
	nvs, err := semver.Make(strings.TrimPrefix(nv, version.VersionPrefix))
	if err != nil {
		fatalExit("Unable to parse %q: %v", nv, err)
	}

	if old == nil || old.KubernetesConfig.KubernetesVersion == "" {
		return nv
	}

	ovs, err := semver.Make(strings.TrimPrefix(old.KubernetesConfig.KubernetesVersion, version.VersionPrefix))
	if err != nil {
		glog.Errorf("Error parsing old version %q: %v", old.KubernetesConfig.KubernetesVersion, err)
	}

	if nvs.LT(ovs) {
		nv = version.VersionPrefix + ovs.String()
		console.ErrStyle("conflict", "Kubernetes downgrade is not supported, will continue to use %v", nv)
		return nv
	}
	if nvs.GT(ovs) {
		console.OutStyle("thumbs-up", "minikube will upgrade the local cluster from Kubernetes %s to %s", ovs, nvs)
	}
	return nv
}

// prepareHostEnvironment adds any requested files into the VM before Kubernetes is started
func prepareHostEnvironment(api libmachine.API, kc cfg.KubernetesConfig) bootstrapper.Bootstrapper {
	bs, err := GetClusterBootstrapper(api, viper.GetString(cmdcfg.Bootstrapper))
	if err != nil {
		reportErrAndExit("Failed to get bootstrapper", err)
	}
	console.OutStyle("copying", "Preparing Kubernetes environment ...")
	for _, eo := range extraOptions {
		console.OutStyle("option", "%s.%s=%s", eo.Component, eo.Key, eo.Value)
	}
	// Loads cached images, generates config files, download binaries
	if err := bs.UpdateCluster(kc); err != nil {
		reportErrAndExit("Failed to update cluster", err)
	}
	if err := bs.SetupCerts(kc); err != nil {
		reportErrAndExit("Failed to setup certs", err)
	}
	return bs
}

// updateKubeConfig sets up kubectl
func updateKubeConfig(h *host.Host, c *cfg.Config) *kubeconfig.KubeConfigSetup {
	addr, err := h.Driver.GetURL()
	if err != nil {
		reportErrAndExit("Failed to get driver URL", err)
	}
	addr = strings.Replace(addr, "tcp://", "https://", -1)
	addr = strings.Replace(addr, ":2376", ":"+strconv.Itoa(c.KubernetesConfig.NodePort), -1)

	kcs := &kubeconfig.KubeConfigSetup{
		ClusterName:          cfg.GetMachineName(),
		ClusterServerAddress: addr,
		ClientCertificate:    constants.MakeMiniPath("client.crt"),
		ClientKey:            constants.MakeMiniPath("client.key"),
		CertificateAuthority: constants.MakeMiniPath("ca.crt"),
		KeepContext:          viper.GetBool(keepContext),
		EmbedCerts:           viper.GetBool(embedCerts),
	}
	kcs.SetKubeConfigFile(cmdutil.GetKubeConfigPath())
	if err := kubeconfig.SetupKubeConfig(kcs); err != nil {
		reportErrAndExit("Failed to setup kubeconfig", err)
	}
	return kcs
}

// configureRuntimes does what needs to happen to get a runtime going.
func configureRuntimes(h *host.Host) {
	runner, err := machine.CommandRunner(h)
	if err != nil {
		reportErrAndExit("Failed to get command runner", err)
	}

	config := cruntime.Config{Type: viper.GetString(containerRuntime), Runner: runner}
	cr, err := cruntime.New(config)
	if err != nil {
		reportErrAndExit(fmt.Sprintf("Failed runtime for %+v", config), err)
		cmdutil.MaybeReportErrorAndExit(err)
	}
	console.OutStyle(cr.Name(), "Configuring %s as your container runtime ...", cr.Name())
	for _, v := range dockerOpt {
		console.OutStyle("option", "opt %s", v)
	}
	for _, v := range dockerEnv {
		console.OutStyle("option", "env %s", v)
	}

	err = cr.Enable()
	if err != nil {
		cmdutil.MaybeReportErrorAndExit(err)
	}

}

// waitCacheImages blocks until the image cache jobs complete
func waitCacheImages(g *errgroup.Group) {
	if !viper.GetBool(cacheImages) {
		return
	}
	console.OutStyle("waiting", "Waiting for image caching to complete ...")
	if err := g.Wait(); err != nil {
		glog.Errorln("Error caching images: ", err)
	}
}

// bootstrapCluster starts Kubernetes using the chosen bootstrapper
func bootstrapCluster(bs bootstrapper.Bootstrapper, kc cfg.KubernetesConfig, preexisting bool) {
	console.OutStyle("pulling", "Pulling images used by Kubernetes %s ...", kc.KubernetesVersion)
	if err := bs.PullImages(kc); err != nil {
		fmt.Printf("Unable to pull images, which may be OK: %v", err)
	}
	// hum. bootstrapper.Bootstrapper should probably have a Name function.
	bsName := viper.GetString(cmdcfg.Bootstrapper)

	if preexisting {
		console.OutStyle("restarting", "Relaunching Kubernetes %s using %s ... ", kc.KubernetesVersion, bsName)
		if err := bs.RestartCluster(kc); err != nil {
			reportErrAndExit("Error restarting cluster", err)
		}
		return
	}

	console.OutStyle("launch", "Launching Kubernetes %s using %s ... ", kc.KubernetesVersion, bsName)
	if err := bs.StartCluster(kc); err != nil {
		reportErrAndExit("Error starting cluster", err)
	}
}

// validateCluster validates that the cluster is well-configured and healthy
func validateCluster(bs bootstrapper.Bootstrapper, ip string) {
	console.OutStyle("verifying-noline", "Verifying component health ...")
	kStat := func() (err error) {
		st, err := bs.GetKubeletStatus()
		console.Out(".")
		if err != nil || st != state.Running.String() {
			return &pkgutil.RetriableError{Err: fmt.Errorf("kubelet unhealthy: %v: %s", err, st)}
		}
		return nil
	}
	err := pkgutil.RetryAfter(20, kStat, 3*time.Second)
	if err != nil {
		reportErrAndExit("kubelet checks failed", err)
	}
	aStat := func() (err error) {
		st, err := bs.GetApiServerStatus(net.ParseIP(ip))
		console.Out(".")
		if err != nil || st != state.Running.String() {
			return &pkgutil.RetriableError{Err: fmt.Errorf("apiserver status=%s err=%v", st, err)}
		}
		return nil
	}

	err = pkgutil.RetryAfter(30, aStat, 10*time.Second)
	if err != nil {
		reportErrAndExit("apiserver checks failed", err)
	}
	console.OutLn("")
}

// configureMounts configures any requested filesystem mounts
func configureMounts() {
	if !viper.GetBool(createMount) {
		return
	}

	console.OutStyle("mount", "Creating mount %s ...", viper.GetString(mountString))
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
		glog.Errorf("Error running command minikube mount %v", err)
		cmdutil.MaybeReportErrorAndExit(err)
	}
	if err := ioutil.WriteFile(filepath.Join(constants.GetMinipath(), constants.MountProcessFileName), []byte(strconv.Itoa(mountCmd.Process.Pid)), 0644); err != nil {
		glog.Errorf("Error writing mount process pid to file: %v", err)
		cmdutil.MaybeReportErrorAndExit(err)
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

// fatalExit is a shortcut for outputting a failure message and exiting.
func fatalExit(format string, a ...interface{}) {
	// use Warning because Error will display a duplicate message
	glog.Warningf(format, a...)
	console.Fatal(format, a...)
	os.Exit(1)
}

// reportFatalExit is a shortcut for outputting an error, reporting it, and exiting.
func reportErrAndExit(msg string, err error) {
	console.Fatal(msg+": %v", err)
	cmdutil.MaybeReportErrorAndExit(err)
	os.Exit(1)
}
