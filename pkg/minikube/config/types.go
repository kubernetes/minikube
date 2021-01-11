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

package config

import (
	"net"
	"time"

	"github.com/blang/semver"
)

// Profile represents a minikube profile
type Profile struct {
	Name   string
	Status string // running, stopped, paused, unknown
	Config *ClusterConfig
}

// ClusterConfig contains the parameters used to start a cluster.
type ClusterConfig struct {
	Name                    string
	KeepContext             bool   // used by start and profile command to or not to switch kubectl's current context
	EmbedCerts              bool   // used by kubeconfig.Setup
	MinikubeISO             string // ISO used for VM-drivers.
	KicBaseImage            string // base-image used for docker/podman drivers.
	Memory                  int
	CPUs                    int
	DiskSize                int
	VMDriver                string // Legacy use only
	Driver                  string
	HyperkitVpnKitSock      string   // Only used by the Hyperkit driver
	HyperkitVSockPorts      []string // Only used by the Hyperkit driver
	DockerEnv               []string // Each entry is formatted as KEY=VALUE.
	ContainerVolumeMounts   []string // Only used by container drivers: Docker, Podman
	InsecureRegistry        []string
	RegistryMirror          []string
	HostOnlyCIDR            string // Only used by the virtualbox driver
	HypervVirtualSwitch     string
	HypervUseExternalSwitch bool
	HypervExternalAdapter   string
	KVMNetwork              string   // Only used by the KVM driver
	KVMQemuURI              string   // Only used by kvm2
	KVMGPU                  bool     // Only used by kvm2
	KVMHidden               bool     // Only used by kvm2
	DockerOpt               []string // Each entry is formatted as KEY=VALUE.
	DisableDriverMounts     bool     // Only used by virtualbox
	NFSShare                []string
	NFSSharesRoot           string
	UUID                    string // Only used by hyperkit to restore the mac address
	NoVTXCheck              bool   // Only used by virtualbox
	DNSProxy                bool   // Only used by virtualbox
	HostDNSResolver         bool   // Only used by virtualbox
	HostOnlyNicType         string // Only used by virtualbox
	NatNicType              string // Only used by virtualbox
	KubernetesConfig        KubernetesConfig
	Nodes                   []Node
	Addons                  map[string]bool
	VerifyComponents        map[string]bool // map of components to verify and wait for after start.
	StartHostTimeout        time.Duration
	ScheduledStop           *ScheduledStopConfig
	ExposedPorts            []string // Only used by the docker and podman driver
	Network                 string   // only used by docker driver
	MultiNodeRequested      bool
}

// KubernetesConfig contains the parameters used to configure the VM Kubernetes.
type KubernetesConfig struct {
	KubernetesVersion   string
	ClusterName         string
	Namespace           string
	APIServerName       string
	APIServerNames      []string
	APIServerIPs        []net.IP
	DNSDomain           string
	ContainerRuntime    string
	CRISocket           string
	NetworkPlugin       string
	FeatureGates        string // https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/
	ServiceCIDR         string // the subnet which Kubernetes services will be deployed to
	ImageRepository     string
	LoadBalancerStartIP string // currently only used by MetalLB addon
	LoadBalancerEndIP   string // currently only used by MetalLB addon
	CustomIngressCert   string // used by Ingress addon
	ExtraOptions        ExtraOptionSlice

	ShouldLoadCachedImages bool

	EnableDefaultCNI bool   // deprecated in preference to CNI
	CNI              string // CNI to use

	// We need to keep these in the short term for backwards compatibility
	NodeIP   string
	NodePort int
	NodeName string
}

// Node contains information about specific nodes in a cluster
type Node struct {
	Name                string
	IP                  string
	Port                int
	KubernetesVersion   string
	PrimaryControlPlane bool
	ControlPlane        bool
	Worker              bool
}

// VersionedExtraOption holds information on flags to apply to a specific range
// of versions
type VersionedExtraOption struct {
	// Special Cases:
	//
	// If LessThanOrEqual and GreaterThanOrEqual are both nil, the flag will be applied
	// to all versions
	//
	// If LessThanOrEqual == GreaterThanOrEqual, the flag will only be applied to that
	// specific version

	// The flag and component that will be set
	Option ExtraOption

	// This flag will only be applied to versions before or equal to this version
	// If it is the default value, it will have no upper bound on versions the
	// flag is applied to
	LessThanOrEqual semver.Version

	// The flag will only be applied to versions after or equal to this version
	// If it is the default value, it will have no lower bound on versions the
	// flag is applied to
	GreaterThanOrEqual semver.Version
}

// ScheduledStopConfig contains information around scheduled stop
// not yet used, will be used to show status of scheduled stop
type ScheduledStopConfig struct {
	InitiationTime int64
	Duration       time.Duration
}
