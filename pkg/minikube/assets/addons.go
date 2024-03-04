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
	"time"

	semver "github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/deploy/addons"
	"k8s.io/minikube/pkg/minikube/config"
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

	return false
}

// IsEnabledOrDefault checks if an Addon is enabled for the given profile. If not found in profile it returns the default state
func (a *Addon) IsEnabledOrDefault(cc *config.ClusterConfig) bool {
	status, ok := cc.Addons[a.Name()]
	if ok {
		return status
	}

	// Return the default unconfigured state of the addon
	return a.enabled
}

// EnableByDefault will enable the addon by default on cluster start
func (a *Addon) EnableByDefault() {
	a.enabled = true
}

// Addons is the list of addons
// TODO: Make dynamically loadable: move this data to a .yaml file within each addon directory
var Addons = map[string]*Addon{
	"auto-pause": NewAddon([]*BinAsset{
		MustBinAsset(
			addons.AutoPauseAssets,
			"auto-pause/auto-pause.yaml",
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
	}, false, "auto-pause", "minikube", "", "", map[string]string{
		"AutoPauseHook": "k8s-minikube/auto-pause-hook:v0.0.5@sha256:d613ed2c891882b602b5aca668e92d4606a1b3832d96750ab25804de15929522",
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
			"storageclass/storageclass.yaml",
			vmpath.GuestAddonsDir,
			"storageclass.yaml",
			"0640"),
	}, true, "default-storageclass", "Kubernetes", "", "https://minikube.sigs.k8s.io/docs/handbook/persistent_volumes/", nil, nil),
	"pod-security-policy": NewAddon([]*BinAsset{
		MustBinAsset(addons.PodSecurityPolicyAssets,
			"pod-security-policy/pod-security-policy.yaml",
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
	}, true, "storage-provisioner", "minikube", "", "", map[string]string{
		"StorageProvisioner": fmt.Sprintf("k8s-minikube/storage-provisioner:%s", version.GetStorageProvisionerVersion()),
	}, map[string]string{
		"StorageProvisioner": "gcr.io",
	}),
	"storage-provisioner-gluster": NewAddon([]*BinAsset{
		MustBinAsset(addons.StorageProvisionerGlusterAssets,
			"storage-provisioner-gluster/storage-gluster-ns.yaml",
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
	"storage-provisioner-rancher": NewAddon([]*BinAsset{
		MustBinAsset(addons.StorageProvisionerRancherAssets,
			"storage-provisioner-rancher/storage-provisioner-rancher.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storage-provisioner-rancher.yaml",
			"0640"),
	}, false, "storage-provisioner-rancher", "3rd party (Rancher)", "", "", map[string]string{
		"LocalPathProvisioner": "rancher/local-path-provisioner:v0.0.22@sha256:e34c88ae0affb1cdefbb874140d6339d4a27ec4ee420ae8199cd839997b05246",
		"Helper":               "busybox:stable@sha256:3fbc632167424a6d997e74f52b878d7cc478225cffac6bc977eedfe51c7f4e79",
	}, map[string]string{
		"LocalPathProvisioner": "docker.io",
		"Helper":               "docker.io",
	}),
	"efk": NewAddon([]*BinAsset{
		MustBinAsset(addons.EfkAssets,
			"efk/elasticsearch-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"elasticsearch-rc.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/elasticsearch-svc.yaml",
			vmpath.GuestAddonsDir,
			"elasticsearch-svc.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/fluentd-es-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"fluentd-es-rc.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/fluentd-es-configmap.yaml",
			vmpath.GuestAddonsDir,
			"fluentd-es-configmap.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/kibana-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"kibana-rc.yaml",
			"0640"),
		MustBinAsset(addons.EfkAssets,
			"efk/kibana-svc.yaml",
			vmpath.GuestAddonsDir,
			"kibana-svc.yaml",
			"0640"),
	}, false, "efk", "3rd party (Elastic)", "", "", map[string]string{
		"Elasticsearch":        "elasticsearch:v5.6.2@sha256:7e95b32a7a2aad0c0db5c881e4a1ce8b7e53236144ae9d9cfb5fbe5608af4ab2",
		"FluentdElasticsearch": "fluentd-elasticsearch:v2.0.2@sha256:d0480bbf2d0de2344036fa3f7034cf7b4b98025a89c71d7f1f1845ac0e7d5a97",
		"Alpine":               "alpine:3.6@sha256:66790a2b79e1ea3e1dabac43990c54aca5d1ddf268d9a5a0285e4167c8b24475",
		"Kibana":               "kibana/kibana:5.6.2@sha256:cd948a9bda4622f1437afc4a3e78be6c8c25fc62f40aa0376f3d690f2436568f",
	}, map[string]string{
		"Elasticsearch":        "registry.k8s.io",
		"FluentdElasticsearch": "registry.k8s.io",
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
		// https://github.com/kubernetes/ingress-nginx/blob/3476232f5c38383dd157ddaff3b4c7cebd57284e/deploy/static/provider/kind/deploy.yaml#L445
		"IngressController": "ingress-nginx/controller:v1.10.0@sha256:42b3f0e5d0846876b1791cd3afeb5f1cbbe4259d6f35651dcc1b5c980925379c",
		// https://github.com/kubernetes/ingress-nginx/blob/3476232f5c38383dd157ddaff3b4c7cebd57284e/deploy/static/provider/kind/deploy.yaml#L552
		"KubeWebhookCertgenCreate": "ingress-nginx/kube-webhook-certgen:v1.4.0@sha256:44d1d0e9f19c63f58b380c5fddaca7cf22c7cee564adeff365225a5df5ef3334",
		// https://github.com/kubernetes/ingress-nginx/blob/3476232f5c38383dd157ddaff3b4c7cebd57284e/deploy/static/provider/kind/deploy.yaml#L601
		"KubeWebhookCertgenPatch": "ingress-nginx/kube-webhook-certgen:v1.4.0@sha256:44d1d0e9f19c63f58b380c5fddaca7cf22c7cee564adeff365225a5df5ef3334",
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
		"IstioOperator": "istio/operator:1.20.3@sha256:c0c04a762fa037c6e249ee08a33a740e115b6774f02b5b32e60a11727c632714",
	}, map[string]string{
		"IstioOperator": "docker.io",
	}),
	"istio": NewAddon([]*BinAsset{
		MustBinAsset(addons.IstioAssets,
			"istio/istio-default-profile.yaml",
			vmpath.GuestAddonsDir,
			"istio-default-profile.yaml",
			"0640"),
	}, false, "istio", "3rd party (Istio)", "", "https://istio.io/latest/docs/setup/platform-setup/minikube/", nil, nil),
	"inspektor-gadget": NewAddon([]*BinAsset{
		MustBinAsset(addons.InspektorGadgetAssets, "inspektor-gadget/ig-namespace.yaml", vmpath.GuestAddonsDir, "ig-namespace.yaml", "0640"),
		MustBinAsset(addons.InspektorGadgetAssets, "inspektor-gadget/ig-serviceaccount.yaml", vmpath.GuestAddonsDir, "ig-serviceaccount.yaml", "0640"),
		MustBinAsset(addons.InspektorGadgetAssets, "inspektor-gadget/ig-role.yaml", vmpath.GuestAddonsDir, "ig-role.yaml", "0640"),
		MustBinAsset(addons.InspektorGadgetAssets, "inspektor-gadget/ig-rolebinding.yaml", vmpath.GuestAddonsDir, "ig-rolebinding.yaml", "0640"),
		MustBinAsset(addons.InspektorGadgetAssets, "inspektor-gadget/ig-clusterrole.yaml", vmpath.GuestAddonsDir, "ig-clusterrole.yaml", "0640"),
		MustBinAsset(addons.InspektorGadgetAssets, "inspektor-gadget/ig-clusterrolebinding.yaml", vmpath.GuestAddonsDir, "ig-clusterrolebinding.yaml", "0640"),
		MustBinAsset(addons.InspektorGadgetAssets, "inspektor-gadget/ig-crd.yaml", vmpath.GuestAddonsDir, "ig-crd.yaml", "0640"),
		MustBinAsset(addons.InspektorGadgetAssets, "inspektor-gadget/ig-daemonset.yaml.tmpl", vmpath.GuestAddonsDir, "ig-daemonset.yaml", "0640"),
	}, false, "inspektor-gadget", "3rd party (inspektor-gadget.io)", "https://github.com/orgs/inspektor-gadget/people", "https://minikube.sigs.k8s.io/docs/handbook/addons/inspektor-gadget/",
		map[string]string{
			"InspektorGadget": "inspektor-gadget/inspektor-gadget:v0.25.1@sha256:efddd4f0a8b51a7c406c67894203bc475198f54809105ce0c2df904a44180e75",
		}, map[string]string{
			"InspektorGadget": "ghcr.io",
		}),
	"kong": NewAddon([]*BinAsset{
		MustBinAsset(addons.KongAssets,
			"kong/kong-ingress-controller.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"kong-ingress-controller.yaml",
			"0640"),
	}, false, "kong", "3rd party (Kong HQ)", "@gAmUssA", "https://minikube.sigs.k8s.io/docs/handbook/addons/kong-ingress/", map[string]string{
		"Kong":        "kong:3.6.0@sha256:2f2832a751eee25b4cb643a9b8b14df602a88c249c2baba46e3e7b51df2bb78b",
		"KongIngress": "kong/kubernetes-ingress-controller:2.12.0@sha256:ef450cd85f4237aa5acf3bfaebd064accdc6a445419bec20a5a68067c55c6809",
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
		"Kubectl": "bitnami/kubectl:1.29.2@sha256:0fdc1ab053addba12b63c232f7085ea1b3103d3bfea06c6210069e22b10d9555",
	}, map[string]string{
		"Kubectl": "docker.io",
	}),
	"metrics-server": NewAddon([]*BinAsset{
		MustBinAsset(addons.MetricsServerAssets,
			"metrics-server/metrics-apiservice.yaml",
			vmpath.GuestAddonsDir,
			"metrics-apiservice.yaml",
			"0640"),
		MustBinAsset(addons.MetricsServerAssets,
			"metrics-server/metrics-server-deployment.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metrics-server-deployment.yaml",
			"0640"),
		MustBinAsset(addons.MetricsServerAssets,
			"metrics-server/metrics-server-rbac.yaml",
			vmpath.GuestAddonsDir,
			"metrics-server-rbac.yaml",
			"0640"),
		MustBinAsset(addons.MetricsServerAssets,
			"metrics-server/metrics-server-service.yaml",
			vmpath.GuestAddonsDir,
			"metrics-server-service.yaml",
			"0640"),
	}, false, "metrics-server", "Kubernetes", "", "", map[string]string{
		"MetricsServer": "metrics-server/metrics-server:v0.7.0@sha256:1c0419326500f1704af580d12a579671b2c3a06a8aa918cd61d0a35fb2d6b3ce",
	}, map[string]string{
		"MetricsServer": "registry.k8s.io",
	}),
	"olm": NewAddon([]*BinAsset{
		MustBinAsset(addons.OlmAssets,
			"olm/crds.yaml",
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
			"registry/registry-svc.yaml",
			vmpath.GuestAddonsDir,
			"registry-svc.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAssets,
			"registry/registry-proxy.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-proxy.yaml",
			"0640"),
	}, false, "registry", "minikube", "", "", map[string]string{
		"Registry":          "registry:2.8.3@sha256:f4e1b878d4bc40a1f65532d68c94dcfbab56aa8cba1f00e355a206e7f6cc9111",
		"KubeRegistryProxy": "k8s-minikube/kube-registry-proxy:0.0.5@sha256:f107ecd58728a2df5f2bb7e087f65f5363d0019b1e1fd476e4ef16065f44abfb",
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
			"registry-aliases/registry-aliases-sa.yaml",
			vmpath.GuestAddonsDir,
			"registry-aliases-sa.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/registry-aliases-sa-crb.yaml",
			vmpath.GuestAddonsDir,
			"registry-aliases-sa-crb.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/registry-aliases-config.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-aliases-config.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/node-etc-hosts-update.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"node-etc-hosts-update.yaml",
			"0640"),
		MustBinAsset(addons.RegistryAliasesAssets,
			"registry-aliases/patch-coredns-job.yaml.tmpl",
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
	}, false, "nvidia-driver-installer", "3rd party (Nvidia)", "", "https://minikube.sigs.k8s.io/docs/tutorials/nvidia_gpu/", map[string]string{
		"NvidiaDriverInstaller": "minikube-nvidia-driver-installer:e2d9b43228decf5d6f7dce3f0a85d390f138fa01",
		"Pause":                 "pause:2.0@sha256:9ce5316f9752b8347484ab0f6778573af15524124d52b93230b9a0dcc987e73e",
	}, map[string]string{
		"NvidiaDriverInstaller": "registry.k8s.io",
		"Pause":                 "registry.k8s.io",
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
		"NvidiaDevicePlugin": "registry.k8s.io",
	}),
	"logviewer": NewAddon([]*BinAsset{
		MustBinAsset(addons.LogviewerAssets,
			"logviewer/logviewer-dp-and-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"logviewer-dp-and-svc.yaml",
			"0640"),
		MustBinAsset(addons.LogviewerAssets,
			"logviewer/logviewer-rbac.yaml",
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
	}, false, "gvisor", "minikube", "", "https://github.com/kubernetes/minikube/blob/master/deploy/addons/gvisor/README.md", map[string]string{
		"GvisorAddon": "k8s-minikube/gvisor-addon:4@sha256:4bdc0bec3f36a32e534d9da98552810c832dd58fd0a361e5c0b836606b164bc2",
	}, map[string]string{
		"GvisorAddon": "gcr.io",
	}),
	"helm-tiller": NewAddon([]*BinAsset{
		MustBinAsset(addons.HelmTillerAssets,
			"helm-tiller/helm-tiller-dp.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"helm-tiller-dp.yaml",
			"0640"),
		MustBinAsset(addons.HelmTillerAssets,
			"helm-tiller/helm-tiller-rbac.yaml",
			vmpath.GuestAddonsDir,
			"helm-tiller-rbac.yaml",
			"0640"),
		MustBinAsset(addons.HelmTillerAssets,
			"helm-tiller/helm-tiller-svc.yaml",
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
	}, false, "ingress-dns", "minikube", "", "https://minikube.sigs.k8s.io/docs/handbook/addons/ingress-dns/", map[string]string{
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
		"Speaker":    "metallb/speaker:v0.9.6@sha256:7a400205b4986acd3d2ff32c29929682b8ff8d830837aff74f787c757176fa9f",
		"Controller": "metallb/controller:v0.9.6@sha256:6932cf255dd7f06f550c7f106b9a206be95f847ab8cb77aafac7acd27def0b00",
	}, map[string]string{
		"Speaker":    "quay.io",
		"Controller": "quay.io",
	}),
	"ambassador": NewAddon([]*BinAsset{
		MustBinAsset(addons.AmbassadorAssets,
			"ambassador/ambassador-operator-crds.yaml",
			vmpath.GuestAddonsDir,
			"ambassador-operator-crds.yaml",
			"0640"),
		MustBinAsset(addons.AmbassadorAssets,
			"ambassador/ambassador-operator.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ambassador-operator.yaml",
			"0640"),
		MustBinAsset(addons.AmbassadorAssets,
			"ambassador/ambassadorinstallation.yaml",
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
			"gcp-auth/gcp-auth-ns.yaml",
			vmpath.GuestAddonsDir,
			"gcp-auth-ns.yaml",
			"0640"),
		MustBinAsset(addons.GcpAuthAssets,
			"gcp-auth/gcp-auth-service.yaml",
			vmpath.GuestAddonsDir,
			"gcp-auth-service.yaml",
			"0640"),
		MustBinAsset(addons.GcpAuthAssets,
			"gcp-auth/gcp-auth-webhook.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gcp-auth-webhook.yaml",
			"0640"),
	}, false, "gcp-auth", "Google", "", "https://minikube.sigs.k8s.io/docs/handbook/addons/gcp-auth/", map[string]string{
		"KubeWebhookCertgen": "ingress-nginx/kube-webhook-certgen:v1.4.0@sha256:44d1d0e9f19c63f58b380c5fddaca7cf22c7cee564adeff365225a5df5ef3334",
		"GCPAuthWebhook":     "k8s-minikube/gcp-auth-webhook:v0.1.1@sha256:01b0de782aa30e7fc91ac5a91b5cc35e95e9679dee7ef07af06457b471f88f32",
	}, map[string]string{
		"GCPAuthWebhook":     "gcr.io",
		"KubeWebhookCertgen": "registry.k8s.io",
	}),
	"volumesnapshots": NewAddon([]*BinAsset{
		// make sure the order of apply. `csi-hostpath-snapshotclass` must be the first position, because it depends on `snapshot.storage.k8s.io_volumesnapshotclasses`
		// if user disable volumesnapshots addon and delete `csi-hostpath-snapshotclass` after `snapshot.storage.k8s.io_volumesnapshotclasses`, kubernetes will return the error
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/csi-hostpath-snapshotclass.yaml",
			vmpath.GuestAddonsDir,
			"csi-hostpath-snapshotclass.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/snapshot.storage.k8s.io_volumesnapshotclasses.yaml",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshotclasses.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/snapshot.storage.k8s.io_volumesnapshotcontents.yaml",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshotcontents.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/snapshot.storage.k8s.io_volumesnapshots.yaml",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshots.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/rbac-volume-snapshot-controller.yaml",
			vmpath.GuestAddonsDir,
			"rbac-volume-snapshot-controller.yaml",
			"0640"),
		MustBinAsset(addons.VolumeSnapshotsAssets,
			"volumesnapshots/volume-snapshot-controller-deployment.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"volume-snapshot-controller-deployment.yaml",
			"0640"),
	}, false, "volumesnapshots", "Kubernetes", "", "https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/", map[string]string{
		"SnapshotController": "sig-storage/snapshot-controller:v6.1.0@sha256:823c75d0c45d1427f6d850070956d9ca657140a7bbf828381541d1d808475280",
	}, map[string]string{
		"SnapshotController": "registry.k8s.io",
	}),
	"csi-hostpath-driver": NewAddon([]*BinAsset{
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-attacher.yaml",
			vmpath.GuestAddonsDir,
			"rbac-external-attacher.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-hostpath.yaml",
			vmpath.GuestAddonsDir,
			"rbac-hostpath.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-health-monitor-controller.yaml",
			vmpath.GuestAddonsDir,
			"rbac-external-health-monitor-controller.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-provisioner.yaml",
			vmpath.GuestAddonsDir,
			"rbac-external-provisioner.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-resizer.yaml",
			vmpath.GuestAddonsDir,
			"rbac-external-resizer.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/rbac/rbac-external-snapshotter.yaml",
			vmpath.GuestAddonsDir,
			"rbac-external-snapshotter.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-attacher.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-attacher.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-driverinfo.yaml",
			vmpath.GuestAddonsDir,
			"csi-hostpath-driverinfo.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-plugin.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-plugin.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-resizer.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-resizer.yaml",
			"0640"),
		MustBinAsset(addons.CsiHostpathDriverAssets,
			"csi-hostpath-driver/deploy/csi-hostpath-storageclass.yaml",
			vmpath.GuestAddonsDir,
			"csi-hostpath-storageclass.yaml",
			"0640"),
	}, false, "csi-hostpath-driver", "Kubernetes", "", "https://minikube.sigs.k8s.io/docs/tutorials/volume_snapshots_and_csi/", map[string]string{
		"Attacher":              "sig-storage/csi-attacher:v4.0.0@sha256:9a685020911e2725ad019dbce6e4a5ab93d51e3d4557f115e64343345e05781b",
		"HostMonitorController": "sig-storage/csi-external-health-monitor-controller:v0.7.0@sha256:80b9ba94aa2afe24553d69bd165a6a51552d1582d68618ec00d3b804a7d9193c",
		"NodeDriverRegistrar":   "sig-storage/csi-node-driver-registrar:v2.6.0@sha256:f1c25991bac2fbb7f5fcf91ed9438df31e30edee6bed5a780464238aa09ad24c",
		"HostPathPlugin":        "sig-storage/hostpathplugin:v1.9.0@sha256:92257881c1d6493cf18299a24af42330f891166560047902b8d431fb66b01af5",
		"LivenessProbe":         "sig-storage/livenessprobe:v2.8.0@sha256:cacee2b5c36dd59d4c7e8469c05c9e4ef53ecb2df9025fa8c10cdaf61bce62f0",
		"Resizer":               "sig-storage/csi-resizer:v1.6.0@sha256:425d8f1b769398127767b06ed97ce62578a3179bcb99809ce93a1649e025ffe7",
		"Snapshotter":           "sig-storage/csi-snapshotter:v6.1.0@sha256:291334908ddf71a4661fd7f6d9d97274de8a5378a2b6fdfeb2ce73414a34f82f",
		"Provisioner":           "sig-storage/csi-provisioner:v3.3.0@sha256:ee3b525d5b89db99da3b8eb521d9cd90cb6e9ef0fbb651e98bb37be78d36b5b8",
	}, map[string]string{
		"Attacher":              "registry.k8s.io",
		"HostMonitorController": "registry.k8s.io",
		"NodeDriverRegistrar":   "registry.k8s.io",
		"HostPathPlugin":        "registry.k8s.io",
		"LivenessProbe":         "registry.k8s.io",
		"Resizer":               "registry.k8s.io",
		"Snapshotter":           "registry.k8s.io",
		"Provisioner":           "registry.k8s.io",
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
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-namespace.yaml", vmpath.GuestAddonsDir, "headlamp-namespace.yaml", "0640"),
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-service.yaml", vmpath.GuestAddonsDir, "headlamp-service.yaml", "0640"),
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-deployment.yaml.tmpl", vmpath.GuestAddonsDir, "headlamp-deployment.yaml", "0640"),
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-serviceaccount.yaml", vmpath.GuestAddonsDir, "headlamp-serviceaccount.yaml", "0640"),
		MustBinAsset(addons.HeadlampAssets, "headlamp/headlamp-clusterrolebinding.yaml", vmpath.GuestAddonsDir, "headlamp-clusterrolebinding.yaml", "0640"),
	}, false, "headlamp", "3rd party (kinvolk.io)", "yolossn", "https://minikube.sigs.k8s.io/docs/handbook/addons/headlamp/",
		map[string]string{
			"Headlamp": "headlamp-k8s/headlamp:v0.22.0@sha256:3c6da859a989f285b2fd2ac2f4763d1884d54a51e4405301e5324e0b2b70bd67",
		},
		map[string]string{
			"Headlamp": "ghcr.io",
		}),
	"cloud-spanner": NewAddon([]*BinAsset{
		MustBinAsset(addons.CloudSpanner, "cloud-spanner/deployment.yaml.tmpl", vmpath.GuestAddonsDir, "deployment.yaml", "0640"),
	}, false, "cloud-spanner", "Google", "", "https://minikube.sigs.k8s.io/docs/handbook/addons/cloud-spanner/", map[string]string{
		"CloudSpanner": "cloud-spanner-emulator/emulator:1.5.14@sha256:41d5dccfcf13817a2348beba0ca7c650ffdd795f7fcbe975b7822c9eed262e15",
	}, map[string]string{
		"CloudSpanner": "gcr.io",
	}),
	"kubeflow": NewAddon([]*BinAsset{
		MustBinAsset(addons.Kubeflow, "kubeflow/kubeflow.yaml", vmpath.GuestAddonsDir, "kubeflow.yaml", "0640"),
	}, false, "kubeflow", "3rd party", "", "", nil, nil,
	),
	"nvidia-device-plugin": NewAddon([]*BinAsset{
		MustBinAsset(addons.NvidiaDevicePlugin, "nvidia-device-plugin/nvidia-device-plugin.yaml.tmpl", vmpath.GuestAddonsDir, "nvidia-device-plugin.yaml", "0640"),
	}, false, "nvidia-device-plugin", "3rd party (NVIDIA)", "", "",
		map[string]string{
			"NvidiaDevicePlugin": "nvidia/k8s-device-plugin:v0.14.5@sha256:50aa9517d771e3b0ffa7fded8f1e988dba680a7ff5efce162ce31d1b5ec043e2",
		}, map[string]string{
			"NvidiaDevicePlugin": "nvcr.io",
		}),
	"yakd": NewAddon([]*BinAsset{
		MustBinAsset(addons.YakdAssets, "yakd/yakd-ns.yaml", vmpath.GuestAddonsDir, "yakd-ns.yaml", "0640"),
		MustBinAsset(addons.YakdAssets, "yakd/yakd-sa.yaml", vmpath.GuestAddonsDir, "yakd-sa.yaml", "0640"),
		MustBinAsset(addons.YakdAssets, "yakd/yakd-crb.yaml", vmpath.GuestAddonsDir, "yakd-crb.yaml", "0640"),
		MustBinAsset(addons.YakdAssets, "yakd/yakd-svc.yaml", vmpath.GuestAddonsDir, "yakd-svc.yaml", "0640"),
		MustBinAsset(addons.YakdAssets, "yakd/yakd-dp.yaml.tmpl", vmpath.GuestAddonsDir, "yakd-dp.yaml", "0640"),
	}, false, "yakd", "3rd party (marcnuri.com)", "manusa", "https://minikube.sigs.k8s.io/docs/handbook/addons/yakd/",
		map[string]string{
			"Yakd": "marcnuri/yakd:0.0.4@sha256:a3f540278e4c11373e15605311851dd9c64d208f4d63e727bccc0e39f9329310",
		},
		map[string]string{
			"Yakd": "docker.io",
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
				delete(newImages, name)
			}
		}
		// Use newly configured custom images.
		images = overrideDefaults(images, newImages)
		// Store custom addon images to be written.
		cc.CustomAddonImages = mergeMaps(cc.CustomAddonImages, newImages)
	}

	// Use previously configured custom registries.
	customRegistries = filterKeySpace(addonDefaultImages, cc.CustomAddonRegistries) // filter by images map because registry map may omit default registry.
	if viper.IsSet(config.AddonRegistries) {
		// Parse the AddonRegistries flag if present.
		newRegistries := parseMapString(viper.GetString(config.AddonRegistries))
		for name := range newRegistries {
			if _, ok := addonDefaultImages[name]; !ok { // check images map because registry map may omitted default registry
				out.WarningT("Ignoring unknown custom registry {{.name}}", out.V{"name": name})
				delete(newRegistries, name)
			}
		}
		// Use newly configured custom registries.
		customRegistries = mergeMaps(customRegistries, newRegistries)
		// Merge newly set registries into custom addon registries to be written.
		cc.CustomAddonRegistries = mergeMaps(cc.CustomAddonRegistries, newRegistries)
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
		AutoPauseInterval       time.Duration
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
		AutoPauseInterval:       cc.AutoPauseInterval,
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
	opts.NetworkInfo["DNSDomain"] = cfg.DNSDomain

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
		// Passed on addon enable: --images=MetricsScraper=registry.k8s.io/echoserver:1.4
		// Without this line the resulting image would be docker.io/registry.k8s.io/echoserver:1.4
		if _, ok := cc.CustomAddonImages[name]; ok {
			opts.Registries[name] = ""
		}

		if enable {
			if override, ok := opts.CustomRegistries[name]; ok {
				out.Infof("Using image {{.registry}}{{.image}}", out.V{
					"registry": override,
					// removing the SHA from UI
					// SHA example docker.io/my_image:v0.0.4@sha256:65e9e69022aa7b0eb1e390e1916e3bf67f75ae5c25987f9154ef3b0e8ab8528b
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
