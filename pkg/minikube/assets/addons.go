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
	"runtime"
	"strings"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/version"
)

// Addon is a named list of assets, that can be enabled
type Addon struct {
	Assets    []*BinAsset
	enabled   bool
	addonName string
	Images    map[string]string

	// Registries currently only shows the default registry of images
	Registries map[string]string
}

// NewAddon creates a new Addon
func NewAddon(assets []*BinAsset, enabled bool, addonName string, images map[string]string, registries map[string]string) *Addon {
	a := &Addon{
		Assets:     assets,
		enabled:    enabled,
		addonName:  addonName,
		Images:     images,
		Registries: registries,
	}
	return a
}

// Name get the addon name
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
	"dashboard": NewAddon([]*BinAsset{
		// We want to create the kubernetes-dashboard ns first so that every subsequent object can be created
		MustBinAsset("deploy/addons/dashboard/dashboard-ns.yaml", vmpath.GuestAddonsDir, "dashboard-ns.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-clusterrole.yaml", vmpath.GuestAddonsDir, "dashboard-clusterrole.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-clusterrolebinding.yaml", vmpath.GuestAddonsDir, "dashboard-clusterrolebinding.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-configmap.yaml", vmpath.GuestAddonsDir, "dashboard-configmap.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-dp.yaml.tmpl", vmpath.GuestAddonsDir, "dashboard-dp.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-role.yaml", vmpath.GuestAddonsDir, "dashboard-role.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-rolebinding.yaml", vmpath.GuestAddonsDir, "dashboard-rolebinding.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-sa.yaml", vmpath.GuestAddonsDir, "dashboard-sa.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-secret.yaml", vmpath.GuestAddonsDir, "dashboard-secret.yaml", "0640"),
		MustBinAsset("deploy/addons/dashboard/dashboard-svc.yaml", vmpath.GuestAddonsDir, "dashboard-svc.yaml", "0640"),
	}, false, "dashboard", map[string]string{
		"Dashboard":      "kubernetesui/dashboard:v2.1.0@sha256:7f80b5ba141bead69c4fee8661464857af300d7d7ed0274cf7beecedc00322e6",
		"MetricsScraper": "kubernetesui/metrics-scraper:v1.0.4@sha256:555981a24f184420f3be0c79d4efb6c948a85cfce84034f85a563f4151a81cbf",
	}, nil),
	"default-storageclass": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/storageclass/storageclass.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storageclass.yaml",
			"0640"),
	}, true, "default-storageclass", nil, nil),
	"pod-security-policy": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/pod-security-policy/pod-security-policy.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"pod-security-policy.yaml",
			"0640"),
	}, false, "pod-security-policy", nil, nil),
	"storage-provisioner": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/storage-provisioner/storage-provisioner.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storage-provisioner.yaml",
			"0640"),
	}, true, "storage-provisioner", map[string]string{
		"StorageProvisioner": fmt.Sprintf("k8s-minikube/storage-provisioner:%s", version.GetStorageProvisionerVersion()),
	}, map[string]string{
		"StorageProvisioner": "gcr.io",
	}),
	"storage-provisioner-gluster": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/storage-provisioner-gluster/storage-gluster-ns.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storage-gluster-ns.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/storage-provisioner-gluster/glusterfs-daemonset.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"glusterfs-daemonset.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/storage-provisioner-gluster/heketi-deployment.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"heketi-deployment.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/storage-provisioner-gluster/storage-provisioner-glusterfile.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"storage-privisioner-glusterfile.yaml",
			"0640"),
	}, false, "storage-provisioner-gluster", map[string]string{
		"Heketi":                 "heketi/heketi:10@sha256:76d5a6a3b7cf083d1e99efa1c15abedbc5c8b73bef3ade299ce9a4c16c9660f8",
		"GlusterfileProvisioner": "gluster/glusterfile-provisioner:latest@sha256:9961a35cb3f06701958e202324141c30024b195579e5eb1704599659ddea5223",
		"GlusterfsServer":        "nixpanic/glusterfs-server:pr_fake-disk@sha256:3c58ae9d4e2007758954879d3f4095533831eb757c64ca6a0e32d1fc53fb6034",
	}, map[string]string{
		"GlusterfsServer": "quay.io",
	}),
	"efk": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/efk/elasticsearch-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"elasticsearch-rc.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/efk/elasticsearch-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"elasticsearch-svc.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/efk/fluentd-es-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"fluentd-es-rc.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/efk/fluentd-es-configmap.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"fluentd-es-configmap.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/efk/kibana-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"kibana-rc.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/efk/kibana-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"kibana-svc.yaml",
			"0640"),
	}, false, "efk", map[string]string{
		"Elasticsearch":        "elasticsearch:v5.6.2@sha256:7e95b32a7a2aad0c0db5c881e4a1ce8b7e53236144ae9d9cfb5fbe5608af4ab2",
		"FluentdElasticsearch": "fluentd-elasticsearch:v2.0.2@sha256:d0480bbf2d0de2344036fa3f7034cf7b4b98025a89c71d7f1f1845ac0e7d5a97",
		"Alpine":               "alpine:3.6@sha256:66790a2b79e1ea3e1dabac43990c54aca5d1ddf268d9a5a0285e4167c8b24475",
		"Kibana":               "kibana/kibana:5.6.2@sha256:cd948a9bda4622f1437afc4a3e78be6c8c25fc62f40aa0376f3d690f2436568f",
	}, map[string]string{
		"Elasticsearch":        "k8s.gcr.io",
		"FluentdElasticsearch": "k8s.gcr.io",
		"Kibana":               "docker.elastic.co",
	}),
	"ingress": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/ingress/ingress-configmap.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ingress-configmap.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/ingress/ingress-rbac.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ingress-rbac.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/ingress/ingress-dp.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ingress-dp.yaml",
			"0640"),
	}, false, "ingress", map[string]string{
		"IngressController":        "k8s-artifacts-prod/ingress-nginx/controller:v0.40.2@sha256:46ba23c3fbaafd9e5bd01ea85b2f921d9f2217be082580edc22e6c704a83f02f",
		"KubeWebhookCertgenCreate": "jettech/kube-webhook-certgen:v1.2.2@sha256:da8122a78d7387909cf34a0f34db0cce672da1379ee4fd57c626a4afe9ac12b7",
		"KubeWebhookCertgenPatch":  "jettech/kube-webhook-certgen:v1.3.0@sha256:ff01fba91131ed260df3f3793009efbf9686f5a5ce78a85f81c386a4403f7689",
	}, map[string]string{
		"IngressController": "us.gcr.io",
	}),
	"istio-provisioner": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/istio-provisioner/istio-operator.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"istio-operator.yaml",
			"0640"),
	}, false, "istio-provisioner", map[string]string{
		"IstioOperator": "istio/operator:1.5.0@sha256:25a6398ed4996a5313767ceb63768d503c266f63506ad3074b30eef6b5b5167e",
	}, nil),
	"istio": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/istio/istio-default-profile.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"istio-default-profile.yaml",
			"0640"),
	}, false, "istio", nil, nil),
	"kubevirt": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/kubevirt/pod.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"pod.yaml",
			"0640"),
	}, false, "kubevirt", map[string]string{
		"Kubectl": "bitnami/kubectl:1.17@sha256:de642e973d3d0ef60e4d0a1f92286a9fdae245535c5990d4762bbe86fcf95887",
	}, nil),
	"metrics-server": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/metrics-server/metrics-apiservice.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metrics-apiservice.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/metrics-server/metrics-server-deployment.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metrics-server-deployment.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/metrics-server/metrics-server-service.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metrics-server-service.yaml",
			"0640"),
	}, false, "metrics-server", map[string]string{
		"MetricsServer": fmt.Sprintf("metrics-server-%s:v0.2.1", runtime.GOARCH),
	}, map[string]string{
		"MetricsServer": "k8s.gcr.io",
	}),
	"olm": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/olm/crds.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"crds.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/olm/olm.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"olm.yaml",
			"0640"),
	}, false, "olm", map[string]string{
		"OLM":                        "operator-framework/olm:0.14.1@sha256:0d15ffb5d10a176ef6e831d7865f98d51255ea5b0d16403618c94a004d049373",
		"UpstreamCommunityOperators": "operator-framework/upstream-community-operators:07bbc13@sha256:cc7b3fdaa1ccdea5866fcd171669dc0ed88d3477779d8ed32e3712c827e38cc0",
	}, map[string]string{
		"OLM":                        "quay.io",
		"UpstreamCommunityOperators": "quay.io",
	}),
	"registry": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/registry/registry-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-rc.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/registry/registry-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-svc.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/registry/registry-proxy.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-proxy.yaml",
			"0640"),
	}, false, "registry", map[string]string{
		"Registry":          "registry:2.7.1@sha256:d5459fcb27aecc752520df4b492b08358a1912fcdfa454f7d2101d4b09991daa",
		"KubeRegistryProxy": "google_containers/kube-registry-proxy:0.4@sha256:1040f25a5273de0d72c54865a8efd47e3292de9fb8e5353e3fa76736b854f2da",
	}, map[string]string{
		"KubeRegistryProxy": "gcr.io",
	}),
	"registry-creds": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/registry-creds/registry-creds-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"registry-creds-rc.yaml",
			"0640"),
	}, false, "registry-creds", map[string]string{
		"RegistryCreds": "upmcenterprises/registry-creds:1.10@sha256:93a633d4f2b76a1c66bf19c664dbddc56093a543de6d54320f19f585ccd7d605",
	}, nil),
	"registry-aliases": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/registry-aliases/registry-aliases-sa.tmpl",
			vmpath.GuestAddonsDir,
			"registry-aliases-sa.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/registry-aliases/registry-aliases-sa-crb.tmpl",
			vmpath.GuestAddonsDir,
			"registry-aliases-sa-crb.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/registry-aliases/registry-aliases-config.tmpl",
			vmpath.GuestAddonsDir,
			"registry-aliases-config.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/registry-aliases/node-etc-hosts-update.tmpl",
			vmpath.GuestAddonsDir,
			"node-etc-hosts-update.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/registry-aliases/patch-coredns-job.tmpl",
			vmpath.GuestAddonsDir,
			"patch-coredns-job.yaml",
			"0640"),
	}, false, "registry-aliases", map[string]string{
		"CoreDNSPatcher": "rhdevelopers/core-dns-patcher@sha256:9220ff32f690c3d889a52afb59ca6fcbbdbd99e5370550cc6fd249adea8ed0a9",
		"Alpine":         "alpine:3.11@sha256:0bd0e9e03a022c3b0226667621da84fc9bf562a9056130424b5bfbd8bcb0397f",
		"Pause":          "google_containers/pause:3.1@sha256:f78411e19d84a252e53bff71a4407a5686c46983a2c2eeed83929b888179acea",
	}, map[string]string{
		"CoreDNSPatcher": "quay.io",
		"Pause":          "gcr.io",
	}),
	"freshpod": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/freshpod/freshpod-rc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"freshpod-rc.yaml",
			"0640"),
	}, false, "freshpod", map[string]string{
		"FreshPod": "google-samples/freshpod:v0.0.1@sha256:b9efde5b509da3fd2959519c4147b653d0c5cefe8a00314e2888e35ecbcb46f9",
	}, map[string]string{
		"FreshPod": "gcr.io",
	}),
	"nvidia-driver-installer": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/gpu/nvidia-driver-installer.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"nvidia-driver-installer.yaml",
			"0640"),
	}, false, "nvidia-driver-installer", map[string]string{
		"NvidiaDriverInstaller": "minikube-nvidia-driver-installer:e2d9b43228decf5d6f7dce3f0a85d390f138fa01",
		"Pause":                 "pause:2.0@sha256:9ce5316f9752b8347484ab0f6778573af15524124d52b93230b9a0dcc987e73e",
	}, map[string]string{
		"NvidiaDriverInstaller": "k8s.gcr.io",
		"Pause":                 "k8s.gcr.io",
	}),
	"nvidia-gpu-device-plugin": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/gpu/nvidia-gpu-device-plugin.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"nvidia-gpu-device-plugin.yaml",
			"0640"),
	}, false, "nvidia-gpu-device-plugin", map[string]string{
		"NvidiaDevicePlugin": "nvidia/k8s-device-plugin:1.0.0-beta4@sha256:94d46bf513cbc43c4d77a364e4bbd409d32d89c8e686e12551cc3eb27c259b90",
	}, nil),
	"logviewer": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/logviewer/logviewer-dp-and-svc.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"logviewer-dp-and-svc.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/logviewer/logviewer-rbac.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"logviewer-rbac.yaml",
			"0640"),
	}, false, "logviewer", map[string]string{
		"LogViewer": "ivans3/minikube-log-viewer:latest@sha256:75854f45305cc47d17b04c6c588fa60777391761f951e3a34161ddf1f1b06405",
	}, nil),
	"gvisor": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/gvisor/gvisor-pod.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gvisor-pod.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/gvisor/gvisor-runtimeclass.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gvisor-runtimeclass.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/gvisor/gvisor-config.toml",
			vmpath.GuestGvisorDir,
			constants.GvisorConfigTomlTargetName,
			"0640"),
	}, false, "gvisor", map[string]string{
		"GvisorAddon": "k8s-minikube/gvisor-addon:3@sha256:23eb17d48a66fc2b09c31454fb54ecae520c3e9c9197ef17fcb398b4f31d505a",
	}, map[string]string{
		"GvisorAddon": "gcr.io",
	}),
	"helm-tiller": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/helm-tiller/helm-tiller-dp.tmpl",
			vmpath.GuestAddonsDir,
			"helm-tiller-dp.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/helm-tiller/helm-tiller-rbac.tmpl",
			vmpath.GuestAddonsDir,
			"helm-tiller-rbac.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/helm-tiller/helm-tiller-svc.tmpl",
			vmpath.GuestAddonsDir,
			"helm-tiller-svc.yaml",
			"0640"),
	}, false, "helm-tiller", map[string]string{
		"Tiller": "kubernetes-helm/tiller:v2.16.12@sha256:6003775d503546087266eda39418d221f9afb5ccfe35f637c32a1161619a3f9c",
	}, map[string]string{
		"Tiller": "gcr.io",
	}),
	"ingress-dns": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/ingress-dns/ingress-dns-pod.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ingress-dns-pod.yaml",
			"0640"),
	}, false, "ingress-dns", map[string]string{
		"IngressDNS": "cryptexlabs/minikube-ingress-dns:0.3.0@sha256:e252d2a4c704027342b303cc563e95d2e71d2a0f1404f55d676390e28d5093ab",
	}, nil),
	"metallb": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/metallb/metallb.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metallb.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/metallb/metallb-config.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"metallb-config.yaml",
			"0640"),
	}, false, "metallb", map[string]string{
		"Speaker":    "metallb/speaker:v0.8.2@sha256:f1941498a28cdb332429e25d18233683da6949ecfc4f6dacf12b1416d7d38263",
		"Controller": "metallb/controller:v0.8.2@sha256:5c050e59074e152711737d2bb9ede96dff67016c80cf25cdf5fc46109718a583",
	}, nil),
	"ambassador": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/ambassador/ambassador-operator-crds.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ambassador-operator-crds.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/ambassador/ambassador-operator.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ambassador-operator.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/ambassador/ambassadorinstallation.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"ambassadorinstallation.yaml",
			"0640"),
	}, false, "ambassador", map[string]string{
		"AmbassadorOperator": "datawire/ambassador-operator:v1.2.3@sha256:492f33e0828a371aa23331d75c11c251b21499e31287f026269e3f6ec6da34ed",
	}, map[string]string{
		"AmbassadorOperator": "quay.io",
	}),
	"gcp-auth": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/gcp-auth/gcp-auth-ns.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gcp-auth-ns.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/gcp-auth/gcp-auth-service.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"gcp-auth-service.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/gcp-auth/gcp-auth-webhook.yaml.tmpl.tmpl",
			vmpath.GuestAddonsDir,
			"gcp-auth-webhook.yaml",
			"0640"),
	}, false, "gcp-auth", map[string]string{
		"KubeWebhookCertgen": "jettech/kube-webhook-certgen:v1.3.0@sha256:ff01fba91131ed260df3f3793009efbf9686f5a5ce78a85f81c386a4403f7689",
		"GCPAuthWebhook":     "k8s-minikube/gcp-auth-webhook:v0.0.4@sha256:65e9e69022aa7b0eb1e390e1916e3bf67f75ae5c25987f9154ef3b0e8ab8528b",
	}, map[string]string{
		"GCPAuthWebhook": "gcr.io",
	}),
	"volumesnapshots": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/volumesnapshots/snapshot.storage.k8s.io_volumesnapshotclasses.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshotclasses.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/volumesnapshots/snapshot.storage.k8s.io_volumesnapshotcontents.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshotcontents.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/volumesnapshots/snapshot.storage.k8s.io_volumesnapshots.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"snapshot.storage.k8s.io_volumesnapshots.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/volumesnapshots/rbac-volume-snapshot-controller.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-volume-snapshot-controller.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/volumesnapshots/volume-snapshot-controller-deployment.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"volume-snapshot-controller-deployment.yaml",
			"0640"),
	}, false, "volumesnapshots", map[string]string{
		"SnapshotController": "k8s-staging-csi/snapshot-controller:v2.0.0-rc2@sha256:9a44a869d23e42f5d7954c9a5c9ec1a76a0a5d6f23fce5e68e1232a017d3d38c",
	}, map[string]string{
		"SnapshotController": "gcr.io",
	}),
	"csi-hostpath-driver": NewAddon([]*BinAsset{
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/rbac/rbac-external-attacher.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-attacher.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/rbac/rbac-external-provisioner.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-provisioner.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/rbac/rbac-external-resizer.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-resizer.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/rbac/rbac-external-snapshotter.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"rbac-external-snapshotter.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/deploy/csi-hostpath-attacher.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-attacher.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/deploy/csi-hostpath-driverinfo.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-driverinfo.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/deploy/csi-hostpath-plugin.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-plugin.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/deploy/csi-hostpath-provisioner.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-provisioner.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/deploy/csi-hostpath-resizer.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-resizer.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/deploy/csi-hostpath-snapshotter.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-snapshotter.yaml",
			"0640"),
		MustBinAsset(
			"deploy/addons/csi-hostpath-driver/deploy/csi-hostpath-storageclass.yaml.tmpl",
			vmpath.GuestAddonsDir,
			"csi-hostpath-storageclass.yaml",
			"0640"),
	}, false, "csi-hostpath-driver", map[string]string{
		"Attacher":            "k8scsi/csi-attacher:v3.0.0-rc1@sha256:8fcb9472310dd424c4da8ee06ff200b5e6f091dff39a079e470599e4d0dcf328",
		"NodeDriverRegistrar": "k8scsi/csi-node-driver-registrar:v1.3.0@sha256:9622c6a6dac7499a055a382930f4de82905a3c5735c0753f7094115c9c871309",
		"HostPathPlugin":      "k8scsi/hostpathplugin:v1.4.0-rc2@sha256:aa223f9df8c1d477a9f2a4a2a7d104561e6d365e54671aacbc770dffcc0683ad",
		"LivenessProbe":       "k8scsi/livenessprobe:v1.1.0@sha256:dde617756e0f602adc566ab71fd885f1dad451ad3fb063ac991c95a2ff47aea5",
		"Resizer":             "k8scsi/csi-resizer:v0.6.0-rc1@sha256:75ad39004ac49267981c9cb3323a7f73f0b203e1c181117363bf215e10144e8a",
		"Snapshotter":         "k8scsi/csi-snapshotter:v2.1.0@sha256:35ead85dd09aa8cc612fdb598d4e0e2f048bef816f1b74df5eeab67cd21b10aa",
		"Provisioner":         "k8s-staging-sig-storage/csi-provisioner:v2.0.0-rc2@sha256:8f36191970a82677ffe222007b08395dd7af0a5bb5b93db0e82523b43de2bfb2",
	}, map[string]string{
		"Attacher":            "quay.io",
		"NodeDriverRegistrar": "quay.io",
		"HostPathPlugin":      "quay.io",
		"LivenessProbe":       "quay.io",
		"Resizer":             "quay.io",
		"Snapshotter":         "quay.io",
		"Provisioner":         "gcr.io",
	}),
}

// GenerateTemplateData generates template data for template assets
func GenerateTemplateData(addon *Addon, cfg config.KubernetesConfig) interface{} {

	a := runtime.GOARCH
	// Some legacy docker images still need the -arch suffix
	// for  less common architectures blank suffix for amd64
	ea := ""
	if runtime.GOARCH != "amd64" {
		ea = "-" + runtime.GOARCH
	}
	opts := struct {
		Arch                string
		ExoticArch          string
		ImageRepository     string
		LoadBalancerStartIP string
		LoadBalancerEndIP   string
		CustomIngressCert   string
		Images              map[string]string
		Registries          map[string]string
		CustomRegistries    map[string]string
	}{
		Arch:                a,
		ExoticArch:          ea,
		ImageRepository:     cfg.ImageRepository,
		LoadBalancerStartIP: cfg.LoadBalancerStartIP,
		LoadBalancerEndIP:   cfg.LoadBalancerEndIP,
		CustomIngressCert:   cfg.CustomIngressCert,
		Images:              addon.Images,
		Registries:          addon.Registries,
		CustomRegistries:    make(map[string]string),
	}
	if opts.ImageRepository != "" && !strings.HasSuffix(opts.ImageRepository, "/") {
		opts.ImageRepository += "/"
	}

	if opts.Images == nil {
		opts.Images = make(map[string]string) // Avoid nil access when rendering
	}

	images := viper.GetString(config.AddonImages)
	if images != "" {
		for _, image := range strings.Split(images, ",") {
			vals := strings.Split(image, "=")
			if len(vals) != 2 || vals[1] == "" {
				out.WarningT("Ignoring invalid custom image {{.conf}}", out.V{"conf": image})
				continue
			}
			if _, ok := opts.Images[vals[0]]; ok {
				opts.Images[vals[0]] = vals[1]
			} else {
				out.WarningT("Ignoring unknown custom image {{.name}}", out.V{"name": vals[0]})
			}
		}
	}

	if opts.Registries == nil {
		opts.Registries = make(map[string]string)
	}

	registries := viper.GetString(config.AddonRegistries)
	if registries != "" {
		for _, registry := range strings.Split(registries, ",") {
			vals := strings.Split(registry, "=")
			if len(vals) != 2 {
				out.WarningT("Ignoring invalid custom registry {{.conf}}", out.V{"conf": registry})
				continue
			}
			if _, ok := opts.Images[vals[0]]; ok { // check images map because registry map may omitted default registry
				opts.CustomRegistries[vals[0]] = vals[1]
			} else {
				out.WarningT("Ignoring unknown custom registry {{.name}}", out.V{"name": vals[0]})
			}
		}
	}

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
	return opts
}
