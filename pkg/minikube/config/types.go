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

	"k8s.io/minikube/pkg/util"
)

// Profile represents a minikube profile
type Profile struct {
	Name   string
	Config *Config
}

// Config contains machine and k8s config
type Config struct {
	MachineConfig    MachineConfig
	KubernetesConfig KubernetesConfig
}

// MachineConfig contains the parameters used to start a cluster.
type MachineConfig struct {
	KeepContext         bool // used by start and profile command to or not to switch kubectl's current context
	MinikubeISO         string
	Memory              int
	CPUs                int
	DiskSize            int
	VMDriver            string
	ContainerRuntime    string
	HyperkitVpnKitSock  string   // Only used by the Hyperkit driver
	HyperkitVSockPorts  []string // Only used by the Hyperkit driver
	DockerEnv           []string // Each entry is formatted as KEY=VALUE.
	InsecureRegistry    []string
	RegistryMirror      []string
	HostOnlyCIDR        string // Only used by the virtualbox driver
	HypervVirtualSwitch string
	KVMNetwork          string             // Only used by the KVM driver
	KVMQemuURI          string             // Only used by kvm2
	KVMGPU              bool               // Only used by kvm2
	KVMHidden           bool               // Only used by kvm2
	Downloader          util.ISODownloader `json:"-"`
	DockerOpt           []string           // Each entry is formatted as KEY=VALUE.
	DisableDriverMounts bool               // Only used by virtualbox
	NFSShare            []string
	NFSSharesRoot       string
	UUID                string // Only used by hyperkit to restore the mac address
	NoVTXCheck          bool   // Only used by virtualbox
	DNSProxy            bool   // Only used by virtualbox
	HostDNSResolver     bool   // Only used by virtualbox
}

// KubernetesConfig contains the parameters used to configure the VM Kubernetes.
type KubernetesConfig struct {
	KubernetesVersion string
	NodeIP            string
	NodePort          int
	NodeName          string
	APIServerName     string
	APIServerNames    []string
	APIServerIPs      []net.IP
	DNSDomain         string
	ContainerRuntime  string
	CRISocket         string
	NetworkPlugin     string
	FeatureGates      string
	ServiceCIDR       string
	ImageRepository   string
	ExtraOptions      util.ExtraOptionSlice

	ShouldLoadCachedImages bool
	EnableDefaultCNI       bool
}
