/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package addons

import "embed"

var (
	// AutoPauseAssets assets for auto-pause addon
	//go:embed auto-pause/*.tmpl
	//go:embed auto-pause/unpause.lua
	AutoPauseAssets embed.FS

	// DashboardAssets assets for dashboard addon
	//go:embed dashboard/*.yaml dashboard/*.tmpl
	DashboardAssets embed.FS

	// DefaultStorageClassAssets assets for default-storageclass addon
	//go:embed storageclass/storageclass.yaml.tmpl
	DefaultStorageClassAssets embed.FS

	// PodSecurityPolicyAssets assets for pod-security-policy addon
	//go:embed pod-security-policy/pod-security-policy.yaml.tmpl
	PodSecurityPolicyAssets embed.FS

	// StorageProvisionerAssets assets for storage-provisioner addon
	//go:embed storage-provisioner/storage-provisioner.yaml.tmpl
	StorageProvisionerAssets embed.FS

	// StorageProvisionerGlusterAssets assets for storage-provisioner-gluster addon
	//go:embed storage-provisioner-gluster/*.tmpl
	StorageProvisionerGlusterAssets embed.FS

	// EfkAssets assets for efk addon
	//go:embed efk/*.tmpl
	EfkAssets embed.FS

	// IngressAssets assets for ingress addon
	//go:embed ingress/*.tmpl
	IngressAssets embed.FS

	// IstioProvisionerAssets assets for istio-provisioner addon
	//go:embed istio-provisioner/istio-operator.yaml.tmpl
	IstioProvisionerAssets embed.FS

	// IstioAssets assets for istio addon
	//go:embed istio/istio-default-profile.yaml.tmpl
	IstioAssets embed.FS

	// KubevirtAssets assets for kubevirt addon
	//go:embed kubevirt/pod.yaml.tmpl
	KubevirtAssets embed.FS

	// MetricsServerAssets assets for metrics-server addon
	//go:embed metrics-server/*.tmpl
	MetricsServerAssets embed.FS

	// OlmAssets assets for olm addon
	//go:embed olm/*.tmpl
	OlmAssets embed.FS

	// RegistryAssets assets for registry addon
	//go:embed registry/*.tmpl
	RegistryAssets embed.FS

	// RegistryCredsAssets assets for registry-creds addon
	//go:embed registry-creds/registry-creds-rc.yaml.tmpl
	RegistryCredsAssets embed.FS

	// RegistryAliasesAssets assets for registry-aliases addon
	//go:embed registry-aliases/*.tmpl
	RegistryAliasesAssets embed.FS

	// FreshpodAssets assets for freshpod addon
	//go:embed freshpod/freshpod-rc.yaml.tmpl
	FreshpodAssets embed.FS

	// NvidiaDriverInstallerAssets assets for nvidia-driver-installer addon
	//go:embed gpu/nvidia-driver-installer.yaml.tmpl
	NvidiaDriverInstallerAssets embed.FS

	// NvidiaGpuDevicePluginAssets assets for nvidia-gpu-device-plugin addon
	//go:embed gpu/nvidia-gpu-device-plugin.yaml.tmpl
	NvidiaGpuDevicePluginAssets embed.FS

	// LogviewerAssets assets for logviewer addon
	//go:embed logviewer/*.tmpl
	LogviewerAssets embed.FS

	// GvisorAssets assets for gvisor addon
	//go:embed gvisor/*.tmpl gvisor/*.toml
	GvisorAssets embed.FS

	// HelmTillerAssets assets for helm-tiller addon
	//go:embed helm-tiller/*.tmpl
	HelmTillerAssets embed.FS

	// IngressDNSAssets assets for ingress-dns addon
	//go:embed ingress-dns/ingress-dns-pod.yaml.tmpl
	IngressDNSAssets embed.FS

	// MetallbAssets assets for metallb addon
	//go:embed metallb/*.tmpl
	MetallbAssets embed.FS

	// AmbassadorAssets assets for ambassador addon
	//go:embed ambassador/*.tmpl
	AmbassadorAssets embed.FS

	// GcpAuthAssets assets for gcp-auth addon
	//go:embed gcp-auth/*.tmpl
	GcpAuthAssets embed.FS

	// VolumeSnapshotsAssets assets for volumesnapshots addon
	//go:embed volumesnapshots/*.tmpl
	VolumeSnapshotsAssets embed.FS

	// CsiHostpathDriverAssets assets for csi-hostpath-driver addon
	//go:embed csi-hostpath-driver/deploy/*.tmpl csi-hostpath-driver/rbac/*.tmpl
	CsiHostpathDriverAssets embed.FS

	// PortainerAssets assets for portainer addon
	//go:embed portainer/portainer.yaml.tmpl
	PortainerAssets embed.FS
)
