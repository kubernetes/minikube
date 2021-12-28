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
	// DefaultAddonRegistryAssets assets for the default addons registry
	// go:embed addon-registry.yaml
	DefaultAddonRegistryAssets embed.FS

	// AutoPauseAssets assets for auto-pause addon
	//go:embed auto-pause/*.tmpl
	//go:embed auto-pause/unpause.lua
	//go:embed auto-pause/auto-pause.addon.yaml
	AutoPauseAssets embed.FS

	// DashboardAssets assets for dashboard addon
	//go:embed dashboard/*.yaml dashboard/*.tmpl
	//go:embed dashboard/dashboard.addon.yaml
	DashboardAssets embed.FS

	// DefaultStorageClassAssets assets for default-storageclass addon
	//go:embed storageclass/storageclass.yaml.tmpl
	//go:embed storageclass/default-storageclass.addon.yaml
	DefaultStorageClassAssets embed.FS

	// PodSecurityPolicyAssets assets for pod-security-policy addon
	//go:embed pod-security-policy/pod-security-policy.yaml.tmpl
	//go:embed pod-security-policy/pod-security-policy.addon.yaml
	PodSecurityPolicyAssets embed.FS

	// StorageProvisionerAssets assets for storage-provisioner addon
	//go:embed storage-provisioner/storage-provisioner.yaml.tmpl
	//go:embed storage-provisioner/storage-provisioner.addon.yaml
	StorageProvisionerAssets embed.FS

	// StorageProvisionerGlusterAssets assets for storage-provisioner-gluster addon
	//go:embed storage-provisioner-gluster/*.tmpl
	//go:embed storage-provisioner-gluster/storage-provisioner-gluster.addon.yaml
	StorageProvisionerGlusterAssets embed.FS

	// EfkAssets assets for efk addon
	//go:embed efk/*.tmpl
	//go:embed efk/efk.addon.yaml
	EfkAssets embed.FS

	// IngressAssets assets for ingress addon
	//go:embed ingress/*.tmpl
	//go:embed ingress/ingress.addon.yaml
	IngressAssets embed.FS

	// IstioProvisionerAssets assets for istio-provisioner addon
	//go:embed istio-provisioner/istio-operator.yaml.tmpl
	//go:embed istio-provisioner/istio-provisioner.addon.yaml
	IstioProvisionerAssets embed.FS

	// IstioAssets assets for istio addon
	//go:embed istio/istio-default-profile.yaml.tmpl
	//go:embed istio/istio.addon.yaml
	IstioAssets embed.FS

	// KubevirtAssets assets for kubevirt addon
	//go:embed kubevirt/pod.yaml.tmpl
	//go:embed kubevirt/kubevirt.addon.yaml
	KubevirtAssets embed.FS

	// MetricsServerAssets assets for metrics-server addon
	//go:embed metrics-server/*.tmpl
	//go:embed metrics-server/metrics-server.addon.yaml
	MetricsServerAssets embed.FS

	// OlmAssets assets for olm addon
	//go:embed olm/*.tmpl
	//go:embed olm/olm.addon.yaml
	OlmAssets embed.FS

	// RegistryAssets assets for registry addon
	//go:embed registry/*.tmpl
	//go:embed registry/registry.addon.yaml
	RegistryAssets embed.FS

	// RegistryCredsAssets assets for registry-creds addon
	//go:embed registry-creds/registry-creds-rc.yaml.tmpl
	//go:embed registry-creds/registry-creds.addon.yaml
	RegistryCredsAssets embed.FS

	// RegistryAliasesAssets assets for registry-aliases addon
	//go:embed registry-aliases/*.tmpl
	//go:embed registry-aliases/registry-aliases.addon.yaml
	RegistryAliasesAssets embed.FS

	// FreshpodAssets assets for freshpod addon
	//go:embed freshpod/freshpod-rc.yaml.tmpl
	//go:embed freshpod/freshpod.addon.yaml
	FreshpodAssets embed.FS

	// NvidiaDriverInstallerAssets assets for nvidia-driver-installer addon
	//go:embed gpu/nvidia-driver-installer.yaml.tmpl
	//go:embed gpu/nvidia-driver-installer.addon.yaml
	NvidiaDriverInstallerAssets embed.FS

	// NvidiaGpuDevicePluginAssets assets for nvidia-gpu-device-plugin addon
	//go:embed gpu/nvidia-gpu-device-plugin.yaml.tmpl
	//go:embed gpu/nvidia-gpu-device-plugin.addon.yaml
	NvidiaGpuDevicePluginAssets embed.FS

	// LogviewerAssets assets for logviewer addon
	//go:embed logviewer/*.tmpl
	//go:embed logviewer/logviewer.addon.yaml
	LogviewerAssets embed.FS

	// GvisorAssets assets for gvisor addon
	//go:embed gvisor/*.tmpl gvisor/*.toml
	//go:embed gvisor/gvisor.addon.yaml
	GvisorAssets embed.FS

	// HelmTillerAssets assets for helm-tiller addon
	//go:embed helm-tiller/*.tmpl
	//go:embed helm-tiller/helm-tiller.addon.yaml
	HelmTillerAssets embed.FS

	// IngressDNSAssets assets for ingress-dns addon
	//go:embed ingress-dns/ingress-dns-pod.yaml.tmpl
	//go:embed ingress-dns/ingress-dns.addon.yaml
	IngressDNSAssets embed.FS

	// MetallbAssets assets for metallb addon
	//go:embed metallb/*.tmpl
	//go:embed metallb/metallb.addon.yaml
	MetallbAssets embed.FS

	// AmbassadorAssets assets for ambassador addon
	//go:embed ambassador/*.tmpl
	//go:embed ambassador/ambassador.addon.yaml
	AmbassadorAssets embed.FS

	// GcpAuthAssets assets for gcp-auth addon
	//go:embed gcp-auth/*.tmpl
	//go:embed gcp-auth/gcp-auth.addon.yaml
	GcpAuthAssets embed.FS

	// VolumeSnapshotsAssets assets for volumesnapshots addon
	//go:embed volumesnapshots/*.tmpl
	//go:embed volumesnapshots/volumesnapshots.addon.yaml
	VolumeSnapshotsAssets embed.FS

	// CsiHostpathDriverAssets assets for csi-hostpath-driver addon
	//go:embed csi-hostpath-driver/deploy/*.tmpl csi-hostpath-driver/rbac/*.tmpl
	//go:embed csi-hostpath-driver/csi-hostpath-driver.addon.yaml
	CsiHostpathDriverAssets embed.FS

	// PortainerAssets assets for portainer addon
	//go:embed portainer/portainer.yaml.tmpl
	//go:embed portainer/portainer.addon.yaml
	PortainerAssets embed.FS

	// AliyunMirror assets for aliyun_mirror.json
	//go:embed aliyun_mirror.json
	AliyunMirror embed.FS
)
