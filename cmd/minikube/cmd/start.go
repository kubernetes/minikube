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
	"math"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/golang/glog"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	gopshost "github.com/shirou/gopsutil/host"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/notify"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/translate"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/lock"
	"k8s.io/minikube/pkg/util/retry"
	"k8s.io/minikube/pkg/version"
)

const (
	isoURL                  = "iso-url"
	memory                  = "memory"
	cpus                    = "cpus"
	humanReadableDiskSize   = "disk-size"
	nfsSharesRoot           = "nfs-shares-root"
	nfsShare                = "nfs-share"
	kubernetesVersion       = "kubernetes-version"
	hostOnlyCIDR            = "host-only-cidr"
	containerRuntime        = "container-runtime"
	criSocket               = "cri-socket"
	networkPlugin           = "network-plugin"
	enableDefaultCNI        = "enable-default-cni"
	hypervVirtualSwitch     = "hyperv-virtual-switch"
	hypervUseExternalSwitch = "hyperv-use-external-switch"
	hypervExternalAdapter   = "hyperv-external-adapter"
	kvmNetwork              = "kvm-network"
	kvmQemuURI              = "kvm-qemu-uri"
	kvmGPU                  = "kvm-gpu"
	kvmHidden               = "kvm-hidden"
	minikubeEnvPrefix       = "MINIKUBE"
	defaultMemorySize       = "2000mb"
	installAddons           = "install-addons"
	defaultDiskSize         = "20000mb"
	keepContext             = "keep-context"
	createMount             = "mount"
	featureGates            = "feature-gates"
	apiServerName           = "apiserver-name"
	apiServerPort           = "apiserver-port"
	dnsDomain               = "dns-domain"
	serviceCIDR             = "service-cluster-ip-range"
	imageRepository         = "image-repository"
	imageMirrorCountry      = "image-mirror-country"
	mountString             = "mount-string"
	disableDriverMounts     = "disable-driver-mounts"
	cacheImages             = "cache-images"
	uuid                    = "uuid"
	vpnkitSock              = "hyperkit-vpnkit-sock"
	vsockPorts              = "hyperkit-vsock-ports"
	embedCerts              = "embed-certs"
	noVTXCheck              = "no-vtx-check"
	downloadOnly            = "download-only"
	dnsProxy                = "dns-proxy"
	hostDNSResolver         = "host-dns-resolver"
	waitUntilHealthy        = "wait"
	force                   = "force"
	dryRun                  = "dry-run"
	interactive             = "interactive"
	waitTimeout             = "wait-timeout"
	nativeSSH               = "native-ssh"
	minimumMemorySize       = "1024mb"
	minimumCPUS             = 2
	minimumDiskSize         = "2000mb"
	autoUpdate              = "auto-update-drivers"
	hostOnlyNicType         = "host-only-nic-type"
	natNicType              = "nat-nic-type"
)

var (
	registryMirror   []string
	dockerEnv        []string
	dockerOpt        []string
	insecureRegistry []string
	apiServerNames   []string
	addonList        []string
	apiServerIPs     []net.IP
	extraOptions     config.ExtraOptionSlice
)

func init() {
	initMinikubeFlags()
	initKubernetesFlags()
	initDriverFlags()
	initNetworkingFlags()
	if err := viper.BindPFlags(startCmd.Flags()); err != nil {
		exit.WithError("unable to bind flags", err)
	}
}

// initMinikubeFlags includes commandline flags for minikube.
func initMinikubeFlags() {
	viper.SetEnvPrefix(minikubeEnvPrefix)
	// Replaces '-' in flags with '_' in env variables
	// e.g. iso-url => $ENVPREFIX_ISO_URL
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	startCmd.Flags().Bool(force, false, "Force minikube to perform possibly dangerous operations")
	startCmd.Flags().Bool(interactive, true, "Allow user prompts for more information")
	startCmd.Flags().Bool(dryRun, false, "dry-run mode. Validates configuration, but does not mutate system state")

	startCmd.Flags().Int(cpus, 2, "Number of CPUs allocated to the minikube VM.")
	startCmd.Flags().String(memory, defaultMemorySize, "Amount of RAM allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g).")
	startCmd.Flags().String(humanReadableDiskSize, defaultDiskSize, "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g).")
	startCmd.Flags().Bool(downloadOnly, false, "If true, only download and cache files for later use - don't install or start anything.")
	startCmd.Flags().Bool(cacheImages, true, "If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --vm-driver=none.")
	startCmd.Flags().String(isoURL, constants.DefaultISOURL, "Location of the minikube iso.")
	startCmd.Flags().Bool(keepContext, false, "This will keep the existing kubectl context and will create a minikube context.")
	startCmd.Flags().Bool(embedCerts, false, "if true, will embed the certs in kubeconfig.")
	startCmd.Flags().String(containerRuntime, "docker", "The container runtime to be used (docker, crio, containerd).")
	startCmd.Flags().Bool(createMount, false, "This will start the mount daemon and automatically mount files into minikube.")
	startCmd.Flags().String(mountString, constants.DefaultMountDir+":/minikube-host", "The argument to pass the minikube mount command on start.")
	startCmd.Flags().StringArrayVar(&addonList, "addons", nil, "Enable addons. see `minikube addons list` for a list of valid addon names.")
	startCmd.Flags().String(criSocket, "", "The cri socket path to be used.")
	startCmd.Flags().String(networkPlugin, "", "The name of the network plugin.")
	startCmd.Flags().Bool(enableDefaultCNI, false, "Enable the default CNI plugin (/etc/cni/net.d/k8s.conf). Used in conjunction with \"--network-plugin=cni\".")
	startCmd.Flags().Bool(waitUntilHealthy, true, "Block until the apiserver is servicing API requests")
	startCmd.Flags().Duration(waitTimeout, 6*time.Minute, "max time to wait per Kubernetes core services to be healthy.")
	startCmd.Flags().Bool(nativeSSH, true, "Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'.")
	startCmd.Flags().Bool(autoUpdate, true, "If set, automatically updates drivers to the latest version. Defaults to true.")
	startCmd.Flags().Bool(installAddons, true, "If set, install addons. Defaults to true.")
}

// initKubernetesFlags inits the commandline flags for kubernetes related options
func initKubernetesFlags() {
	startCmd.Flags().String(kubernetesVersion, "", "The kubernetes version that the minikube VM will use (ex: v1.2.3)")
	startCmd.Flags().Var(&extraOptions, "extra-config",
		`A set of key=value pairs that describe configuration that may be passed to different components.
		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
		Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
		Valid kubeadm parameters: `+fmt.Sprintf("%s, %s", strings.Join(bsutil.KubeadmExtraArgsWhitelist[bsutil.KubeadmCmdParam], ", "), strings.Join(bsutil.KubeadmExtraArgsWhitelist[bsutil.KubeadmConfigParam], ",")))
	startCmd.Flags().String(featureGates, "", "A set of key=value pairs that describe feature gates for alpha/experimental features.")
	startCmd.Flags().String(dnsDomain, constants.ClusterDNSDomain, "The cluster dns domain name used in the kubernetes cluster")
	startCmd.Flags().Int(apiServerPort, constants.APIServerPort, "The apiserver listening port")
	startCmd.Flags().String(apiServerName, constants.APIServerName, "The apiserver name which is used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().StringArrayVar(&apiServerNames, "apiserver-names", nil, "A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().IPSliceVar(&apiServerIPs, "apiserver-ips", nil, "A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
}

// initDriverFlags inits the commandline flags for vm drivers
func initDriverFlags() {
	startCmd.Flags().String("vm-driver", "", fmt.Sprintf("Driver is one of: %v (defaults to auto-detect)", driver.DisplaySupportedDrivers()))
	startCmd.Flags().Bool(disableDriverMounts, false, "Disables the filesystem mounts provided by the hypervisors")

	// kvm2
	startCmd.Flags().String(kvmNetwork, "default", "The KVM network name. (kvm2 driver only)")
	startCmd.Flags().String(kvmQemuURI, "qemu:///system", "The KVM QEMU connection URI. (kvm2 driver only)")
	startCmd.Flags().Bool(kvmGPU, false, "Enable experimental NVIDIA GPU support in minikube")
	startCmd.Flags().Bool(kvmHidden, false, "Hide the hypervisor signature from the guest in minikube (kvm2 driver only)")

	// virtualbox
	startCmd.Flags().String(hostOnlyCIDR, "192.168.99.1/24", "The CIDR to be used for the minikube VM (virtualbox driver only)")
	startCmd.Flags().Bool(dnsProxy, false, "Enable proxy for NAT DNS requests (virtualbox driver only)")
	startCmd.Flags().Bool(hostDNSResolver, true, "Enable host resolver for NAT DNS requests (virtualbox driver only)")
	startCmd.Flags().Bool(noVTXCheck, false, "Disable checking for the availability of hardware virtualization before the vm is started (virtualbox driver only)")
	startCmd.Flags().String(hostOnlyNicType, "virtio", "NIC Type used for host only network. One of Am79C970A, Am79C973, 82540EM, 82543GC, 82545EM, or virtio (virtualbox driver only)")
	startCmd.Flags().String(natNicType, "virtio", "NIC Type used for host only network. One of Am79C970A, Am79C973, 82540EM, 82543GC, 82545EM, or virtio (virtualbox driver only)")

	// hyperkit
	startCmd.Flags().StringSlice(vsockPorts, []string{}, "List of guest VSock ports that should be exposed as sockets on the host (hyperkit driver only)")
	startCmd.Flags().String(uuid, "", "Provide VM UUID to restore MAC address (hyperkit driver only)")
	startCmd.Flags().String(vpnkitSock, "", "Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock (hyperkit driver only)")
	startCmd.Flags().StringSlice(nfsShare, []string{}, "Local folders to share with Guest via NFS mounts (hyperkit driver only)")
	startCmd.Flags().String(nfsSharesRoot, "/nfsshares", "Where to root the NFS Shares, defaults to /nfsshares (hyperkit driver only)")

	// hyperv
	startCmd.Flags().String(hypervVirtualSwitch, "", "The hyperv virtual switch name. Defaults to first found. (hyperv driver only)")
	startCmd.Flags().Bool(hypervUseExternalSwitch, false, "Whether to use external switch over Default Switch if virtual switch not explicitly specified. (hyperv driver only)")
	startCmd.Flags().String(hypervExternalAdapter, "", "External Adapter on which external switch will be created if no external switch is found. (hyperv driver only)")
}

// initNetworkingFlags inits the commandline flags for connectivity related flags for start
func initNetworkingFlags() {
	startCmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", nil, "Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.")
	startCmd.Flags().StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	startCmd.Flags().String(imageRepository, "", "Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to \"auto\" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers")
	startCmd.Flags().String(imageMirrorCountry, "", "Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn.")
	startCmd.Flags().String(serviceCIDR, constants.DefaultServiceCIDR, "The CIDR to be used for service cluster IPs.")
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

// platform generates a user-readable platform message
func platform() string {
	var s strings.Builder

	// Show the distro version if possible
	hi, err := gopshost.Info()
	if err == nil {
		s.WriteString(fmt.Sprintf("%s %s", strings.Title(hi.Platform), hi.PlatformVersion))
		glog.Infof("hostinfo: %+v", hi)
	} else {
		glog.Warningf("gopshost.Info returned error: %v", err)
		s.WriteString(runtime.GOOS)
	}

	vsys, vrole, err := gopshost.Virtualization()
	if err != nil {
		glog.Warningf("gopshost.Virtualization returned error: %v", err)
	} else {
		glog.Infof("virtualization: %s %s", vsys, vrole)
	}

	// This environment is exotic, let's output a bit more.
	if vrole == "guest" || runtime.GOARCH != "amd64" {
		if vsys != "" {
			s.WriteString(fmt.Sprintf(" (%s/%s)", vsys, runtime.GOARCH))
		} else {
			s.WriteString(fmt.Sprintf(" (%s)", runtime.GOARCH))
		}
	}
	return s.String()
}

// runStart handles the executes the flow of "minikube start"
func runStart(cmd *cobra.Command, args []string) {
	displayVersion(version.GetVersion())
	displayEnviron(os.Environ())

	// if --registry-mirror specified when run minikube start,
	// take arg precedence over MINIKUBE_REGISTRY_MIRROR
	// actually this is a hack, because viper 1.0.0 can assign env to variable if StringSliceVar
	// and i can't update it to 1.4.0, it affects too much code
	// other types (like String, Bool) of flag works, so imageRepository, imageMirrorCountry
	// can be configured as MINIKUBE_IMAGE_REPOSITORY and IMAGE_MIRROR_COUNTRY
	// this should be updated to documentation
	if len(registryMirror) == 0 {
		registryMirror = viper.GetStringSlice("registry_mirror")
	}

	existing, err := config.Load(viper.GetString(config.MachineProfile))
	if err != nil && !config.IsNotExist(err) {
		exit.WithCodeT(exit.Data, "Unable to load config: {{.error}}", out.V{"error": err})
	}

	ds := selectDriver(existing)
	driverName := ds.Name
	glog.Infof("selected driver: %s", driverName)
	validateDriver(ds, existing)
	err = autoSetDriverOptions(cmd, driverName)
	if err != nil {
		glog.Errorf("Error autoSetOptions : %v", err)
	}

	validateFlags(cmd, driverName)
	validateUser(driverName)

	// Download & update the driver, even in --download-only mode
	if !viper.GetBool(dryRun) {
		updateDriver(driverName)
	}

	k8sVersion := getKubernetesVersion(existing)
	mc, n, err := generateCfgFromFlags(cmd, k8sVersion, driverName)
	if err != nil {
		exit.WithError("Failed to generate config", err)
	}

	// This is about as far as we can go without overwriting config files
	if viper.GetBool(dryRun) {
		out.T(out.DryRun, `dry-run validation complete!`)
		return
	}

	cacheISO(&mc, driverName)

	if viper.GetBool(nativeSSH) {
		ssh.SetDefaultClient(ssh.Native)
	} else {
		ssh.SetDefaultClient(ssh.External)
	}

	// Now that the ISO is downloaded, pull images in the background while the VM boots.
	var cacheGroup errgroup.Group
	beginCacheRequiredImages(&cacheGroup, mc.KubernetesConfig.ImageRepository, k8sVersion)

	// Abstraction leakage alert: startHost requires the config to be saved, to satistfy pkg/provision/buildroot.
	// Hence, saveConfig must be called before startHost, and again afterwards when we know the IP.
	if err := saveConfig(&mc); err != nil {
		exit.WithError("Failed to save config", err)
	}

	// exits here in case of --download-only option.
	handleDownloadOnly(&cacheGroup, k8sVersion)
	mRunner, preExists, machineAPI, host := startMachine(&mc, &n)
	defer machineAPI.Close()
	// configure the runtime (docker, containerd, crio)
	cr := configureRuntimes(mRunner, driverName, mc.KubernetesConfig)
	showVersionInfo(k8sVersion, cr)
	waitCacheRequiredImages(&cacheGroup)

	// Must be written before bootstrap, otherwise health checks may flake due to stale IP
	kubeconfig, err := setupKubeconfig(host, &mc, &n, mc.Name)
	if err != nil {
		exit.WithError("Failed to setup kubeconfig", err)
	}

	// setup kubeadm (must come after setupKubeconfig)
	bs := setupKubeAdm(machineAPI, mc, n)

	// pull images or restart cluster
	bootstrapCluster(bs, cr, mRunner, mc)
	configureMounts()

	// enable addons, both old and new!
	if viper.GetBool(installAddons) {
		existingAddons := map[string]bool{}
		if existing != nil && existing.Addons != nil {
			existingAddons = existing.Addons
		}
		addons.Start(viper.GetString(config.MachineProfile), existingAddons, addonList)
	}

	if err = cacheAndLoadImagesInConfig(); err != nil {
		out.T(out.FailureType, "Unable to load cached images from config file.")
	}

	// special ops for none , like change minikube directory.
	if driverName == driver.None {
		prepareNone()
	}

	// Skip pre-existing, because we already waited for health
	if viper.GetBool(waitUntilHealthy) && !preExists {
		if err := bs.WaitForCluster(mc, viper.GetDuration(waitTimeout)); err != nil {
			exit.WithError("Wait failed", err)
		}
	}
	if err := showKubectlInfo(kubeconfig, k8sVersion, mc.Name); err != nil {
		glog.Errorf("kubectl info: %v", err)
	}
}

func updateDriver(driverName string) {
	v, err := version.GetSemverVersion()
	if err != nil {
		out.WarningT("Error parsing minikube version: {{.error}}", out.V{"error": err})
	} else if err := driver.InstallOrUpdate(driverName, localpath.MakeMiniPath("bin"), v, viper.GetBool(interactive), viper.GetBool(autoUpdate)); err != nil {
		out.WarningT("Unable to update {{.driver}} driver: {{.error}}", out.V{"driver": driverName, "error": err})
	}
}

func cacheISO(cfg *config.MachineConfig, driverName string) {
	if !driver.BareMetal(driverName) && !driver.IsKIC(driverName) {
		if err := cluster.CacheISO(*cfg); err != nil {
			exit.WithError("Failed to cache ISO", err)
		}
	}
}

func displayVersion(version string) {
	prefix := ""
	if viper.GetString(config.MachineProfile) != constants.DefaultMachineName {
		prefix = fmt.Sprintf("[%s] ", viper.GetString(config.MachineProfile))
	}

	versionState := out.Happy
	if notify.MaybePrintUpdateTextFromGithub() {
		versionState = out.Meh
	}

	out.T(versionState, "{{.prefix}}minikube {{.version}} on {{.platform}}", out.V{"prefix": prefix, "version": version, "platform": platform()})
}

// displayEnviron makes the user aware of environment variables that will affect how minikube operates
func displayEnviron(env []string) {
	for _, kv := range env {
		bits := strings.SplitN(kv, "=", 2)
		k := bits[0]
		v := bits[1]
		if strings.HasPrefix(k, "MINIKUBE_") || k == constants.KubeconfigEnvVar {
			out.T(out.Option, "{{.key}}={{.value}}", out.V{"key": k, "value": v})
		}
	}
}

func setupKubeconfig(h *host.Host, c *config.MachineConfig, n *config.Node, clusterName string) (*kubeconfig.Settings, error) {
	addr, err := h.Driver.GetURL()
	if err != nil {
		exit.WithError("Failed to get driver URL", err)
	}
	if !driver.IsKIC(h.DriverName) {
		addr = strings.Replace(addr, "tcp://", "https://", -1)
		addr = strings.Replace(addr, ":2376", ":"+strconv.Itoa(n.Port), -1)
	}

	if c.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.Replace(addr, n.IP, c.KubernetesConfig.APIServerName, -1)
	}
	kcs := &kubeconfig.Settings{
		ClusterName:          clusterName,
		ClusterServerAddress: addr,
		ClientCertificate:    localpath.MakeMiniPath("client.crt"),
		ClientKey:            localpath.MakeMiniPath("client.key"),
		CertificateAuthority: localpath.MakeMiniPath("ca.crt"),
		KeepContext:          viper.GetBool(keepContext),
		EmbedCerts:           viper.GetBool(embedCerts),
	}

	kcs.SetPath(kubeconfig.PathFromEnv())
	if err := kubeconfig.Update(kcs); err != nil {
		return kcs, err
	}
	return kcs, nil
}

func handleDownloadOnly(cacheGroup *errgroup.Group, k8sVersion string) {
	// If --download-only, complete the remaining downloads and exit.
	if !viper.GetBool(downloadOnly) {
		return
	}
	if err := doCacheBinaries(k8sVersion); err != nil {
		exit.WithError("Failed to cache binaries", err)
	}
	waitCacheRequiredImages(cacheGroup)
	if err := saveImagesToTarFromConfig(); err != nil {
		exit.WithError("Failed to cache images to tar", err)
	}
	out.T(out.Check, "Download complete!")
	os.Exit(0)

}

func startMachine(cfg *config.MachineConfig, node *config.Node) (runner command.Runner, preExists bool, machineAPI libmachine.API, host *host.Host) {
	m, err := machine.NewAPIClient()
	if err != nil {
		exit.WithError("Failed to get machine client", err)
	}
	host, preExists = startHost(m, *cfg)
	runner, err = machine.CommandRunner(host)
	if err != nil {
		exit.WithError("Failed to get command runner", err)
	}

	ip := validateNetwork(host, runner)

	// Bypass proxy for minikube's vm host ip
	err = proxy.ExcludeIP(ip)
	if err != nil {
		out.ErrT(out.FailureType, "Failed to set NO_PROXY Env. Please use `export NO_PROXY=$NO_PROXY,{{.ip}}`.", out.V{"ip": ip})
	}
	// Save IP to configuration file for subsequent use
	node.IP = ip

	if err := saveNodeToConfig(cfg, node); err != nil {
		exit.WithError("Failed to save config", err)
	}

	return runner, preExists, m, host
}

func showVersionInfo(k8sVersion string, cr cruntime.Manager) {
	version, _ := cr.Version()
	out.T(cr.Style(), "Preparing Kubernetes {{.k8sVersion}} on {{.runtime}} {{.runtimeVersion}} ...", out.V{"k8sVersion": k8sVersion, "runtime": cr.Name(), "runtimeVersion": version})
	for _, v := range dockerOpt {
		out.T(out.Option, "opt {{.docker_option}}", out.V{"docker_option": v})
	}
	for _, v := range dockerEnv {
		out.T(out.Option, "env {{.docker_env}}", out.V{"docker_env": v})
	}
}

func showKubectlInfo(kcs *kubeconfig.Settings, k8sVersion string, machineName string) error {
	if kcs.KeepContext {
		out.T(out.Kubectl, "To connect to this cluster, use: kubectl --context={{.name}}", out.V{"name": kcs.ClusterName})
	} else {
		out.T(out.Ready, `Done! kubectl is now configured to use "{{.name}}"`, out.V{"name": machineName})
	}

	path, err := exec.LookPath("kubectl")
	if err != nil {
		out.T(out.Tip, "For best results, install kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl/")
		return nil
	}

	j, err := exec.Command(path, "version", "--client", "--output=json").Output()
	if err != nil {
		return errors.Wrap(err, "exec")
	}

	cv := struct {
		ClientVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"clientVersion"`
	}{}
	err = json.Unmarshal(j, &cv)
	if err != nil {
		return errors.Wrap(err, "unmarshal")
	}

	client, err := semver.Make(strings.TrimPrefix(cv.ClientVersion.GitVersion, version.VersionPrefix))
	if err != nil {
		return errors.Wrap(err, "client semver")
	}

	cluster := semver.MustParse(strings.TrimPrefix(k8sVersion, version.VersionPrefix))
	minorSkew := int(math.Abs(float64(int(client.Minor) - int(cluster.Minor))))
	glog.Infof("kubectl: %s, cluster: %s (minor skew: %d)", client, cluster, minorSkew)

	if client.Major != cluster.Major || minorSkew > 1 {
		out.WarningT("{{.path}} is version {{.client_version}}, and is incompatible with Kubernetes {{.cluster_version}}. You will need to update {{.path}} or use 'minikube kubectl' to connect with this cluster",
			out.V{"path": path, "client_version": client, "cluster_version": cluster})
	}
	return nil
}

func selectDriver(existing *config.MachineConfig) registry.DriverState {
	// Technically unrelated, but important to perform before detection
	driver.SetLibvirtURI(viper.GetString(kvmQemuURI))

	if viper.GetString("vm-driver") != "" {
		ds := driver.Status(viper.GetString("vm-driver"))
		out.T(out.Sparkle, `Using the {{.driver}} driver based on user configuration`, out.V{"driver": ds.String()})
		return ds
	}

	// By default, the driver is whatever we used last time
	if existing != nil && existing.VMDriver != "" {
		ds := driver.Status(existing.VMDriver)
		out.T(out.Sparkle, `Using the {{.driver}} driver based on existing profile`, out.V{"driver": ds.String()})
		return ds
	}

	pick, alts := driver.Suggest(driver.Choices())
	if pick.Name == "" {
		exit.WithCodeT(exit.Config, "Unable to determine a default driver to use. Try specifying --vm-driver, or see https://minikube.sigs.k8s.io/docs/start/")
	}

	if len(alts) > 1 {
		altNames := []string{}
		for _, a := range alts {
			altNames = append(altNames, a.String())
		}
		out.T(out.Sparkle, `Automatically selected the {{.driver}} driver. Other choices: {{.alternates}}`, out.V{"driver": pick.Name, "alternates": strings.Join(altNames, ", ")})
	} else {
		out.T(out.Sparkle, `Automatically selected the {{.driver}} driver`, out.V{"driver": pick.String()})
	}
	return pick
}

// validateDriver validates that the selected driver appears sane, exits if not
func validateDriver(ds registry.DriverState, existing *config.MachineConfig) {
	name := ds.Name
	glog.Infof("validating driver %q against %+v", name, existing)
	if !driver.Supported(name) {
		exit.WithCodeT(exit.Unavailable, "The driver {{.experimental}} '{{.driver}}' is not supported on {{.os}}", out.V{"driver": name, "os": runtime.GOOS})
	}

	st := ds.State
	glog.Infof("status for %s: %+v", name, st)

	if st.Error != nil {
		out.ErrLn("")

		out.WarningT("'{{.driver}}' driver reported an issue: {{.error}}", out.V{"driver": name, "error": st.Error})
		out.ErrT(out.Tip, "Suggestion: {{.fix}}", out.V{"fix": translate.T(st.Fix)})
		if st.Doc != "" {
			out.ErrT(out.Documentation, "Documentation: {{.url}}", out.V{"url": st.Doc})
		}
		out.ErrLn("")

		if !st.Installed && !viper.GetBool(force) {
			if existing != nil && name == existing.VMDriver {
				exit.WithCodeT(exit.Unavailable, "{{.driver}} does not appear to be installed, but is specified by an existing profile. Please run 'minikube delete' or install {{.driver}}", out.V{"driver": name})
			}
			exit.WithCodeT(exit.Unavailable, "{{.driver}} does not appear to be installed", out.V{"driver": name})
		}
	}

	if existing == nil {
		return
	}

	api, err := machine.NewAPIClient()
	if err != nil {
		glog.Warningf("selectDriver NewAPIClient: %v", err)
		return
	}

	machineName := viper.GetString(config.MachineProfile)
	h, err := api.Load(machineName)
	if err != nil {
		glog.Warningf("selectDriver api.Load: %v", err)
		return
	}

	if h.Driver.DriverName() == name {
		return
	}

	out.ErrT(out.Conflict, `The existing "{{.profile_name}}" VM that was created using the "{{.old_driver}}" driver, and is incompatible with the "{{.driver}}" driver.`,
		out.V{"profile_name": machineName, "driver": name, "old_driver": h.Driver.DriverName()})

	out.ErrT(out.Workaround, `To proceed, either:

    1) Delete the existing "{{.profile_name}}" cluster using: '{{.command}} delete'

    * or *

    2) Start the existing "{{.profile_name}}" cluster using: '{{.command}} start --vm-driver={{.old_driver}}'
	`, out.V{"command": minikubeCmd(), "old_driver": h.Driver.DriverName(), "profile_name": machineName})

	exit.WithCodeT(exit.Config, "Exiting.")
}

func selectImageRepository(mirrorCountry string) (bool, string, error) {
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
		pauseImage := images.Pause(repo)
		ref, err := name.ParseReference(pauseImage, name.WeakValidation)
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

// Return a minikube command containing the current profile name
func minikubeCmd() string {
	if viper.GetString(config.MachineProfile) != constants.DefaultMachineName {
		return fmt.Sprintf("minikube -p %s", config.MachineProfile)
	}
	return "minikube"
}

// validateUser validates minikube is run by the recommended user (privileged or regular)
func validateUser(drvName string) {
	u, err := user.Current()
	if err != nil {
		glog.Errorf("Error getting the current user: %v", err)
		return
	}

	useForce := viper.GetBool(force)

	if driver.BareMetal(drvName) && u.Uid != "0" && !useForce {
		exit.WithCodeT(exit.Permissions, `The "{{.driver_name}}" driver requires root privileges. Please run minikube using 'sudo minikube --vm-driver={{.driver_name}}'.`, out.V{"driver_name": drvName})
	}

	if driver.BareMetal(drvName) || u.Uid != "0" {
		return
	}

	out.T(out.Stopped, `The "{{.driver_name}}" driver should not be used with root privileges.`, out.V{"driver_name": drvName})
	out.T(out.Tip, "If you are running minikube within a VM, consider using --vm-driver=none:")
	out.T(out.Documentation, "  https://minikube.sigs.k8s.io/docs/reference/drivers/none/")

	if !useForce {
		os.Exit(exit.Permissions)
	}
	_, err = config.Load(viper.GetString(config.MachineProfile))
	if err == nil || !config.IsNotExist(err) {
		out.T(out.Tip, "Tip: To remove this root owned cluster, run: sudo {{.cmd}} delete", out.V{"cmd": minikubeCmd()})
	}
	if !useForce {
		exit.WithCodeT(exit.Permissions, "Exiting")
	}
}

// validateDiskSize validates the disk size matches the minimum recommended
func validateDiskSize() {
	diskSizeMB := pkgutil.CalculateSizeInMB(viper.GetString(humanReadableDiskSize))
	if diskSizeMB < pkgutil.CalculateSizeInMB(minimumDiskSize) && !viper.GetBool(force) {
		exit.WithCodeT(exit.Config, "Requested disk size {{.requested_size}} is less than minimum of {{.minimum_size}}", out.V{"requested_size": diskSizeMB, "minimum_size": pkgutil.CalculateSizeInMB(minimumDiskSize)})
	}
}

// validateMemorySize validates the memory size matches the minimum recommended
func validateMemorySize() {
	memorySizeMB := pkgutil.CalculateSizeInMB(viper.GetString(memory))
	if memorySizeMB < pkgutil.CalculateSizeInMB(minimumMemorySize) && !viper.GetBool(force) {
		exit.WithCodeT(exit.Config, "Requested memory allocation {{.requested_size}} is less than the minimum allowed of {{.minimum_size}}", out.V{"requested_size": memorySizeMB, "minimum_size": pkgutil.CalculateSizeInMB(minimumMemorySize)})
	}
	if memorySizeMB < pkgutil.CalculateSizeInMB(defaultMemorySize) && !viper.GetBool(force) {
		out.T(out.Notice, "Requested memory allocation ({{.memory}}MB) is less than the default memory allocation of {{.default_memorysize}}MB. Beware that minikube might not work correctly or crash unexpectedly.",
			out.V{"memory": memorySizeMB, "default_memorysize": pkgutil.CalculateSizeInMB(defaultMemorySize)})
	}
}

// validateCPUCount validates the cpu count matches the minimum recommended
func validateCPUCount(local bool) {
	var cpuCount int
	if local {
		// Uses the gopsutil cpu package to count the number of physical cpu cores
		ci, err := cpu.Counts(false)
		if err != nil {
			glog.Warningf("Unable to get CPU info: %v", err)
		} else {
			cpuCount = ci
		}
	} else {
		cpuCount = viper.GetInt(cpus)
	}
	if cpuCount < minimumCPUS && !viper.GetBool(force) {
		exit.UsageT("Requested cpu count {{.requested_cpus}} is less than the minimum allowed of {{.minimum_cpus}}", out.V{"requested_cpus": cpuCount, "minimum_cpus": minimumCPUS})
	}
}

// validateFlags validates the supplied flags against known bad combinations
func validateFlags(cmd *cobra.Command, drvName string) {
	validateDiskSize()
	validateMemorySize()

	if driver.BareMetal(drvName) {
		if viper.GetString(config.MachineProfile) != constants.DefaultMachineName {
			exit.WithCodeT(exit.Config, "The 'none' driver does not support multiple profiles: https://minikube.sigs.k8s.io/docs/reference/drivers/none/")
		}

		if cmd.Flags().Changed(cpus) {
			out.WarningT("The 'none' driver does not respect the --cpus flag")
		}
		if cmd.Flags().Changed(memory) {
			out.WarningT("The 'none' driver does not respect the --memory flag")
		}

		runtime := viper.GetString(containerRuntime)
		if runtime != "docker" {
			out.WarningT("Using the '{{.runtime}}' runtime with the 'none' driver is an untested configuration!", out.V{"runtime": runtime})
		}
	}

	validateCPUCount(driver.BareMetal(drvName))

	// check that kubeadm extra args contain only whitelisted parameters
	for param := range extraOptions.AsMap().Get(bsutil.Kubeadm) {
		if !config.ContainsParam(bsutil.KubeadmExtraArgsWhitelist[bsutil.KubeadmCmdParam], param) &&
			!config.ContainsParam(bsutil.KubeadmExtraArgsWhitelist[bsutil.KubeadmConfigParam], param) {
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
				exit.UsageT("Sorry, the url provided with the --registry-mirror flag is invalid: {{.url}}", out.V{"url": loc})
			}

		}
	}
}

// doCacheBinaries caches Kubernetes binaries in the foreground
func doCacheBinaries(k8sVersion string) error {
	return machine.CacheBinariesForBootstrapper(k8sVersion, viper.GetString(cmdcfg.Bootstrapper))
}

// beginCacheRequiredImages caches images required for kubernetes version in the background
func beginCacheRequiredImages(g *errgroup.Group, imageRepository string, k8sVersion string) {
	if !viper.GetBool(cacheImages) {
		return
	}

	g.Go(func() error {
		return machine.CacheImagesForBootstrapper(imageRepository, k8sVersion, viper.GetString(cmdcfg.Bootstrapper))
	})
}

// waitCacheRequiredImages blocks until the required images are all cached.
func waitCacheRequiredImages(g *errgroup.Group) {
	if !viper.GetBool(cacheImages) {
		return
	}
	if err := g.Wait(); err != nil {
		glog.Errorln("Error caching images: ", err)
	}
}

// generateCfgFromFlags generates config.Config based on flags and supplied arguments
func generateCfgFromFlags(cmd *cobra.Command, k8sVersion string, drvName string) (config.MachineConfig, config.Node, error) {
	r, err := cruntime.New(cruntime.Config{Type: viper.GetString(containerRuntime)})
	if err != nil {
		return config.MachineConfig{}, config.Node{}, err
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
	if _, ok := r.(*cruntime.Docker); ok && !cmd.Flags().Changed("docker-env") {
		setDockerProxy()
	}

	repository := viper.GetString(imageRepository)
	mirrorCountry := strings.ToLower(viper.GetString(imageMirrorCountry))
	if strings.ToLower(repository) == "auto" || mirrorCountry != "" {
		found, autoSelectedRepository, err := selectImageRepository(mirrorCountry)
		if err != nil {
			exit.WithError("Failed to check main repository and mirrors for images for images", err)
		}

		if !found {
			if autoSelectedRepository == "" {
				exit.WithCodeT(exit.Failure, "None of the known repositories is accessible. Consider specifying an alternative image repository with --image-repository flag")
			} else {
				out.WarningT("None of the known repositories in your location are accessible. Using {{.image_repository_name}} as fallback.", out.V{"image_repository_name": autoSelectedRepository})
			}
		}

		repository = autoSelectedRepository
	}

	if cmd.Flags().Changed(imageRepository) {
		out.T(out.SuccessType, "Using image repository {{.name}}", out.V{"name": repository})
	}

	var kubeNodeName string
	if drvName != driver.None {
		kubeNodeName = viper.GetString(config.MachineProfile)
	}

	// Create the initial node, which will necessarily be a control plane
	cp := config.Node{
		Port:              viper.GetInt(apiServerPort),
		KubernetesVersion: k8sVersion,
		Name:              kubeNodeName,
		ControlPlane:      true,
		Worker:            true,
	}

	cfg := config.MachineConfig{
		Name:                    viper.GetString(config.MachineProfile),
		KeepContext:             viper.GetBool(keepContext),
		EmbedCerts:              viper.GetBool(embedCerts),
		MinikubeISO:             viper.GetString(isoURL),
		Memory:                  pkgutil.CalculateSizeInMB(viper.GetString(memory)),
		CPUs:                    viper.GetInt(cpus),
		DiskSize:                pkgutil.CalculateSizeInMB(viper.GetString(humanReadableDiskSize)),
		VMDriver:                drvName,
		HyperkitVpnKitSock:      viper.GetString(vpnkitSock),
		HyperkitVSockPorts:      viper.GetStringSlice(vsockPorts),
		NFSShare:                viper.GetStringSlice(nfsShare),
		NFSSharesRoot:           viper.GetString(nfsSharesRoot),
		DockerEnv:               dockerEnv,
		DockerOpt:               dockerOpt,
		InsecureRegistry:        insecureRegistry,
		RegistryMirror:          registryMirror,
		HostOnlyCIDR:            viper.GetString(hostOnlyCIDR),
		HypervVirtualSwitch:     viper.GetString(hypervVirtualSwitch),
		HypervUseExternalSwitch: viper.GetBool(hypervUseExternalSwitch),
		HypervExternalAdapter:   viper.GetString(hypervExternalAdapter),
		KVMNetwork:              viper.GetString(kvmNetwork),
		KVMQemuURI:              viper.GetString(kvmQemuURI),
		KVMGPU:                  viper.GetBool(kvmGPU),
		KVMHidden:               viper.GetBool(kvmHidden),
		Downloader:              pkgutil.DefaultDownloader{},
		DisableDriverMounts:     viper.GetBool(disableDriverMounts),
		UUID:                    viper.GetString(uuid),
		NoVTXCheck:              viper.GetBool(noVTXCheck),
		DNSProxy:                viper.GetBool(dnsProxy),
		HostDNSResolver:         viper.GetBool(hostDNSResolver),
		HostOnlyNicType:         viper.GetString(hostOnlyNicType),
		NatNicType:              viper.GetString(natNicType),
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion:      k8sVersion,
			ClusterName:            viper.GetString(config.MachineProfile),
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
		Nodes: []config.Node{cp},
	}
	return cfg, cp, nil
}

// setDockerProxy sets the proxy environment variables in the docker environment.
func setDockerProxy() {
	for _, k := range proxy.EnvVars {
		if v := os.Getenv(k); v != "" {
			// convert https_proxy to HTTPS_PROXY for linux
			// TODO (@medyagh): if user has both http_proxy & HTTPS_PROXY set merge them.
			k = strings.ToUpper(k)
			if k == "HTTP_PROXY" || k == "HTTPS_PROXY" {
				if strings.HasPrefix(v, "localhost") || strings.HasPrefix(v, "127.0") {
					out.WarningT("Not passing {{.name}}={{.value}} to docker env.", out.V{"name": k, "value": v})
					continue
				}
			}
			dockerEnv = append(dockerEnv, fmt.Sprintf("%s=%s", k, v))
		}
	}
}

// autoSetDriverOptions sets the options needed for specific vm-driver automatically.
func autoSetDriverOptions(cmd *cobra.Command, drvName string) (err error) {
	err = nil
	hints := driver.FlagDefaults(drvName)
	if !cmd.Flags().Changed("extra-config") && len(hints.ExtraOptions) > 0 {
		for _, eo := range hints.ExtraOptions {
			glog.Infof("auto setting extra-config to %q.", eo)
			err = extraOptions.Set(eo)
			if err != nil {
				err = errors.Wrapf(err, "setting extra option %s", eo)
			}
		}
	}

	if !cmd.Flags().Changed(cacheImages) {
		viper.Set(cacheImages, hints.CacheImages)
	}

	if !cmd.Flags().Changed(containerRuntime) && hints.ContainerRuntime != "" {
		viper.Set(containerRuntime, hints.ContainerRuntime)
		glog.Infof("auto set %s to %q.", containerRuntime, hints.ContainerRuntime)
	}

	if !cmd.Flags().Changed(cmdcfg.Bootstrapper) && hints.Bootstrapper != "" {
		viper.Set(cmdcfg.Bootstrapper, hints.Bootstrapper)
		glog.Infof("auto set %s to %q.", cmdcfg.Bootstrapper, hints.Bootstrapper)

	}

	return err
}

// prepareNone prepares the user and host for the joy of the "none" driver
func prepareNone() {
	out.T(out.StartingNone, "Configuring local host environment ...")
	if viper.GetBool(config.WantNoneDriverWarning) {
		out.T(out.Empty, "")
		out.WarningT("The 'none' driver provides limited isolation and may reduce system security and reliability.")
		out.WarningT("For more information, see:")
		out.T(out.URL, "https://minikube.sigs.k8s.io/docs/reference/drivers/none/")
		out.T(out.Empty, "")
	}

	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") == "" {
		home := os.Getenv("HOME")
		out.WarningT("kubectl and minikube configuration will be stored in {{.home_folder}}", out.V{"home_folder": home})
		out.WarningT("To use kubectl or minikube commands as your own user, you may need to relocate them. For example, to overwrite your own settings, run:")

		out.T(out.Empty, "")
		out.T(out.Command, "sudo mv {{.home_folder}}/.kube {{.home_folder}}/.minikube $HOME", out.V{"home_folder": home})
		out.T(out.Command, "sudo chown -R $USER $HOME/.kube $HOME/.minikube")
		out.T(out.Empty, "")

		out.T(out.Tip, "This can also be done automatically by setting the env var CHANGE_MINIKUBE_NONE_USER=true")
	}

	if err := pkgutil.MaybeChownDirRecursiveToMinikubeUser(localpath.MiniPath()); err != nil {
		exit.WithCodeT(exit.Permissions, "Failed to change permissions for {{.minikube_dir_path}}: {{.error}}", out.V{"minikube_dir_path": localpath.MiniPath(), "error": err})
	}
}

// startHost starts a new minikube host using a VM or None
func startHost(api libmachine.API, mc config.MachineConfig) (*host.Host, bool) {
	exists, err := api.Exists(mc.Name)
	if err != nil {
		exit.WithError("Failed to check if machine exists", err)
	}

	host, err := cluster.StartHost(api, mc)
	if err != nil {
		// If virtual machine does not exist due to user interrupt cancel(i.e. Ctrl + C), initialize exists flag
		if err == cluster.ErrorMachineNotExist {
			// If Machine does not exist, of course the machine does not have kubeadm config files
			// In order not to determine the machine has kubeadm config files, initialize exists flag
			// ※ If exists flag is true, minikube determines the machine has kubeadm config files
			return host, false
		}
		exit.WithError("Unable to start VM. Please investigate and run 'minikube delete' if possible", err)
	}
	return host, exists
}

// validateNetwork tries to catch network problems as soon as possible
func validateNetwork(h *host.Host, r command.Runner) string {
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
				out.WarningT("You appear to be using a proxy, but your NO_PROXY environment does not include the minikube IP ({{.ip_address}}). Please see {{.documentation_url}} for more details", out.V{"ip_address": ip, "documentation_url": "https://minikube.sigs.k8s.io/docs/reference/networking/proxy/"})
				warnedOnce = true
			}
		}
	}

	if !driver.BareMetal(h.Driver.DriverName()) && !driver.IsKIC(h.Driver.DriverName()) {
		trySSH(h, ip)
	}

	tryLookup(r)
	tryRegistry(r)
	return ip
}

func trySSH(h *host.Host, ip string) {
	if viper.GetBool(force) {
		return
	}

	sshAddr := net.JoinHostPort(ip, "22")

	dial := func() (err error) {
		d := net.Dialer{Timeout: 3 * time.Second}
		conn, err := d.Dial("tcp", sshAddr)
		if err != nil {
			out.WarningT("Unable to verify SSH connectivity: {{.error}}. Will retry...", out.V{"error": err})
			return err
		}
		_ = conn.Close()
		return nil
	}

	if err := retry.Expo(dial, time.Second, 13*time.Second); err != nil {
		exit.WithCodeT(exit.IO, `minikube is unable to connect to the VM: {{.error}}

	This is likely due to one of two reasons:

	- VPN or firewall interference
	- {{.hypervisor}} network configuration issue

	Suggested workarounds:

	- Disable your local VPN or firewall software
	- Configure your local VPN or firewall to allow access to {{.ip}}
	- Restart or reinstall {{.hypervisor}}
	- Use an alternative --vm-driver
	- Use --force to override this connectivity check
	`, out.V{"error": err, "hypervisor": h.Driver.DriverName(), "ip": ip})
	}
}

func tryLookup(r command.Runner) {
	// DNS check
	if rr, err := r.RunCmd(exec.Command("nslookup", "kubernetes.io", "-type=ns")); err != nil {
		glog.Infof("%s failed: %v which might be okay will retry nslookup without query type", rr.Args, err)
		// will try with without query type for ISOs with different busybox versions.
		if _, err = r.RunCmd(exec.Command("nslookup", "kubernetes.io")); err != nil {
			glog.Warningf("nslookup failed: %v", err)
			out.WarningT("Node may be unable to resolve external DNS records")
		}
	}
}
func tryRegistry(r command.Runner) {
	// Try an HTTPS connection to the image repository
	proxy := os.Getenv("HTTPS_PROXY")
	opts := []string{"-sS"}
	if proxy != "" && !strings.HasPrefix(proxy, "localhost") && !strings.HasPrefix(proxy, "127.0") {
		opts = append([]string{"-x", proxy}, opts...)
	}

	repo := viper.GetString(imageRepository)
	if repo == "" {
		repo = images.DefaultKubernetesRepo
	}

	opts = append(opts, fmt.Sprintf("https://%s/", repo))
	if rr, err := r.RunCmd(exec.Command("curl", opts...)); err != nil {
		glog.Warningf("%s failed: %v", rr.Args, err)
		out.WarningT("VM is unable to access {{.repository}}, you may need to configure a proxy or set --image-repository", out.V{"repository": repo})
	}
}

// getKubernetesVersion ensures that the requested version is reasonable
func getKubernetesVersion(old *config.MachineConfig) string {
	paramVersion := viper.GetString(kubernetesVersion)

	if paramVersion == "" { // if the user did not specify any version then ...
		if old != nil { // .. use the old version from config (if any)
			paramVersion = old.KubernetesConfig.KubernetesVersion
		}
		if paramVersion == "" { // .. otherwise use the default version
			paramVersion = constants.DefaultKubernetesVersion
		}
	}

	nvs, err := semver.Make(strings.TrimPrefix(paramVersion, version.VersionPrefix))
	if err != nil {
		exit.WithCodeT(exit.Data, `Unable to parse "{{.kubernetes_version}}": {{.error}}`, out.V{"kubernetes_version": paramVersion, "error": err})
	}
	nv := version.VersionPrefix + nvs.String()

	if old == nil || old.KubernetesConfig.KubernetesVersion == "" {
		return nv
	}

	oldestVersion, err := semver.Make(strings.TrimPrefix(constants.OldestKubernetesVersion, version.VersionPrefix))
	if err != nil {
		exit.WithCodeT(exit.Data, "Unable to parse oldest Kubernetes version from constants: {{.error}}", out.V{"error": err})
	}
	defaultVersion, err := semver.Make(strings.TrimPrefix(constants.DefaultKubernetesVersion, version.VersionPrefix))
	if err != nil {
		exit.WithCodeT(exit.Data, "Unable to parse default Kubernetes version from constants: {{.error}}", out.V{"error": err})
	}

	if nvs.LT(oldestVersion) {
		out.WarningT("Specified Kubernetes version {{.specified}} is less than the oldest supported version: {{.oldest}}", out.V{"specified": nvs, "oldest": constants.OldestKubernetesVersion})
		if viper.GetBool(force) {
			out.WarningT("Kubernetes {{.version}} is not supported by this release of minikube", out.V{"version": nvs})
		} else {
			exit.WithCodeT(exit.Data, "Sorry, Kubernetes {{.version}} is not supported by this release of minikube", out.V{"version": nvs})
		}
	}

	ovs, err := semver.Make(strings.TrimPrefix(old.KubernetesConfig.KubernetesVersion, version.VersionPrefix))
	if err != nil {
		glog.Errorf("Error parsing old version %q: %v", old.KubernetesConfig.KubernetesVersion, err)
	}

	if nvs.LT(ovs) {
		nv = version.VersionPrefix + ovs.String()
		profileArg := ""
		if old.Name != constants.DefaultMachineName {
			profileArg = fmt.Sprintf("-p %s", old.Name)
		}
		exit.WithCodeT(exit.Config, `Error: You have selected Kubernetes v{{.new}}, but the existing cluster for your profile is running Kubernetes v{{.old}}. Non-destructive downgrades are not supported, but you can proceed by performing one of the following options:

* Recreate the cluster using Kubernetes v{{.new}}: Run "minikube delete {{.profile}}", then "minikube start {{.profile}} --kubernetes-version={{.new}}"
* Create a second cluster with Kubernetes v{{.new}}: Run "minikube start -p <new name> --kubernetes-version={{.new}}"
* Reuse the existing cluster with Kubernetes v{{.old}} or newer: Run "minikube start {{.profile}} --kubernetes-version={{.old}}"`, out.V{"new": nvs, "old": ovs, "profile": profileArg})

	}
	if defaultVersion.GT(nvs) {
		out.T(out.ThumbsUp, "Kubernetes {{.new}} is now available. If you would like to upgrade, specify: --kubernetes-version={{.new}}", out.V{"new": defaultVersion})
	}
	return nv
}

// setupKubeAdm adds any requested files into the VM before Kubernetes is started
func setupKubeAdm(mAPI libmachine.API, cfg config.MachineConfig, node config.Node) bootstrapper.Bootstrapper {
	bs, err := getClusterBootstrapper(mAPI, viper.GetString(cmdcfg.Bootstrapper))
	if err != nil {
		exit.WithError("Failed to get bootstrapper", err)
	}
	for _, eo := range extraOptions {
		out.T(out.Option, "{{.extra_option_component_name}}.{{.key}}={{.value}}", out.V{"extra_option_component_name": eo.Component, "key": eo.Key, "value": eo.Value})
	}
	// Loads cached images, generates config files, download binaries
	if err := bs.UpdateCluster(cfg); err != nil {
		exit.WithError("Failed to update cluster", err)
	}
	if err := bs.SetupCerts(cfg.KubernetesConfig, node); err != nil {
		exit.WithError("Failed to setup certs", err)
	}
	return bs
}

// configureRuntimes does what needs to happen to get a runtime going.
func configureRuntimes(runner cruntime.CommandRunner, drvName string, k8s config.KubernetesConfig) cruntime.Manager {
	config := cruntime.Config{Type: viper.GetString(containerRuntime), Runner: runner, ImageRepository: k8s.ImageRepository, KubernetesVersion: k8s.KubernetesVersion}
	cr, err := cruntime.New(config)
	if err != nil {
		exit.WithError("Failed runtime", err)
	}

	disableOthers := true
	if driver.BareMetal(drvName) {
		disableOthers = false
	}
	err = cr.Enable(disableOthers)
	if err != nil {
		exit.WithError("Failed to enable container runtime", err)
	}

	return cr
}

// bootstrapCluster starts Kubernetes using the chosen bootstrapper
func bootstrapCluster(bs bootstrapper.Bootstrapper, r cruntime.Manager, runner command.Runner, mc config.MachineConfig) {
	out.T(out.Launch, "Launching Kubernetes ... ")
	if err := bs.StartCluster(mc); err != nil {
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
	if err := lock.WriteFile(filepath.Join(localpath.MiniPath(), constants.MountProcessFileName), []byte(strconv.Itoa(mountCmd.Process.Pid)), 0644); err != nil {
		exit.WithError("Error writing mount pid", err)
	}
}

// saveConfig saves profile cluster configuration in $MINIKUBE_HOME/profiles/<profilename>/config.json
func saveConfig(clusterCfg *config.MachineConfig) error {
	return config.SaveProfile(viper.GetString(config.MachineProfile), clusterCfg)
}

func saveNodeToConfig(cfg *config.MachineConfig, node *config.Node) error {
	for i, n := range cfg.Nodes {
		if n.Name == node.Name {
			cfg.Nodes[i] = *node
			break
		}
	}
	return saveConfig(cfg)
}
