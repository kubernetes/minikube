/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/cni"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	pkgutil "k8s.io/minikube/pkg/util"
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
	cniFlag                 = "cni"
	hypervVirtualSwitch     = "hyperv-virtual-switch"
	hypervUseExternalSwitch = "hyperv-use-external-switch"
	hypervExternalAdapter   = "hyperv-external-adapter"
	kvmNetwork              = "kvm-network"
	kvmQemuURI              = "kvm-qemu-uri"
	kvmGPU                  = "kvm-gpu"
	kvmHidden               = "kvm-hidden"
	minikubeEnvPrefix       = "MINIKUBE"
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
	waitComponents          = "wait"
	force                   = "force"
	dryRun                  = "dry-run"
	interactive             = "interactive"
	waitTimeout             = "wait-timeout"
	nativeSSH               = "native-ssh"
	minUsableMem            = 953  // 1GB In MiB: Kubernetes will not start with less
	minRecommendedMem       = 1907 // 2GB In MiB: Warn at no lower than existing configurations
	minimumCPUS             = 2
	minimumDiskSize         = 2000
	autoUpdate              = "auto-update-drivers"
	hostOnlyNicType         = "host-only-nic-type"
	natNicType              = "nat-nic-type"
	nodes                   = "nodes"
	preload                 = "preload"
	deleteOnFailure         = "delete-on-failure"
	forceSystemd            = "force-systemd"
	kicBaseImage            = "base-image"
	ports                   = "ports"
	startNamespace          = "namespace"
	trace                   = "trace"
)

var (
	outputFormat string
)

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

	startCmd.Flags().Int(cpus, 2, "Number of CPUs allocated to Kubernetes.")
	startCmd.Flags().String(memory, "", "Amount of RAM to allocate to Kubernetes (format: <number>[<unit>], where unit = b, k, m or g).")
	startCmd.Flags().String(humanReadableDiskSize, defaultDiskSize, "Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g).")
	startCmd.Flags().Bool(downloadOnly, false, "If true, only download and cache files for later use - don't install or start anything.")
	startCmd.Flags().Bool(cacheImages, true, "If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --driver=none.")
	startCmd.Flags().StringSlice(isoURL, download.DefaultISOURLs(), "Locations to fetch the minikube ISO from.")
	startCmd.Flags().String(kicBaseImage, kic.BaseImage, "The base image to use for docker/podman drivers. Intended for local development.")
	startCmd.Flags().Bool(keepContext, false, "This will keep the existing kubectl context and will create a minikube context.")
	startCmd.Flags().Bool(embedCerts, false, "if true, will embed the certs in kubeconfig.")
	startCmd.Flags().String(containerRuntime, "docker", fmt.Sprintf("The container runtime to be used (%s).", strings.Join(cruntime.ValidRuntimes(), ", ")))
	startCmd.Flags().Bool(createMount, false, "This will start the mount daemon and automatically mount files into minikube.")
	startCmd.Flags().String(mountString, constants.DefaultMountDir+":/minikube-host", "The argument to pass the minikube mount command on start.")
	startCmd.Flags().StringArrayVar(&config.AddonList, "addons", nil, "Enable addons. see `minikube addons list` for a list of valid addon names.")
	startCmd.Flags().String(criSocket, "", "The cri socket path to be used.")
	startCmd.Flags().String(networkPlugin, "", "Kubelet network plug-in to use (default: auto)")
	startCmd.Flags().Bool(enableDefaultCNI, false, "DEPRECATED: Replaced by --cni=bridge")
	startCmd.Flags().String(cniFlag, "", "CNI plug-in to use. Valid options: auto, bridge, calico, cilium, flannel, kindnet, or path to a CNI manifest (default: auto)")
	startCmd.Flags().StringSlice(waitComponents, kverify.DefaultWaitList, fmt.Sprintf("comma separated list of Kubernetes components to verify and wait for after starting a cluster. defaults to %q, available options: %q . other acceptable values are 'all' or 'none', 'true' and 'false'", strings.Join(kverify.DefaultWaitList, ","), strings.Join(kverify.AllComponentsList, ",")))
	startCmd.Flags().Duration(waitTimeout, 6*time.Minute, "max time to wait per Kubernetes or host to be healthy.")
	startCmd.Flags().Bool(nativeSSH, true, "Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'.")
	startCmd.Flags().Bool(autoUpdate, true, "If set, automatically updates drivers to the latest version. Defaults to true.")
	startCmd.Flags().Bool(installAddons, true, "If set, install addons. Defaults to true.")
	startCmd.Flags().IntP(nodes, "n", 1, "The number of nodes to spin up. Defaults to 1.")
	startCmd.Flags().Bool(preload, true, "If set, download tarball of preloaded images if available to improve start time. Defaults to true.")
	startCmd.Flags().Bool(deleteOnFailure, false, "If set, delete the current cluster if start fails and try again. Defaults to false.")
	startCmd.Flags().Bool(forceSystemd, false, "If set, force the container runtime to use sytemd as cgroup manager. Currently available for docker and crio. Defaults to false.")
	startCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Format to print stdout in. Options include: [text,json]")
	startCmd.Flags().StringP(trace, "", "", "Send trace events. Options include: [gcp]")
}

// initKubernetesFlags inits the commandline flags for Kubernetes related options
func initKubernetesFlags() {
	startCmd.Flags().String(kubernetesVersion, "", fmt.Sprintf("The Kubernetes version that the minikube VM will use (ex: v1.2.3, 'stable' for %s, 'latest' for %s). Defaults to 'stable'.", constants.DefaultKubernetesVersion, constants.NewestKubernetesVersion))
	startCmd.Flags().String(startNamespace, "default", "The named space to activate after start")
	startCmd.Flags().Var(&config.ExtraOptions, "extra-config",
		`A set of key=value pairs that describe configuration that may be passed to different components.
		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
		Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
		Valid kubeadm parameters: `+fmt.Sprintf("%s, %s", strings.Join(bsutil.KubeadmExtraArgsAllowed[bsutil.KubeadmCmdParam], ", "), strings.Join(bsutil.KubeadmExtraArgsAllowed[bsutil.KubeadmConfigParam], ",")))
	startCmd.Flags().String(featureGates, "", "A set of key=value pairs that describe feature gates for alpha/experimental features.")
	startCmd.Flags().String(dnsDomain, constants.ClusterDNSDomain, "The cluster dns domain name used in the Kubernetes cluster")
	startCmd.Flags().Int(apiServerPort, constants.APIServerPort, "The apiserver listening port")
	startCmd.Flags().String(apiServerName, constants.APIServerName, "The authoritative apiserver hostname for apiserver certificates and connectivity. This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().StringSliceVar(&apiServerNames, "apiserver-names", nil, "A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
	startCmd.Flags().IPSliceVar(&apiServerIPs, "apiserver-ips", nil, "A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine")
}

// initDriverFlags inits the commandline flags for vm drivers
func initDriverFlags() {
	startCmd.Flags().String("driver", "", fmt.Sprintf("Driver is one of: %v (defaults to auto-detect)", driver.DisplaySupportedDrivers()))
	startCmd.Flags().String("vm-driver", "", "DEPRECATED, use `driver` instead.")
	startCmd.Flags().Bool(disableDriverMounts, false, "Disables the filesystem mounts provided by the hypervisors")
	startCmd.Flags().Bool("vm", false, "Filter to use only VM Drivers")

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
	startCmd.Flags().String(natNicType, "virtio", "NIC Type used for nat network. One of Am79C970A, Am79C973, 82540EM, 82543GC, 82545EM, or virtio (virtualbox driver only)")

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

	// docker & podman
	startCmd.Flags().StringSlice(ports, []string{}, "List of ports that should be exposed (docker and podman driver only)")
}

// initNetworkingFlags inits the commandline flags for connectivity related flags for start
func initNetworkingFlags() {
	startCmd.Flags().StringSliceVar(&insecureRegistry, "insecure-registry", nil, "Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.")
	startCmd.Flags().StringSliceVar(&registryMirror, "registry-mirror", nil, "Registry mirrors to pass to the Docker daemon")
	startCmd.Flags().String(imageRepository, "", "Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to \"auto\" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers")
	startCmd.Flags().String(imageMirrorCountry, "", "Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn.")
	startCmd.Flags().String(serviceCIDR, constants.DefaultServiceCIDR, "The CIDR to be used for service cluster IPs.")
	startCmd.Flags().StringArrayVar(&config.DockerEnv, "docker-env", nil, "Environment variables to pass to the Docker daemon. (format: key=value)")
	startCmd.Flags().StringArrayVar(&config.DockerOpt, "docker-opt", nil, "Specify arbitrary flags to pass to the Docker daemon. (format: key=value)")
}

// ClusterFlagValue returns the current cluster name based on flags
func ClusterFlagValue() string {
	return viper.GetString(config.ProfileName)
}

// generateClusterConfig generate a config.ClusterConfig based on flags or existing cluster config
func generateClusterConfig(cmd *cobra.Command, existing *config.ClusterConfig, k8sVersion string, drvName string) (config.ClusterConfig, config.Node, error) {
	var cc config.ClusterConfig
	if existing != nil {
		cc = updateExistingConfigFromFlags(cmd, existing)
	} else {
		klog.Info("no existing cluster config was found, will generate one from the flags ")
		sysLimit, containerLimit, err := memoryLimits(drvName)
		if err != nil {
			klog.Warningf("Unable to query memory limits: %+v", err)
		}

		mem := suggestMemoryAllocation(sysLimit, containerLimit, viper.GetInt(nodes))
		if cmd.Flags().Changed(memory) {
			var err error
			mem, err = pkgutil.CalculateSizeInMB(viper.GetString(memory))
			if err != nil {
				exit.Message(reason.Usage, "Generate unable to parse memory '{{.memory}}': {{.error}}", out.V{"memory": viper.GetString(memory), "error": err})
			}
			if driver.IsKIC(drvName) && mem > containerLimit {
				exit.Message(reason.Usage, "{{.driver_name}} has only {{.container_limit}}MB memory but you specified {{.specified_memory}}MB", out.V{"container_limit": containerLimit, "specified_memory": mem, "driver_name": driver.FullName(drvName)})
			}
		} else {
			validateRequestedMemorySize(mem, drvName)
			klog.Infof("Using suggested %dMB memory alloc based on sys=%dMB, container=%dMB", mem, sysLimit, containerLimit)
		}

		diskSize, err := pkgutil.CalculateSizeInMB(viper.GetString(humanReadableDiskSize))
		if err != nil {
			exit.Message(reason.Usage, "Generate unable to parse disk size '{{.diskSize}}': {{.error}}", out.V{"diskSize": viper.GetString(humanReadableDiskSize), "error": err})
		}

		repository := viper.GetString(imageRepository)
		mirrorCountry := strings.ToLower(viper.GetString(imageMirrorCountry))
		if strings.ToLower(repository) == "auto" || (mirrorCountry != "" && repository == "") {
			found, autoSelectedRepository, err := selectImageRepository(mirrorCountry, semver.MustParse(strings.TrimPrefix(k8sVersion, version.VersionPrefix)))
			if err != nil {
				exit.Error(reason.InetRepo, "Failed to check main repository and mirrors for images", err)
			}

			if !found {
				if autoSelectedRepository == "" {
					exit.Message(reason.InetReposUnavailable, "None of the known repositories are accessible. Consider specifying an alternative image repository with --image-repository flag")
				} else {
					out.WarningT("None of the known repositories in your location are accessible. Using {{.image_repository_name}} as fallback.", out.V{"image_repository_name": autoSelectedRepository})
				}
			}

			repository = autoSelectedRepository
		}

		if cmd.Flags().Changed(imageRepository) || cmd.Flags().Changed(imageMirrorCountry) {
			out.Step(style.Success, "Using image repository {{.name}}", false, out.V{"name": repository})
		}

		// Backwards compatibility with --enable-default-cni
		chosenCNI := viper.GetString(cniFlag)
		if viper.GetBool(enableDefaultCNI) && !cmd.Flags().Changed(cniFlag) {
			klog.Errorf("Found deprecated --enable-default-cni flag, setting --cni=bridge")
			chosenCNI = "bridge"
		}
		// networkPlugin cni deprecation warning
		chosenNetworkPlugin := viper.GetString(networkPlugin)
		if chosenNetworkPlugin == "cni" {
			out.WarningT("With --network-plugin=cni, you will need to provide your own CNI. See --cni flag as a user-friendly alternative")
		}

		cc = config.ClusterConfig{
			Name:                    ClusterFlagValue(),
			KeepContext:             viper.GetBool(keepContext),
			EmbedCerts:              viper.GetBool(embedCerts),
			MinikubeISO:             viper.GetString(isoURL),
			KicBaseImage:            viper.GetString(kicBaseImage),
			Memory:                  mem,
			CPUs:                    viper.GetInt(cpus),
			DiskSize:                diskSize,
			Driver:                  drvName,
			HyperkitVpnKitSock:      viper.GetString(vpnkitSock),
			HyperkitVSockPorts:      viper.GetStringSlice(vsockPorts),
			NFSShare:                viper.GetStringSlice(nfsShare),
			NFSSharesRoot:           viper.GetString(nfsSharesRoot),
			DockerEnv:               config.DockerEnv,
			DockerOpt:               config.DockerOpt,
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
			DisableDriverMounts:     viper.GetBool(disableDriverMounts),
			UUID:                    viper.GetString(uuid),
			NoVTXCheck:              viper.GetBool(noVTXCheck),
			DNSProxy:                viper.GetBool(dnsProxy),
			HostDNSResolver:         viper.GetBool(hostDNSResolver),
			HostOnlyNicType:         viper.GetString(hostOnlyNicType),
			NatNicType:              viper.GetString(natNicType),
			StartHostTimeout:        viper.GetDuration(waitTimeout),
			ExposedPorts:            viper.GetStringSlice(ports),
			KubernetesConfig: config.KubernetesConfig{
				KubernetesVersion:      k8sVersion,
				ClusterName:            ClusterFlagValue(),
				Namespace:              viper.GetString(startNamespace),
				APIServerName:          viper.GetString(apiServerName),
				APIServerNames:         apiServerNames,
				APIServerIPs:           apiServerIPs,
				DNSDomain:              viper.GetString(dnsDomain),
				FeatureGates:           viper.GetString(featureGates),
				ContainerRuntime:       viper.GetString(containerRuntime),
				CRISocket:              viper.GetString(criSocket),
				NetworkPlugin:          chosenNetworkPlugin,
				ServiceCIDR:            viper.GetString(serviceCIDR),
				ImageRepository:        repository,
				ExtraOptions:           config.ExtraOptions,
				ShouldLoadCachedImages: viper.GetBool(cacheImages),
				CNI:                    chosenCNI,
				NodePort:               viper.GetInt(apiServerPort),
			},
		}
		cc.VerifyComponents = interpretWaitFlag(*cmd)
		if viper.GetBool(createMount) && driver.IsKIC(drvName) {
			cc.ContainerVolumeMounts = []string{viper.GetString(mountString)}
		}
		cnm, err := cni.New(cc)
		if err != nil {
			return cc, config.Node{}, errors.Wrap(err, "cni")
		}

		if _, ok := cnm.(cni.Disabled); !ok {
			klog.Infof("Found %q CNI - setting NetworkPlugin=cni", cnm)
			cc.KubernetesConfig.NetworkPlugin = "cni"
		}
	}

	klog.Infof("config:\n%+v", cc)

	r, err := cruntime.New(cruntime.Config{Type: cc.KubernetesConfig.ContainerRuntime})
	if err != nil {
		return cc, config.Node{}, errors.Wrap(err, "new runtime manager")
	}

	// Feed Docker our host proxy environment by default, so that it can pull images
	// doing this for both new config and existing, in case proxy changed since previous start
	if _, ok := r.(*cruntime.Docker); ok {
		proxy.SetDockerEnv()
	}

	var kubeNodeName string
	if driver.BareMetal(cc.Driver) {
		kubeNodeName = "m01"
	}
	return createNode(cc, kubeNodeName, existing)
}

// upgradeExistingConfig upgrades legacy configuration files
func upgradeExistingConfig(cc *config.ClusterConfig) {
	if cc == nil {
		return
	}

	if cc.VMDriver != "" && cc.Driver == "" {
		klog.Infof("config upgrade: Driver=%s", cc.VMDriver)
		cc.Driver = cc.VMDriver
	}

	if cc.Name == "" {
		klog.Infof("config upgrade: Name=%s", ClusterFlagValue())
		cc.Name = ClusterFlagValue()
	}

	if cc.KicBaseImage == "" {
		// defaults to kic.BaseImage
		cc.KicBaseImage = viper.GetString(kicBaseImage)
		klog.Infof("config upgrade: KicBaseImage=%s", cc.KicBaseImage)
	}
}

// updateExistingConfigFromFlags will update the existing config from the flags - used on a second start
// skipping updating existing docker env , docker opt, InsecureRegistry, registryMirror, extra-config, apiserver-ips
func updateExistingConfigFromFlags(cmd *cobra.Command, existing *config.ClusterConfig) config.ClusterConfig { //nolint to suppress cyclomatic complexity 45 of func `updateExistingConfigFromFlags` is high (> 30)

	validateFlags(cmd, existing.Driver)

	cc := *existing

	if cmd.Flags().Changed(containerRuntime) {
		cc.KubernetesConfig.ContainerRuntime = viper.GetString(containerRuntime)
	}

	if cmd.Flags().Changed(keepContext) {
		cc.KeepContext = viper.GetBool(keepContext)
	}

	if cmd.Flags().Changed(embedCerts) {
		cc.EmbedCerts = viper.GetBool(embedCerts)
	}

	if cmd.Flags().Changed(isoURL) {
		cc.MinikubeISO = viper.GetString(isoURL)
	}

	if cc.Memory == 0 {
		klog.Info("Existing config file was missing memory. (could be an old minikube config), will use the default value")
		memInMB, err := pkgutil.CalculateSizeInMB(viper.GetString(memory))
		if err != nil {
			klog.Warningf("error calculate memory size in mb : %v", err)
		}
		cc.Memory = memInMB
	}

	if cmd.Flags().Changed(memory) {
		memInMB, err := pkgutil.CalculateSizeInMB(viper.GetString(memory))
		if err != nil {
			klog.Warningf("error calculate memory size in mb : %v", err)
		}
		if memInMB != cc.Memory {
			out.WarningT("You cannot change the memory size for an exiting minikube cluster. Please first delete the cluster.")
		}
	}

	// validate the memory size in case user changed their system memory limits (example change docker desktop or upgraded memory.)
	validateRequestedMemorySize(cc.Memory, cc.Driver)

	if cc.CPUs == 0 {
		klog.Info("Existing config file was missing cpu. (could be an old minikube config), will use the default value")
		cc.CPUs = viper.GetInt(cpus)
	}
	if cmd.Flags().Changed(cpus) {
		if viper.GetInt(cpus) != cc.CPUs {
			out.WarningT("You cannot change the CPUs for an existing minikube cluster. Please first delete the cluster.")
		}
	}

	if cmd.Flags().Changed(humanReadableDiskSize) {
		memInMB, err := pkgutil.CalculateSizeInMB(viper.GetString(humanReadableDiskSize))
		if err != nil {
			klog.Warningf("error calculate disk size in mb : %v", err)
		}

		if memInMB != existing.DiskSize {
			out.WarningT("You cannot change the Disk size for an exiting minikube cluster. Please first delete the cluster.")
		}
	}

	if cmd.Flags().Changed(vpnkitSock) {
		cc.HyperkitVpnKitSock = viper.GetString(vpnkitSock)
	}

	if cmd.Flags().Changed(vsockPorts) {
		cc.HyperkitVSockPorts = viper.GetStringSlice(vsockPorts)
	}

	if cmd.Flags().Changed(nfsShare) {
		cc.NFSShare = viper.GetStringSlice(nfsShare)
	}

	if cmd.Flags().Changed(nfsSharesRoot) {
		cc.NFSSharesRoot = viper.GetString(nfsSharesRoot)
	}

	if cmd.Flags().Changed(hostOnlyCIDR) {
		cc.HostOnlyCIDR = viper.GetString(hostOnlyCIDR)
	}

	if cmd.Flags().Changed(hypervVirtualSwitch) {
		cc.HypervVirtualSwitch = viper.GetString(hypervVirtualSwitch)
	}

	if cmd.Flags().Changed(hypervUseExternalSwitch) {
		cc.HypervUseExternalSwitch = viper.GetBool(hypervUseExternalSwitch)
	}

	if cmd.Flags().Changed(hypervExternalAdapter) {
		cc.HypervExternalAdapter = viper.GetString(hypervExternalAdapter)
	}

	if cmd.Flags().Changed(kvmNetwork) {
		cc.KVMNetwork = viper.GetString(kvmNetwork)
	}

	if cmd.Flags().Changed(kvmQemuURI) {
		cc.KVMQemuURI = viper.GetString(kvmQemuURI)
	}

	if cmd.Flags().Changed(kvmGPU) {
		cc.KVMGPU = viper.GetBool(kvmGPU)
	}

	if cmd.Flags().Changed(kvmHidden) {
		cc.KVMHidden = viper.GetBool(kvmHidden)
	}

	if cmd.Flags().Changed(disableDriverMounts) {
		cc.DisableDriverMounts = viper.GetBool(disableDriverMounts)
	}

	if cmd.Flags().Changed(uuid) {
		cc.UUID = viper.GetString(uuid)
	}

	if cmd.Flags().Changed(noVTXCheck) {
		cc.NoVTXCheck = viper.GetBool(noVTXCheck)
	}

	if cmd.Flags().Changed(dnsProxy) {
		cc.DNSProxy = viper.GetBool(dnsProxy)
	}

	if cmd.Flags().Changed(hostDNSResolver) {
		cc.HostDNSResolver = viper.GetBool(hostDNSResolver)
	}

	if cmd.Flags().Changed(hostOnlyNicType) {
		cc.HostOnlyNicType = viper.GetString(hostOnlyNicType)
	}

	if cmd.Flags().Changed(natNicType) {
		cc.NatNicType = viper.GetString(natNicType)
	}

	if cmd.Flags().Changed(kubernetesVersion) {
		cc.KubernetesConfig.KubernetesVersion = getKubernetesVersion(existing)
	}

	if cmd.Flags().Changed(startNamespace) {
		cc.KubernetesConfig.Namespace = viper.GetString(startNamespace)
	}

	if cmd.Flags().Changed(apiServerName) {
		cc.KubernetesConfig.APIServerName = viper.GetString(apiServerName)
	}

	if cmd.Flags().Changed("apiserver-names") {
		cc.KubernetesConfig.APIServerNames = viper.GetStringSlice("apiserver-names")
	}

	if cmd.Flags().Changed(apiServerPort) {
		cc.KubernetesConfig.NodePort = viper.GetInt(apiServerPort)
	}

	if cmd.Flags().Changed(vsockPorts) {
		cc.ExposedPorts = viper.GetStringSlice(ports)
	}

	// pre minikube 1.9.2 cc.KubernetesConfig.NodePort was not populated.
	// in minikube config there were two fields for api server port.
	// one in cc.KubernetesConfig.NodePort and one in cc.Nodes.Port
	// this makes sure api server port not be set as 0!
	if existing.KubernetesConfig.NodePort == 0 {
		cc.KubernetesConfig.NodePort = viper.GetInt(apiServerPort)
	}

	if cmd.Flags().Changed(dnsDomain) {
		cc.KubernetesConfig.DNSDomain = viper.GetString(dnsDomain)
	}

	if cmd.Flags().Changed(featureGates) {
		cc.KubernetesConfig.FeatureGates = viper.GetString(featureGates)
	}

	if cmd.Flags().Changed(containerRuntime) {
		cc.KubernetesConfig.ContainerRuntime = viper.GetString(containerRuntime)
	}

	if cmd.Flags().Changed(criSocket) {
		cc.KubernetesConfig.CRISocket = viper.GetString(criSocket)
	}

	if cmd.Flags().Changed(networkPlugin) {
		cc.KubernetesConfig.NetworkPlugin = viper.GetString(networkPlugin)
	}

	if cmd.Flags().Changed(serviceCIDR) {
		cc.KubernetesConfig.ServiceCIDR = viper.GetString(serviceCIDR)
	}

	if cmd.Flags().Changed(cacheImages) {
		cc.KubernetesConfig.ShouldLoadCachedImages = viper.GetBool(cacheImages)
	}

	if cmd.Flags().Changed(imageRepository) {
		cc.KubernetesConfig.ImageRepository = viper.GetString(imageRepository)
	}

	if cmd.Flags().Changed(enableDefaultCNI) && !cmd.Flags().Changed(cniFlag) {
		if viper.GetBool(enableDefaultCNI) {
			klog.Errorf("Found deprecated --enable-default-cni flag, setting --cni=bridge")
			cc.KubernetesConfig.CNI = "bridge"
		}
	}

	if cmd.Flags().Changed(cniFlag) {
		cc.KubernetesConfig.CNI = viper.GetString(cniFlag)
	}

	if cmd.Flags().Changed(waitComponents) {
		cc.VerifyComponents = interpretWaitFlag(*cmd)
	}

	// Handle flags and legacy configuration upgrades that do not contain KicBaseImage
	if cmd.Flags().Changed(kicBaseImage) || cc.KicBaseImage == "" {
		cc.KicBaseImage = viper.GetString(kicBaseImage)
	}

	return cc
}

// interpretWaitFlag interprets the wait flag and respects the legacy minikube users
// returns map of components to wait for
func interpretWaitFlag(cmd cobra.Command) map[string]bool {
	if !cmd.Flags().Changed(waitComponents) {
		klog.Infof("Wait components to verify : %+v", kverify.DefaultComponents)
		return kverify.DefaultComponents
	}

	waitFlags, err := cmd.Flags().GetStringSlice(waitComponents)
	if err != nil {
		klog.Warningf("Failed to read --wait from flags: %v.\n Moving on will use the default wait components: %+v", err, kverify.DefaultComponents)
		return kverify.DefaultComponents
	}

	if len(waitFlags) == 1 {
		// respecting legacy flag before minikube 1.9.0, wait flag was boolean
		if waitFlags[0] == "false" || waitFlags[0] == "none" {
			klog.Infof("Waiting for no components: %+v", kverify.NoComponents)
			return kverify.NoComponents
		}
		// respecting legacy flag before minikube 1.9.0, wait flag was boolean
		if waitFlags[0] == "true" || waitFlags[0] == "all" {
			klog.Infof("Waiting for all components: %+v", kverify.AllComponents)
			return kverify.AllComponents
		}
	}

	waitComponents := kverify.NoComponents
	for _, wc := range waitFlags {
		seen := false
		for _, valid := range kverify.AllComponentsList {
			if wc == valid {
				waitComponents[wc] = true
				seen = true
				continue
			}
		}
		if !seen {
			klog.Warningf("The value %q is invalid for --wait flag. valid options are %q", wc, strings.Join(kverify.AllComponentsList, ","))
		}
	}
	klog.Infof("Waiting for components: %+v", waitComponents)
	return waitComponents
}
