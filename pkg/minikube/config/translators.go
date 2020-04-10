package config

import (
	v162 "k8s.io/minikube/pkg/minikube/config/v162"
)

func tryTranslate(vcontran []VersionConfigTranslator) (interface{}, error) {

	// Get the previous version translator
	previousVersion := vcontran[0]
	previousVersionConfig, err := previousVersion.TryLoadFromFile()
	// if the translator couldn't load the config from the file, then it's probably even an older version, so let's recurse again deeper
	if (err != nil || !previousVersion.IsValid(previousVersionConfig)) && len(vcontran) > 1 {
		var err error
		previousVersionConfig, err = tryTranslate(vcontran[1:])
		if err != nil {
			// Ah too bad, even the older versions couldn't translate it, this would bubble up to the end.
			return nil, err
		}
	}
	// Yes! The previous recurse iteration returned a successful and valid previousVersionConfig. Now let's translate it to the next version
	translatedConfig, err := previousVersion.TranslateToNextVersion(previousVersionConfig)
	return translatedConfig, err
}

var versionConfigTranslators = []VersionConfigTranslator{
	{
		TryLoadFromFile: func() (interface{}, error) {
			return nil, nil
		},
		TranslateToNextVersion: func(config interface{}) (interface{}, error) {
			return translateFrom163ToCurrent(config.(v162.MachineConfig))
		},
	},
}

type VersionConfigTranslator struct {
	TryLoadFromFile        tryLoadFromFile
	TranslateToNextVersion translateToNextVersion
	IsValid                isValid
}
type translateToNextVersion func(interface{}) (interface{}, error)
type tryLoadFromFile func() (interface{}, error)
type isValid func(interface{}) bool

func translateFrom163ToCurrent(oldConfig v162.MachineConfig) (*ClusterConfig, error) {
	hypervUseExternalSwitch := (oldConfig.HypervVirtualSwitch != "")

	return &ClusterConfig{
		Name:                    oldConfig.Name,
		KeepContext:             oldConfig.KeepContext,
		EmbedCerts:              oldConfig.EmbedCerts,
		MinikubeISO:             oldConfig.MinikubeISO,
		Memory:                  oldConfig.Memory,
		CPUs:                    oldConfig.CPUs,
		DiskSize:                oldConfig.DiskSize,
		Driver:                  oldConfig.VMDriver,
		HyperkitVpnKitSock:      oldConfig.HyperkitVpnKitSock,
		HyperkitVSockPorts:      oldConfig.HyperkitVSockPorts,
		DockerEnv:               oldConfig.DockerEnv,
		InsecureRegistry:        oldConfig.InsecureRegistry,
		RegistryMirror:          oldConfig.RegistryMirror,
		HostOnlyCIDR:            oldConfig.HostOnlyCIDR,
		HypervUseExternalSwitch: hypervUseExternalSwitch,
		KVMNetwork:              oldConfig.KVMNetwork,
		KVMQemuURI:              oldConfig.KVMQemuURI,
		KVMGPU:                  oldConfig.KVMGPU,
		KVMHidden:               oldConfig.KVMHidden,
		DockerOpt:               oldConfig.DockerOpt,
		DisableDriverMounts:     oldConfig.DisableDriverMounts,
		NFSShare:                oldConfig.NFSShare,
		NFSSharesRoot:           oldConfig.NFSSharesRoot,
		UUID:                    oldConfig.UUID,
		NoVTXCheck:              oldConfig.NoVTXCheck,
		DNSProxy:                oldConfig.DNSProxy,
		HostDNSResolver:         oldConfig.HostDNSResolver,
		HostOnlyNicType:         oldConfig.HostOnlyNicType,
		NatNicType:              oldConfig.NatNicType,
		Nodes: []Node{Node{
			Name:              oldConfig.KubernetesConfig.NodeName,
			ControlPlane:      true, //TODO: make sure that this is correct
			IP:                oldConfig.KubernetesConfig.NodeIP,
			KubernetesVersion: oldConfig.KubernetesConfig.KubernetesVersion,
			Port:              oldConfig.KubernetesConfig.NodePort,
			Worker:            true,
		}},
		KubernetesConfig: KubernetesConfig{
			KubernetesVersion: oldConfig.KubernetesConfig.KubernetesVersion,
			ClusterName:       oldConfig.Name,
			APIServerName:     oldConfig.KubernetesConfig.APIServerName,
			APIServerNames:    oldConfig.KubernetesConfig.APIServerNames,
			APIServerIPs:      oldConfig.KubernetesConfig.APIServerIPs,
			DNSDomain:         oldConfig.KubernetesConfig.DNSDomain,
			ContainerRuntime:  oldConfig.KubernetesConfig.ContainerRuntime,
			CRISocket:         oldConfig.KubernetesConfig.CRISocket,
			NetworkPlugin:     oldConfig.KubernetesConfig.NetworkPlugin,
			FeatureGates:      oldConfig.KubernetesConfig.FeatureGates,
			ServiceCIDR:       oldConfig.KubernetesConfig.ServiceCIDR,
			ImageRepository:   oldConfig.KubernetesConfig.ImageRepository,
			// ExtraOptions:           oldConfig.KubernetesConfig.ExtraOptions,
			ShouldLoadCachedImages: oldConfig.KubernetesConfig.ShouldLoadCachedImages,
			EnableDefaultCNI:       oldConfig.KubernetesConfig.EnableDefaultCNI,
			NodeIP:                 oldConfig.KubernetesConfig.NodeIP,
			NodePort:               oldConfig.KubernetesConfig.NodePort,
			NodeName:               oldConfig.KubernetesConfig.NodeName,
		},
	}, nil
}

func translateFrom152ToCurrent(oldConfig v162.MachineConfig) (*ClusterConfig, error) {
	hypervUseExternalSwitch := (oldConfig.HypervVirtualSwitch != "")

	return &ClusterConfig{
		Name:                    oldConfig.Name,
		KeepContext:             oldConfig.KeepContext,
		EmbedCerts:              oldConfig.EmbedCerts,
		MinikubeISO:             oldConfig.MinikubeISO,
		Memory:                  oldConfig.Memory,
		CPUs:                    oldConfig.CPUs,
		DiskSize:                oldConfig.DiskSize,
		Driver:                  oldConfig.VMDriver,
		HyperkitVpnKitSock:      oldConfig.HyperkitVpnKitSock,
		HyperkitVSockPorts:      oldConfig.HyperkitVSockPorts,
		DockerEnv:               oldConfig.DockerEnv,
		InsecureRegistry:        oldConfig.InsecureRegistry,
		RegistryMirror:          oldConfig.RegistryMirror,
		HostOnlyCIDR:            oldConfig.HostOnlyCIDR,
		HypervUseExternalSwitch: hypervUseExternalSwitch,
		KVMNetwork:              oldConfig.KVMNetwork,
		KVMQemuURI:              oldConfig.KVMQemuURI,
		KVMGPU:                  oldConfig.KVMGPU,
		KVMHidden:               oldConfig.KVMHidden,
		DockerOpt:               oldConfig.DockerOpt,
		DisableDriverMounts:     oldConfig.DisableDriverMounts,
		NFSShare:                oldConfig.NFSShare,
		NFSSharesRoot:           oldConfig.NFSSharesRoot,
		UUID:                    oldConfig.UUID,
		NoVTXCheck:              oldConfig.NoVTXCheck,
		DNSProxy:                oldConfig.DNSProxy,
		HostDNSResolver:         oldConfig.HostDNSResolver,
		HostOnlyNicType:         oldConfig.HostOnlyNicType,
		NatNicType:              oldConfig.NatNicType,
		Nodes: []Node{{
			Name:              oldConfig.KubernetesConfig.NodeName,
			ControlPlane:      true, //TODO: make sure that this is correct
			IP:                oldConfig.KubernetesConfig.NodeIP,
			KubernetesVersion: oldConfig.KubernetesConfig.KubernetesVersion,
			Port:              oldConfig.KubernetesConfig.NodePort,
			Worker:            true,
		}},
		KubernetesConfig: KubernetesConfig{
			KubernetesVersion: oldConfig.KubernetesConfig.KubernetesVersion,
			ClusterName:       oldConfig.Name,
			APIServerName:     oldConfig.KubernetesConfig.APIServerName,
			APIServerNames:    oldConfig.KubernetesConfig.APIServerNames,
			APIServerIPs:      oldConfig.KubernetesConfig.APIServerIPs,
			DNSDomain:         oldConfig.KubernetesConfig.DNSDomain,
			ContainerRuntime:  oldConfig.KubernetesConfig.ContainerRuntime,
			CRISocket:         oldConfig.KubernetesConfig.CRISocket,
			NetworkPlugin:     oldConfig.KubernetesConfig.NetworkPlugin,
			FeatureGates:      oldConfig.KubernetesConfig.FeatureGates,
			ServiceCIDR:       oldConfig.KubernetesConfig.ServiceCIDR,
			ImageRepository:   oldConfig.KubernetesConfig.ImageRepository,
			// ExtraOptions:           oldConfig.KubernetesConfig.ExtraOptions,
			ShouldLoadCachedImages: oldConfig.KubernetesConfig.ShouldLoadCachedImages,
			EnableDefaultCNI:       oldConfig.KubernetesConfig.EnableDefaultCNI,
			NodeIP:                 oldConfig.KubernetesConfig.NodeIP,
			NodePort:               oldConfig.KubernetesConfig.NodePort,
			NodeName:               oldConfig.KubernetesConfig.NodeName,
		},
	}, nil
}
