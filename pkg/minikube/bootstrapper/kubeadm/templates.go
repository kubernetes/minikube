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

package kubeadm

import (
	"fmt"
	"sort"
	"text/template"
)

// configTmplV1Alpha1 is for Kubernetes v1.11
var configTmplV1Alpha1 = template.Must(template.New("configTmpl-v1alpha1").Funcs(template.FuncMap{
	"printMapInOrder": printMapInOrder,
}).Parse(`apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
{{if .NoTaintMaster}}noTaintMaster: true{{end}}
api:
  advertiseAddress: {{.AdvertiseAddress}}
  bindPort: {{.APIServerPort}}
  controlPlaneEndpoint: localhost
kubernetesVersion: {{.KubernetesVersion}}
certificatesDir: {{.CertDir}}
networking:
  serviceSubnet: {{.ServiceCIDR}}
etcd:
  dataDir: {{.EtcdDataDir}}
nodeName: {{.NodeName}}
{{if .ImageRepository}}imageRepository: {{.ImageRepository}}
{{end}}{{if .CRISocket}}criSocket: {{.CRISocket}}
{{end}}{{range .ExtraArgs}}{{.Component}}ExtraArgs:{{range $i, $val := printMapInOrder .Options ": " }}
  {{$val}}{{end}}
{{end}}{{if .FeatureArgs}}featureGates: {{range $i, $val := .FeatureArgs}}
  {{$i}}: {{$val}}{{end}}
{{end}}`))

// configTmplV1Alpha3 is for Kubernetes v1.12
var configTmplV1Alpha3 = template.Must(template.New("configTmpl-v1alpha3").Funcs(template.FuncMap{
	"printMapInOrder": printMapInOrder,
}).Parse(`apiVersion: kubeadm.k8s.io/v1alpha3
kind: InitConfiguration
apiEndpoint:
  advertiseAddress: {{.AdvertiseAddress}}
  bindPort: {{.APIServerPort}}
bootstrapTokens:
  - groups:
      - system:bootstrappers:kubeadm:default-node-token
    ttl: 24h0m0s
    usages:
      - signing
      - authentication
nodeRegistration:
  criSocket: {{if .CRISocket}}{{.CRISocket}}{{else}}/var/run/dockershim.sock{{end}}
  name: {{.NodeName}}
  taints: []
---
apiVersion: kubeadm.k8s.io/v1alpha3
kind: ClusterConfiguration
{{if .ImageRepository}}imageRepository: {{.ImageRepository}}
{{end}}{{range .ExtraArgs}}{{.Component}}ExtraArgs:{{range $i, $val := printMapInOrder .Options ": " }}
  {{$val}}{{end}}
{{end -}}
{{if .FeatureArgs}}featureGates: {{range $i, $val := .FeatureArgs}}
  {{$i}}: {{$val}}{{end}}
{{end -}}
certificatesDir: {{.CertDir}}
clusterName: kubernetes
controlPlaneEndpoint: localhost:{{.APIServerPort}}
etcd:
  local:
    dataDir: {{.EtcdDataDir}}
kubernetesVersion: {{.KubernetesVersion}}
networking:
  dnsDomain: cluster.local
  podSubnet: ""
  serviceSubnet: {{.ServiceCIDR}}
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
evictionHard:
  nodefs.available: "0%"
  nodefs.inodesFree: "0%"
  imagefs.available: "0%"
`))

// configTmplV1Beta1 is for Kubernetes v1.13+
var configTmplV1Beta1 = template.Must(template.New("configTmpl-v1beta1").Funcs(template.FuncMap{
	"printMapInOrder": printMapInOrder,
}).Parse(`apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: {{.AdvertiseAddress}}
  bindPort: {{.APIServerPort}}
bootstrapTokens:
  - groups:
      - system:bootstrappers:kubeadm:default-node-token
    ttl: 24h0m0s
    usages:
      - signing
      - authentication
nodeRegistration:
  criSocket: {{if .CRISocket}}{{.CRISocket}}{{else}}/var/run/dockershim.sock{{end}}
  name: {{.NodeName}}
  taints: []
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
{{ if .ImageRepository}}imageRepository: {{.ImageRepository}}
{{end}}{{range .ExtraArgs}}{{.Component}}:
  extraArgs:
{{- range $i, $val := printMapInOrder .Options ": " }}
    {{$val}}
{{- end}}
{{end -}}
{{if .FeatureArgs}}featureGates:
{{range $i, $val := .FeatureArgs}}{{$i}}: {{$val}}
{{end -}}{{end -}}
certificatesDir: {{.CertDir}}
clusterName: kubernetes
controlPlaneEndpoint: localhost:{{.APIServerPort}}
dns:
  type: CoreDNS
etcd:
  local:
    dataDir: {{.EtcdDataDir}}
kubernetesVersion: {{.KubernetesVersion}}
networking:
  dnsDomain: cluster.local
  podSubnet: ""
  serviceSubnet: {{.ServiceCIDR}}
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
imageGCHighThresholdPercent: 100
evictionHard:
  nodefs.available: "0%"
  nodefs.inodesFree: "0%"
  imagefs.available: "0%"
`))

var kubeletSystemdTemplate = template.Must(template.New("kubeletSystemdTemplate").Parse(`
[Unit]
{{if or (eq .ContainerRuntime "cri-o") (eq .ContainerRuntime "cri")}}Wants=crio.service{{else}}Wants=docker.socket{{end}}

[Service]
ExecStart=
ExecStart=/usr/bin/kubelet{{if .ExtraOptions}} {{.ExtraOptions}}{{end}}

[Install]
`))

const kubeletService = `
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=http://kubernetes.io/docs/

[Service]
ExecStart=/usr/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
`

var kubeadmInitTemplate = template.Must(template.New("kubeadmInitTemplate").Parse(`
sudo /usr/bin/kubeadm init --config {{.KubeadmConfigFile}} {{.ExtraOptions}} {{if .SkipPreflightChecks}}--skip-preflight-checks{{else}}{{range .Preflights}}--ignore-preflight-errors={{.}} {{end}}{{end}}
`))

// printMapInOrder sorts the keys and prints the map in order, combining key
// value pairs with the separator character
//
// Note: this is not necessary, but makes testing easy
func printMapInOrder(m map[string]string, sep string) []string {
	if m == nil {
		return nil
	}
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		keys[i] = fmt.Sprintf("%s%s\"%s\"", k, sep, m[k])
	}
	return keys
}
