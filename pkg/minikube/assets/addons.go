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

package assets

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/deploy/addons"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/version"
)

// Addon is a named list of assets, that can be enabled
type Addon struct {
	Assets             []*BinAsset
	enabled            bool
	addonName          string
	Maintainer         string
	VerifiedMaintainer string
	Docs               string
	Images             map[string]string

	// Registries currently only shows the default registry of images
	Registries map[string]string
}

// NetworkInfo contains control plane node IP address used for add on template
type NetworkInfo struct {
	ControlPlaneNodeIP   string
	ControlPlaneNodePort int
}

// NewAddon creates a new Addon
func NewAddon(assets []*BinAsset, enabled bool, addonName, maintainer, verifiedMaintainer, docs string, images, registries map[string]string) *Addon {
	return &Addon{
		Assets:             assets,
		enabled:            enabled,
		addonName:          addonName,
		Maintainer:         maintainer,
		VerifiedMaintainer: verifiedMaintainer,
		Docs:               docs,
		Images:             images,
		Registries:         registries,
	}
}

// Name gets the addon name
func (a *Addon) Name() string {
	return a.addonName
}

// IsEnabled checks if an Addon is enabled for the given profile
func (a *Addon) IsEnabled(cc *config.ClusterConfig) bool {
	status, ok := cc.Addons[a.Name()]
	if ok {
		return status
	}

	// Return the default unconfigured state of the addon
	return a.enabled
}

// Addons is the list of addons
// TODO: Make dynamically loadable: move this data to a .yaml file within each addon directory
var Addons = map[string]*Addon{
	"auto-pause": NewAddon([]*BinAsset{
		MustBinAsset(
			addons.AutoPauseAssets,
			"auto-pause/auto-pause.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"auto-pause.yaml",
			"0640"),
		MustBinAsset(
			addons.AutoPauseAssets,
			"auto-pause/auto-pause-hook.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"auto-pause-hook.yaml",
			"0640"),
		MustBinAsset(
			addons.AutoPauseAssets,
			"auto-pause/haproxy.cfg.tmpl",
			vmpath.GuestPersistentDir,
			"haproxy.cfg",
			"0640"),
		MustBinAsset(
			addons.AutoPauseAssets,
			"auto-pause/unpause.lua",
			vmpath.GuestPersistentDir,
			"unpause.lua",
			"0640"),
		MustBinAsset(
			addons.AutoPauseAssets,
			"auto-pause/auto-pause.service.tmpl",
			"/etc/systemd/system/",
			"auto-pause.service",
			"0640"),

		// GuestPersistentDir
	}, false, "auto-pause", "Google", "", "", map[string]string{
		"AutoPauseHook": "k8s-minikube/auto-pause-hook:v0.0.3@sha256:2ff12878e569802afa75baf65701ab913a88fe61a1cbb22fdc852b1345af2a5c",
	}, map[string]string{
		"AutoPauseHook": "gcr.io",
	}),
	"dashboard": NewAddon([]*BinAsset{
		// We want to create the kubernetes-dashboard ns first so that every subsequent object can be created
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-ns.yaml", vmpath.GuestAddonsDir, "dashboard-ns.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-clusterrole.yaml", vmpath.GuestAddonsDir, "dashboard-clusterrole.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-clusterrolebinding.yaml", vmpath.GuestAddonsDir, "dashboard-clusterrolebinding.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-configmap.yaml", vmpath.GuestAddonsDir, "dashboard-configmap.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-dp.yaml.tmpl", vmpath.GuestAddonsDir, "dashboard-dp.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-role.yaml", vmpath.GuestAddonsDir, "dashboard-role.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-rolebinding.yaml", vmpath.GuestAddonsDir, "dashboard-rolebinding.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-sa.yaml", vmpath.GuestAddonsDir, "dashboard-sa.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-secret.yaml", vmpath.GuestAddonsDir, "dashboard-secret.yaml", "0640"),
		MustBinAsset(addons.DashboardAssets, "dashboard/dashboard-svc.yaml", vmpath.GuestAddonsDir, "dashboard-svc.yaml", "0640"),
	}, false, "dashboard", "Kubernetes", "", "https://minikube.sigs.k8s.io/docs/handbook/dashboard/", map[string]string{
		"Dashboard":      "kubernetesui/dashboard:v2.7.0@sha256:2e500d29e9d5f4a086b908eb8dfe7ecac57d2ab09d65b24f588b1d449841ef93",
		"MetricsScraper": "kubernetesui/metrics-scraper:v1.0.8@sha256:76049887f07a0476dc93efc2d3569b9529bf982b22d29f356092ce206e98765c",
	}, map[string]string{
		"Dashboard":      "docker.io",
		"MetricsScraper": "docker.io",
	}),
	"default-storageclass": NewAddon([]*BinAsset{
		MustBinAsset(addons.DefaultStorageClassAssets,
			"storageclass/storageclass.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storageclass.yaml",
			"0640"),
	}, true, "default-storageclass", "Kubernetes", "", "https://minikube.sigs.k8s.io/docs/handbook/persistent_volumes/", nil, nil),
	"pod-security-policy": NewAddon([]*BinAsset{
		MustBinAsset(addons.PodSecurityPolicyAssets,
			"pod-security-policy/pod-security-policy.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"pod-security-policy.yaml",
			"0640"),
	}, false, "pod-security-policy", "3rd party (unknown)", "", "", nil, nil),
	"storage-provisioner": NewAddon([]*BinAsset{
		MustBinAsset(addons.StorageProvisionerAssets,
			"storage-provisioner/storage-provisioner.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storage-provisioner.yaml",
			"0640"),
	}, true, "storage-provisioner", "Google", "", "", map[string]string{
		"StorageProvisioner": fmt.Sprintf("k8s-minikube/storage-provisioner:%s", version.GetStorageProvisionerVersion()),
	}, map[string]string{
		"StorageProvisioner": "gcr.io",
	}),
	"storage-provisioner-gluster": NewAddon([]*BinAsset{
		MustBinAsset(addons.StorageProvisionerGlusterAssets,
			"storage-provisioner-gluster/storage-gluster-ns.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storage-gluster-ns.yaml",
			"0640"),
		MustBinAsset(addons.StorageProvisionerGlusterAssets,
			"storage-provisioner-gluster/glusterfs-daemonset.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"glusterfs-daemonset.yaml",
			"0640"),
		MustBinAsset(addons.StorageProvisionerGlusterAssets,
			"storage-provisioner-gluster/heketi-deployment.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"heketi-deployment.yaml",
			"0640"),
		MustBinAsset(addons.StorageProvisionerGlusterAssets,
			"storage-provisioner-gluster/storage-provisioner-glusterfile.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storage-provisioner-glusterfile.yaml",
			"0640"),
	}, false, "storage-provisioner-gluster", "3rd party (Gluster)", "", "", map[string]string{
		"Heketi":                 "heketi/heketi:10@sha256:76d5a6a3b7cf083d1e99efa1c15abedbc5c8b73bef3ade299ce9a4c16c9660f8",
		"GlusterfileProvisioner": "gluster/glusterfile-provisioner:latest@sha256:9961a35cb3f06701958e202324141c30024b195579e5eb1704599659ddea5223",
		"GlusterfsServer":        "gluster/gluster-centos:latest@sha256:8167034b9abf2d16581f3f4571507ce7d716fb58b927d7627ef72264f802e908",
	}, map[string]string{
		"Heketi":                 "docker.io",
		"GlusterfsServer":        "docker.io",
		"GlusterfileProvisioner": "docker.io",
	}),
	"efk": NewAddon([]*BinAsset{
		MustBinAsset(addons.EfkAssets,
			"efk/elasticsearch-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"elasticsearch-rc.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/elasticsearch-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"elasticsearch-svc.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/fluentd-es-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"fluentd-es-rc.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/fluentd-es-configmap.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"fluentd-es-configmap.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/kibana-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"kibana-rc.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/kibana-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"kibana-svc.yaml",
			"0640"),
	}, false, "efk", "3rd party (Elastic)", "", "", map[string]string{
		"Elasticsearch":        "elasticsearch:v5.6.2@sha256:7e95b32a7a2aad0c0db5c881e4a1ce8b7e53236144ae9d9cfb5fbe5608af4ab2",
		"FluentdElasticsearch": "fluentd-elasticsearch:v2.0.2@sha256:d0480bbf2d0de2344036fa3f7034cf7b4b98025a89c71d7f1f1845ac0e7d5a97",
		"Alpine":               "alpine:3.6@sha256:66790a2b79e1ea3e1dabac43990c54aca5d1ddf268d9a5a0285e4167c8b24475",
		"Kibana":               "kibana/kibana:5.6.2@sha256:cd948a9bda4622f1437afc4a3e78be6c8c25fc62f40aa0376f3d690f2436568f",
	}, map[string]string{
		"Elasticsearch":        "k8s.gcr.io",
		"FluentdElasticsearch": "k8s.gcr.io",
		"Kibana":               "docker.elastic.co",
		"Alpine":               "docker.io",
	}),
	"ingress": NewAddon([]*BinAsset{
		MustBinAsset(addons.IngressAssets,
			"ingress/ingress-deploy.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ingress-deploy.yaml",
			"0640"),
	}, false, "ingress", "Kubernetes", "", "https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/", map[string]string{
		// https://github.com/kubernetes/ingress-nginx/blob/63dbbdbb3a099417f411fbd1d684fa2a287c96cd/deploy/static/provider/kind/deploy.yaml#L458
		"IngressController": "ingress-nginx/controller:v1.5.1@sha256:4ba73c697770664c1e00e9f968de14e08f606ff961c76e5d7033a4a9c593c629",
		// https://github.com/kubernetes/ingress-nginx/blob/63dbbdbb3a099417f411fbd1d684fa2a287c96cd/deploy/static/provider/kind/deploy.yaml#L565
		"KubeWebhookCertgenCreate": "ingress-nginx/kube-webhook-certgen:v20220916-gd32f8c343@sha256:39c5b2e3310dc4264d638ad28d9d1d96c4cbb2b2dcfb52368fe4e3c63f61e10f",
		// https://github.com/kubernetes/ingress-nginx/blob/63dbbdbb3a099417f411fbd1d684fa2a287c96cd/deploy/static/provider/kind/deploy.yaml#L614
		"KubeWebhookCertgenPatch": "ingress-nginx/kube-webhook-certgen:v20220916-gd32f8c343@sha256:39c5b2e3310dc4264d638ad28d9d1d96c4cbb2b2dcfb52368fe4e3c63f61e10f",
	}, map[string]string{
		"IngressController":        "registry.k8s.io",
		"KubeWebhookCertgenCreate": "registry.k8s.io",
		"KubeWebhookCertgenPatch":  "registry.k8s.io",
	}),
	"istio-provisioner": NewAddon([]*BinAsset{
		MustBinAsset(addons.IstioProvisionerAssets,
			"istio-provisioner/istio-operator.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"istio-operator.yaml",
			"0640"),
	}, false, "istio-provisioner", "3rd party (Istio)", "", "https://istio.io/latest/docs/setup/platform-setup/minikube/", map[string]string{
		"IstioOperator": "istio/operator:1.12.2@sha256:42c7609872882cb88728a1592561b4046dac6d05b6002cbdc815b84c86a24f08",
	}, map[string]string{
		"IstioOperator": "docker.io",
	}),
	"istio": NewAddon([]*BinAsset{
		MustBinAsset(addons.IstioAssets,
			"istio/istio-default-profile.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"istio-default-profile.yaml",
			"0640"),
	}, false, "istio", "3rd party (Istio)", "", "https://istio.io/latest/docs/setup/platform-setup/minikube/", nil, nil),
	"kong": NewAddon([]*BinAsset{
		MustBinAsset(addons.KongAssets,
			"kong/kong-ingress-controller.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"kong-ingress-controller.yaml",
			"0640"),
	}, false, "kong", "3rd party (Kong HQ)", "", "https://minikube.sigs.k8s.io/docs/handbook/addons/kong-ingress/", map[string]string{
		"Kong":        "kong:2.7@sha256:4d3e93207305ace881fe9e95ac27717b6fbdd9e0ec1873c34e94908a4f4c9335",
		"KongIngress": "kong/kubernetes-ingress-controller:2.1.1@sha256:60e4102ab2da7f61e9c478747f0762d06a6166b5f300526b237ed7354c3cb4c8",
	}, map[string]string{
		"Kong":        "docker.io",
		"KongIngress": "docker.io",
	}),
	"kubevirt": NewAddon([]*BinAsset{
		MustBinAsset(addons.KubevirtAssets,
			"kubevirt/pod.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"pod.yaml",
			"0640"),
	}, false, "kubevirt", "3rd party (KubeVirt)", "", "https://minikube.sigs.k8s.io/docs/tutorials/kubevirt/", map[string]string{
		"Kubectl": "bitnami/kubectl:1.24.7@sha256:195f5a7a40cfb06e308701ae850abfa436d23baf9d39c0282298e540c9d07863",
	}, map[string]string{
		"Kubectl": "docker.io",
	}),
	"metrics-server": NewAddon([]*BinAsset{
		MustBinAsset(addons.MetricsServerAssets,
			"metrics-server/metrics-apiservice.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metrics-apiservice.yaml",
			"0640"),
		MustBinAsset(addons.MetricsServerAssets,
			"metrics-server/metrics-server-deployment.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metrics-server-deployment.yaml",
			"0640"),
		MustBinAsset(addons.MetricsServerAssets,
			"metrics-server/metrics-server-rbac.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metrics-server-rbac.yaml",
			"0640"),
		MustBinAsset(addons.MetricsServerAssets,
			"metrics-server/metrics-server-service.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metrics-server-service.yaml",
			"0640"),
	}, false, "metrics-server", "Kubernetes", "", "", map[string]string{
		"MetricsServer": "metrics-server/metrics-server:v0.6.2@sha256:f977ad859fb500c1302d9c3428c6271db031bb7431e7076213b676b345a88dc2",
	}, map[string]string{
		"MetricsServer": "registry.k8s.io",
	}),
	"olm": NewAddon([]*BinAsset{
		MustBinAsset(addons.OlmAssets,
			"olm/crds.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"crds.yaml",
			"0640"),
		MustBinAsset(addons.OlmAssets,
			"olm/olm.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"olm.yaml",
			"0640"),
	}, false, "olm", "3rd party (Operator Framework)", "", "", map[string]string{
		"OLM": "operator-framework/olm@sha256:e74b2ac57963c7f3ba19122a8c31c9f2a0deb3c0c5cac9e5323ccffd0ca198ed",
		// operator-framework/community-operators was deprecated: https://github.com/operator-framework/community-operators#repository-is-obsolete; switching to OperatorHub.io instead
		"UpstreamCommunityOperators": "operatorhubio/catalog@sha256:e08a1cd21fe72dd1be92be738b4bf1515298206dac5479c17a4b3ed119e30bd4",
	}, map[string]string{
		"OLM":                        "quay.io",
		"UpstreamCommunityOperators": "quay.io",
	}),
	"registry": NewAddon([]*BinAsset{
		MustBinAsset(addons.RegistryAssets,
			"registry/registry-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-rc.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAssets,
			"registry/registry-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-svc.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAssets,
			"registry/registry-proxy.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-proxy.yaml",
			"0640"),
	}, false, "registry", "Google", "", "", map[string]string{
		"Registry":          "registry:2.8.1@sha256:83bb78d7b28f1ac99c68133af32c93e9a1c149bcd3cb6e683a3ee56e312f1c96",
		"KubeRegistryProxy": "google_containers/kube-registry-proxy:0.4@sha256:1040f25a5273de0d72c54865a8efd47e3292de9fb8e5353e3fa76736b854f2da",
	}, map[string]string{
		"KubeRegistryProxy": "gcr.io",
		"Registry":          "docker.io",
	}),
	"registry-creds": NewAddon([]*BinAsset{
		MustBinAsset(addons.RegistryCredsAssets,
			"registry-creds/registry-creds-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-creds-rc.yaml",
			"0640"),
	}, false, "registry-creds", "3rd party (UPMC Enterprises)", "", "https://minikube.sigs.k8s.io/docs/handbook/registry/", map[string]string{
		"RegistryCreds": "upmcenterprises/registry-creds:1.10@sha256:93a633d4f2b76a1c66bf19c664dbddc56093a543de6d54320f19f585ccd7d605",
	}, map[string]string{
		"RegistryCreds": "docker.io",
	}),
	"registry-aliases": NewAddon([]*BinAsset{
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/registry-aliases-sa.tmpl",
			vmpath.GuestAddonsDir,
			"registry-aliases-sa.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/registry-aliases-sa-crb.tmpl",
			vmpath.GuestAddonsDir,
			"registry-aliases-sa-crb.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/registry-aliases-config.tmpl",
			vmpath.GuestAddonsDir,
			"registry-aliases-config.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/node-etc-hosts-update.tmpl",
			vmpath.GuestAddonsDir,
			"node-etc-hosts-update.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/patch-coredns-job.tmpl",
			vmpath.GuestAddonsDir,
			"patch-coredns-job.yaml",
			"0640"),
	}, false, "registry-aliases", "3rd party (unknown)", "", "", map[string]string{
		"CoreDNSPatcher": "rhdevelopers/core-dns-patcher@sha256:9220ff32f690c3d889a52afb59ca6fcbbdbd99e5370550cc6fd249adea8ed0a9",
		"Alpine":         "alpine:3.11@sha256:0bd0e9e03a022c3b0226667621da84fc9bf562a9056130424b5bfbd8bcb0397f",
		"Pause":          "google_containers/pause:3.1@sha256:f78411e19d84a252e53bff71a4407a5686c46983a2c2eeed83929b888179acea",
	}, map[string]string{
		"CoreDNSPatcher": "quay.io",
		"Pause":          "gcr.io",
		"Alpine":         "docker.io",
	}),
	"freshpod": NewAddon([]*BinAsset{
		MustBinAsset(addons.FreshpodAssets,
			"freshpod/freshpod-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"freshpod-rc.yaml",
			"0640"),
	}, false, "freshpod", "Google", "", "https://github.com/GoogleCloudPlatform/freshpod", map[string]string{
		"FreshPod": "google-samples/freshpod:v0.0.1@sha256:b9efde5b509da3fd2959519c4147b653d0c5cefe8a00314e2888e35ecbcb46f9",
	}, map[string]string{
		"FreshPod": "gcr.io",
	}),
	"nvidia-driver-installer": NewAddon([]*BinAsset{
		MustBinAsset(addons.NvidiaDriverInstallerAssets,
			"gpu/nvidia-driver-installer.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"nvidia-driver-installer.yaml",
			"0640"),
	}, false, "nvidia-driver-installer", "Google", "", "https://minikube.sigs.k8s.io/docs/tutorials/nvidia_gpu/", map[string]string{
		"NvidiaDriverInstaller": "minikube-nvidia-driver-installer:e2d9b43228decf5d6f7dce3f0a85d390f138fa01",
		"Pause":                 "pause:2.0@sha256:9ce5316f9752b8347484ab0f6778573af15524124d52b93230b9a0dcc987e73e",
	}, map[string]string{
		"NvidiaDriverInstaller": "k8s.gcr.io",
		"Pause":                 "k8s.gcr.io",
	}),
	"nvidia-gpu-device-plugin": NewAddon([]*BinAsset{
		MustBinAsset(addons.NvidiaGpuDevicePluginAssets,
			"gpu/nvidia-gpu-device-plugin.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"nvidia-gpu-device-plugin.yaml",
			"0640"),
	}, false, "nvidia-gpu-device-plugin", "3rd party (Nvidia)", "", "https://minikube.sigs.k8s.io/docs/tutorials/nvidia_gpu/", map[string]string{
		"NvidiaDevicePlugin": "nvidia-gpu-device-plugin@sha256:4b036e8844920336fa48f36edeb7d4398f426d6a934ba022848deed2edbf09aa",
	}, map[string]string{
		"NvidiaDevicePlugin": "k8s.gcr.io",
	}),
	"logviewer": NewAddon([]*BinAsset{
		MustBinAsset(addons.LogviewerAssets,
			"logviewer/logviewer-dp-and-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"logviewer-dp-and-svc.yaml",
			"0640"),
		MustBinAsset(addons.LogviewerAssets,
			"logviewer/logviewer-rbac.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"logviewer-rbac.yaml",
			"0640"),
	}, false, "logviewer", "3rd party (unknown)", "", "", map[string]string{
		"LogViewer": "ivans3/minikube-log-viewer:latest@sha256:75854f45305cc47d17b04c6c588fa60777391761f951e3a34161ddf1f1b06405",
	}, map[string]string{
		"LogViewer": "docker.io",
	}),
	"gvisor": NewAddon([]*BinAsset{
		MustBinAsset(addons.GvisorAssets,
			"gvisor/gvisor-pod.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gvisor-pod.yaml",
			"0640"),
		MustBinAsset(addons.GvisorAssets,
			"gvisor/gvisor-runtimeclass.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gvisor-runtimeclass.yaml",
			"0640"),
		MustBinAsset(addons.GvisorAssets,
			"gvisor/gvisor-config.toml",
			vmpath.GuestGvisorDir,
			constants.GvisorConfigTomlTargetName,
			"0640"),
	}, false, "gvisor", "Google", "", "https://github.com/kubernetes/minikube/blob/master/deploy/addons/gvisor/README.md", map[string]string{
		"GvisorAddon": "k8s-minikube/gvisor-addon:3@sha256:23eb17d48a66fc2b09c31454fb54ecae520c3e9c9197ef17fcb398b4f31d505a",
	}, map[string]string{
		"GvisorAddon": "gcr.io",
	}),
	"helm-tiller": NewAddon([]*BinAsset{
		MustBinAsset(addons.HelmTillerAssets,
			"helm-tiller/helm-tiller-dp.tmpl",
			vmpath.GuestAddonsDir,
			"helm-tiller-dp.yaml",
			"0640"),
		MustBinAsset(addons.HelmTillerAssets,
			"helm-tiller/helm-tiller-rbac.tmpl",
			vmpath.GuestAddonsDir,
			"helm-tiller-rbac.yaml",
			"0640"),
		MustBinAsset(addons.HelmTillerAssets,
			"helm-tiller/helm-tiller-svc.tmpl",
			vmpath.GuestAddonsDir,
			"helm-tiller-svc.yaml",
			"0640"),
	}, false, "helm-tiller", "3rd party (Helm)", "", "https://v2.helm.sh/docs/using_helm/", map[string]string{
		"Tiller": "helm/tiller:v2.17.0@sha256:4c43eb385032945cad047d2350e4945d913b90b3ab43ee61cecb32a495c6df0f",
	}, map[string]string{
		// GCR is deprecated in helm
		// https://github.com/helm/helm/issues/10004#issuecomment-894478908
		"Tiller": "ghcr.io",
	}),
	"ingress-dns": NewAddon([]*BinAsset{
		MustBinAsset(addons.IngressDNSAssets,
			"ingress-dns/ingress-dns-pod.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ingress-dns-pod.yaml",
			"0640"),
	}, false, "ingress-dns", "Google", "", "https://minikube.sigs.k8s.io/docs/handbook/addons/ingress-dns/", map[string]string{
		"IngressDNS": "k8s-minikube/minikube-ingress-dns:0.0.2@sha256:4abe27f9fc03fedab1d655e2020e6b165faf3bf6de1088ce6cf215a75b78f05f",
	}, map[string]string{
		"IngressDNS": "gcr.io",
	}),
	"metallb": NewAddon([]*BinAsset{
		MustBinAsset(addons.MetallbAssets,
			"metallb/metallb.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metallb.yaml",
			"0640"),
		MustBinAsset(addons.MetallbAssets,
			"metallb/metallb-config.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metallb-config.yaml",
			"0640"),
	}, false, "metallb", "3rd party (MetalLB)", "", "", map[string]string{
		"Speaker":    "metallb/speaker:v0.9.6@sha256:c66585a805bed1a3b829d8fb4a4aab9d87233497244ebff96f1b88f1e7f8f991",
		"Controller": "metallb/controller:v0.9.6@sha256:fbfdb9d3f55976b0ee38f3309d83a4ca703efcf15d6ca7889cd8189142286502",
	}, map[string]string{
		"Speaker":    "docker.io",
		"Controller": "docker.io",
	}),
	"ambassador": NewAddon([]*BinAsset{
		MustBinAsset(addons.AmbassadorAssets,
			"ambassador/ambassador-operator-crds.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ambassador-operator-crds.yaml",
			"0640"),
		MustBinAsset(addons.AmbassadorAssets,
			"ambassador/ambassador-operator.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ambassador-operator.yaml",
			"0640"),
		MustBinAsset(addons.AmbassadorAssets,
			"ambassador/ambassadorinstallation.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ambassadorinstallation.yaml",
			"0640"),
	}, false, "ambassador", "3rd party (Ambassador)", "", "https://minikube.sigs.k8s.io/docs/tutorials/ambassador_ingress_controller/", map[string]string{
		"AmbassadorOperator": "datawire/ambassador-operator:v1.2.3@sha256:492f33e0828a371aa23331d75c11c251b21499e31287f026269e3f6ec6da34ed",
	}, map[string]string{
		"AmbassadorOperator": "quay.io",
	}),
	"gcp-auth": NewAddon([]*BinAsset{
		MustBinAsset(addons.GcpAuthAssets,
			"gcp-auth/gcp-auth-ns.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gcp-auth-ns.yaml",
			"0640"),
		MustBinAsset(addons.GcpAuthAssets,
			"gcp-auth/gcp-auth-service.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gcp-auth-service.yaml",
			"0640"),
		MustBinAsset(addons.GcpAuthAssets,
			"gcp-auth/gcp-auth-webhook.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gcp-auth-webhook.yaml",
			"0640"),
	}, false, "gcp-auth", "Google", "", "https://minikube.sigs.k8s.io/docs/handbook/addons/gcp-auth/", map[string]string{
		"KubeWebhookCertgen": "ingress-nginx/kube-webhook-certgen:v1.0@sha256:f3b6b39a6062328c095337b4cadcefd1612348fdd5190b1dcbcb9b9e90bd8068",
		"GCPAuthWebhook":     "k8s-minikube/gcp-auth-webhook:v0.0.13@sha256:08a49cb7a588d81723b7e02c16082c75418b6e0a54cf2e44668bd77f79a41a40",
	}, map[string]string{
		"GCPAuthWebhook":     "gcr.io",
		"KubeWebhookCertgen": "k8s.gcr.io",
	}),
	"volumesnapshots": NewAddon([]*BinAsset{
		// make sure the order of apply. `csi-hostpath-snapshotclass` must be the first position, because it depends on `snapshot.storage.k8s.io_volumesnapshotclasses`
		// if user disable volumesnapshots addon and delete `csi-hostpath-snapshotclass` after `snapshot.storage.k8s.io_volumesnapshotclasses`, kubernetes will return the error
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/csi-hostpath-snapshotclass.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-snapshotclass.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/snapshot.storage.k8s.io_volumesnapshotclasses.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshotclasses.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/snapshot.storage.k8s.io_volumesnapshotcontents.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshotcontents.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/snapshot.storage.k8s.io_volumesnapshots.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshots.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/rbac-volume-snapshot-controller.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-volume-snapshot-controller.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/volume-snapshot-controller-deployment.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"volume-snapshot-controller-deployment.yaml",
			"0640"),
	}, false, "volumesnapshots", "Kubernetes", "", "https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/", map[string]string{
		"SnapshotController": "sig-storage/snapshot-controller:v4.0.0@sha256:00fcc441ea9f72899c25eed61d602272a2a58c5f0014332bdcb5ac24acef08e4",
	}, map[string]string{
		"SnapshotController": "k8s.gcr.io",
	}),
	"csi-hostpath-driver": NewAddon([]*BinAsset{
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-attacher.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-attacher.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-health-monitor-agent.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-health-monitor-agent.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-health-monitor-controller.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-health-monitor-controller.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-provisioner.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-provisioner.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-resizer.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-resizer.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-snapshotter.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-snapshotter.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-attacher.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-attacher.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-driverinfo.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-driverinfo.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-plugin.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-plugin.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-provisioner.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-provisioner.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-resizer.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-resizer.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-snapshotter.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-snapshotter.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-storageclass.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-storageclass.yaml",
			"0640"),
	}, false, "csi-hostpath-driver", "Kubernetes", "", "https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/", map[string]string{
		"Attacher":              "sig-storage/csi-attacher:v3.1.0@sha256:50c3cfd458fc8e0bf3c8c521eac39172009382fc66dc5044a330d137c6ed0b09",
		"HostMonitorAgent":      "sig-storage/csi-external-health-monitor-agent:v0.2.0@sha256:c20d4a4772599e68944452edfcecc944a1df28c19e94b942d526ca25a522ea02",
		"HostMonitorController": "sig-storage/csi-external-health-monitor-controller:v0.2.0@sha256:14988b598a180cc0282f3f4bc982371baf9a9c9b80878fb385f8ae8bd04ecf16",
		"NodeDriverRegistrar":   "sig-storage/csi-node-driver-registrar:v2.0.1@sha256:e07f914c32f0505e4c470a62a40ee43f84cbf8dc46ff861f31b14457ccbad108",
		"HostPathPlugin":        "sig-storage/hostpathplugin:v1.6.0@sha256:b526bd29630261eceecf2d38c84d4f340a424d57e1e2661111e2649a4663b659",
		"LivenessProbe":         "sig-storage/livenessprobe:v2.2.0@sha256:48da0e4ed7238ad461ea05f68c25921783c37b315f21a5c5a2780157a6460994",
		"Resizer":               "sig-storage/csi-resizer:v1.1.0@sha256:7a5ba58a44e0d749e0767e4e37315bcf6a61f33ce3185c1991848af4db0fb70a",
		"Snapshotter":           "sig-storage/csi-snapshotter:v4.0.0@sha256:51f2dfde5bccac7854b3704689506aeecfb793328427b91115ba253a93e60782",
		"Provisioner":           "sig-storage/csi-provisioner:v2.1.0@sha256:20c828075d1e36f679d6a91e905b0927141eef5e15be0c9a1ca4a6a0ed9313d2",
	}, map[string]string{
		"Attacher":              "k8s.gcr.io",
		"HostMonitorAgent":      "k8s.gcr.io",
		"HostMonitorController": "k8s.gcr.io",
		"NodeDriverRegistrar":   "k8s.gcr.io",
		"HostPathPlugin":        "k8s.gcr.io",
		"LivenessProbe":         "k8s.gcr.io",
		"Resizer":               "k8s.gcr.io",
		"Snapshotter":           "k8s.gcr.io",
		"Provisioner":           "k8s.gcr.io",
	}),
	"portainer": NewAddon([]*BinAsset{
		MustBinAsset(addons.PortainerAssets,
			"portainer/portainer.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"portainer.yaml",
			"0640"),
	}, false, "portainer", "3rd party (Portainer.io)", "", "", map[string]string{
		"Portainer": "portainer/portainer-ce:2.15.1@sha256:5466af30b8eaf3f75edd3c74703d1c9973f0963acd6ef164913ea6f195d640c2",
	}, map[string]string{
		"Portainer": "docker.io",
	}),
	"inaccel": NewAddon([]*BinAsset{
		MustBinAsset(addons.InAccelAssets,
			"inaccel/fpga-operator.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"fpga-operator.yaml",
			"0640"),
	}, false, "inaccel", "3rd party (InAccel [info@inaccel.com])", "", "", map[string]string{
		"Helm3": "alpine/helm:3.9.0@sha256:9f4bf4d24241f983910550b1fe8688571cd684046500abe58cef14308f9cb19e",
	}, map[string]string{
		"Helm3": "docker.io",
	}),
	"headlamp": NewAddon([]*BinAsset{
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-namespace.yaml", vmpath.GuestAddonsDir, "headlamp-namespace.yaml", "6040"),
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-service.yaml", vmpath.GuestAddonsDir, "headlamp-service.yaml", "6040"),
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-deployment.yaml.tmpl", vmpath.GuestAddonsDir, "headlamp-deployment.yaml", "6040"),
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-serviceaccount.yaml", vmpath.GuestAddonsDir, "headlamp-serviceaccount.yaml", "6040"),
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-clusterrolebinding.yaml", vmpath.GuestAddonsDir, "headlamp-clusterrolebinding.yaml", "6040"),
	}, false, "headlamp", "3rd party (kinvolk.io)", "", "https://minikube.sigs.k8s.io/docs/handbook/addons/headlamp/",
		map[string]string{
			"Headlamp": "kinvolk/headlamp:v0.14.1@sha256:162f6888461362125f4f8c60cd0a9c4031a81de15f1ed4c67bd3600b48b24029",
		},
		map[string]string{
			"Headlamp": "ghcr.io",
		}),
	"cloud-spanner": NewAddon([]*BinAsset{
		MustBinAsset(addons.CloudSpanner, "cloud-spanner/deployment.yaml", vmpath.GuestAddonsDir, "deployment.yaml", "6040"),
	}, false, "cloud-spanner", "Google", "", "https://minikube.sigs.k8s.io/docs/handbook/addons/cloud-spanner/", map[string]string{
		"CloudSpanner": "cloud-spanner-emulator/emulator:1.4.9@sha256:ae507799872ab29109e17b4e8a8efbf57bf7ce227017f88516c9311bc83c4b51",
	}, map[string]string{
		"CloudSpanner": "gcr.io",
	}),
}

// parseMapString creates a map based on `str` which is encoded as <key1>=<value1>,<key2>=<value2>,...
func parseMapString(str string) map[string]string {
	mapResult := make(map[string]string)
	if str == "" {
		return mapResult
	}
	for _, pairText := range strings.Split(str, ",") {
		vals := strings.Split(pairText, "=")
		if len(vals) != 2 {
			out.WarningT("Ignoring invalid pair entry {{.pair}}", out.V{"pair": pairText})
			continue
		}
		mapResult[vals[0]] = vals[1]
	}
	return mapResult
}

// mergeMaps returns a map with the union of `source` and `override` where collisions take the value of `override`.
func mergeMaps(source, override map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range source {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// filterKeySpace creates a map of the values in `target` where the keys are also in `keySpace`.
func filterKeySpace(keySpace, target map[string]string) map[string]string {
	result := make(map[string]string)
	for name := range keySpace {
		if value, ok := target[name]; ok {
			result[name] = value
		}
	}
	return result
}

// overrideDefaults creates a copy of `def` where `override` replaces any of its values that `override` contains.
func overrideDefaults(def, override map[string]string) map[string]string {
	return mergeMaps(def, filterKeySpace(def, override))
}

// SelectAndPersistImages selects which images to use based on addon default images, previously persisted images, and newly requested images - which are then persisted for future enables.
func SelectAndPersistImages(addon *Addon, cc *config.ClusterConfig) (images, customRegistries map[string]string, _ error) {
	addonDefaultImages := addon.Images
	if addonDefaultImages == nil {
		addonDefaultImages = make(map[string]string)
	}

	// Use previously configured custom images.
	images = overrideDefaults(addonDefaultImages, cc.CustomAddonImages)
	if viper.IsSet(config.AddonImages) {
		// Parse the AddonImages flag if present.
		newImages := parseMapString(viper.GetString(config.AddonImages))
		for name, image := range newImages {
			if image == "" {
				out.WarningT("Ignoring empty custom image {{.name}}", out.V{"name": name})
				delete(newImages, name)
				continue
			}
			if _, ok := addonDefaultImages[name]; !ok {
				out.WarningT("Ignoring unknown custom image {{.name}}", out.V{"name": name})
			}
		}
		// Use newly configured custom images.
		images = overrideDefaults(addonDefaultImages, newImages)
		// Store custom addon images to be written.
		cc.CustomAddonImages = mergeMaps(cc.CustomAddonImages, newImages)
	}

	// Use previously configured custom registries.
	customRegistries = filterKeySpace(addonDefaultImages, cc.CustomAddonRegistries) // filter by images map because registry map may omit default registry.
	if viper.IsSet(config.AddonRegistries) {
		// Parse the AddonRegistries flag if present.
		customRegistries = parseMapString(viper.GetString(config.AddonRegistries))
		for name := range customRegistries {
			if _, ok := addonDefaultImages[name]; !ok { // check images map because registry map may omitted default registry
				out.WarningT("Ignoring unknown custom registry {{.name}}", out.V{"name": name})
				delete(customRegistries, name)
			}
		}
		// Since registry map may omit default registry, any previously set custom registries for these images must be cleared.
		for name := range addonDefaultImages {
			delete(cc.CustomAddonRegistries, name)
		}
		// Merge newly set registries into custom addon registries to be written.
		cc.CustomAddonRegistries = mergeMaps(cc.CustomAddonRegistries, customRegistries)
	}

	// If images or registries were specified, save the config afterward.
	if viper.IsSet(config.AddonImages) || viper.IsSet(config.AddonRegistries) {
		// Since these values are only set when a user enables an addon, it is safe to refer to the profile name.
		// Whether err is nil or not we still return here.
		return images, customRegistries, config.Write(viper.GetString(config.ProfileName), cc)
	}
	return images, customRegistries, nil
}

// GenerateTemplateData generates template data for template assets
func GenerateTemplateData(addon *Addon, cc *config.ClusterConfig, netInfo NetworkInfo, images, customRegistries map[string]string, enable bool) interface{} {
	cfg := cc.KubernetesConfig
	a := runtime.GOARCH
	// Some legacy docker images still need the -arch suffix
	// for less common architectures blank suffix for amd64
	ea := ""
	if runtime.GOARCH != "amd64" {
		ea = "-" + runtime.GOARCH
	}

	v, err := util.ParseKubernetesVersion(cfg.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}

	opts := struct {
		KubernetesVersion       map[string]uint64
		PreOneTwentyKubernetes  bool
		Arch                    string
		ExoticArch              string
		ImageRepository         string
		LoadBalancerStartIP     string
		LoadBalancerEndIP       string
		CustomIngressCert       string
		IngressAPIVersion       string
		ContainerRuntime        string
		RegistryAliases         string
		Images                  map[string]string
		Registries              map[string]string
		CustomRegistries        map[string]string
		NetworkInfo             map[string]string
		Environment             map[string]string
		LegacyPodSecurityPolicy bool
		LegacyRuntimeClass      bool
	}{
		KubernetesVersion:      make(map[string]uint64),
		PreOneTwentyKubernetes: false,
		Arch:                   a,
		ExoticArch:             ea,
		ImageRepository:        cfg.ImageRepository,
		LoadBalancerStartIP:    cfg.LoadBalancerStartIP,
		LoadBalancerEndIP:      cfg.LoadBalancerEndIP,
		CustomIngressCert:      cfg.CustomIngressCert,
		RegistryAliases:        cfg.RegistryAliases,
		IngressAPIVersion:      "v1", // api version for ingress (eg, "v1beta1"; defaults to "v1" for k8s 1.19+)
		ContainerRuntime:       cfg.ContainerRuntime,
		Images:                 images,
		Registries:             addon.Registries,
		CustomRegistries:       customRegistries,
		NetworkInfo:            make(map[string]string),
		Environment: map[string]string{
			"MockGoogleToken": os.Getenv("MOCK_GOOGLE_TOKEN"),
		},
		LegacyPodSecurityPolicy: v.LT(semver.Version{Major: 1, Minor: 25}),
		LegacyRuntimeClass:      v.LT(semver.Version{Major: 1, Minor: 25}),
	}
	if opts.ImageRepository != "" && !strings.HasSuffix(opts.ImageRepository, "/") {
		opts.ImageRepository += "/"
	}
	if opts.Registries == nil {
		opts.Registries = make(map[string]string)
	}

	// maintain backwards compatibility with k8s < v1.19
	// by using v1beta1 instead of v1 api version for ingress
	if semver.MustParseRange("<1.19.0")(v) {
		opts.IngressAPIVersion = "v1beta1"
	}
	if semver.MustParseRange("<1.20.0")(v) {
		opts.PreOneTwentyKubernetes = true
	}

	// Network info for generating template
	opts.NetworkInfo["ControlPlaneNodeIP"] = netInfo.ControlPlaneNodeIP
	opts.NetworkInfo["ControlPlaneNodePort"] = fmt.Sprint(netInfo.ControlPlaneNodePort)

	// Append postfix "/" to registries
	for k, v := range opts.Registries {
		if v != "" && !strings.HasSuffix(v, "/") {
			opts.Registries[k] = v + "/"
		}
	}

	for k, v := range opts.CustomRegistries {
		if v != "" && !strings.HasSuffix(v, "/") {
			opts.CustomRegistries[k] = v + "/"
		}
	}

	for name, image := range opts.Images {
		if _, ok := opts.Registries[name]; !ok {
			opts.Registries[name] = "" // Avoid nil access when rendering
		}

		// tl;dr If the user specified a custom image remove the default registry
		// Without the line below, if you try to overwrite an image the default registry is still used in the templating
		// Example - image name: MetricsScraper, default registry: docker.io, default image: kubernetesui/metrics-scraper
		// Passed on addon enable: --images=MetricsScraper=k8s.gcr.io/echoserver:1.4
		// Without this line the resulting image would be docker.io/k8s.gcr.io/echoserver:1.4
		if _, ok := cc.CustomAddonImages[name]; ok {
			opts.Registries[name] = ""
		}

		if enable {
			if override, ok := opts.CustomRegistries[name]; ok {
				out.Infof("Using image {{.registry}}{{.image}}", out.V{
					"registry": override,
					// removing the SHA from UI
					// SHA example gcr.io/k8s-minikube/gcp-auth-webhook:v0.0.4@sha256:65e9e69022aa7b0eb1e390e1916e3bf67f75ae5c25987f9154ef3b0e8ab8528b
					"image": strings.Split(image, "@")[0],
				})
			} else if opts.ImageRepository != "" {
				out.Infof("Using image {{.registry}}{{.image}} (global image repository)", out.V{
					"registry": opts.ImageRepository,
					"image":    image,
				})
			} else {
				out.Infof("Using image {{.registry}}{{.image}}", out.V{
					"registry": opts.Registries[name],
					"image":    strings.Split(image, "@")[0],
				})
			}
		}
	}
	return opts
}
