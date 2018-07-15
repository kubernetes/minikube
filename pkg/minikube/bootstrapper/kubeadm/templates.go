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

var kubeadmConfigTemplate = template.Must(template.New("kubeadmConfigTemplate").Funcs(template.FuncMap{
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
{{range .ExtraArgs}}{{.Component}}:{{range $i, $val := printMapInOrder .Options ": " }}
  {{$val}}{{end}}
{{end}}{{if .FeatureArgs}}featureGates: {{range $i, $val := .FeatureArgs}}
  {{$i}}: {{$val}}{{end}}
{{end}}`))

var kubeletSystemdTemplate = template.Must(template.New("kubeletSystemdTemplate").Parse(`
[Unit]
{{if or (eq .ContainerRuntime "cri-o") (eq .ContainerRuntime "cri")}}Wants=crio.service{{else}}Wants=docker.socket{{end}}

[Service]
ExecStart=
ExecStart=/usr/bin/kubelet {{.ExtraOptions}} {{if .FeatureGates}}--feature-gates={{.FeatureGates}}{{end}}

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

var kubeadmRestoreTemplate = template.Must(template.New("kubeadmRestoreTemplate").Parse(`
sudo kubeadm alpha phase certs all --config {{.KubeadmConfigFile}} &&
sudo /usr/bin/kubeadm alpha phase kubeconfig all --config {{.KubeadmConfigFile}} &&
sudo /usr/bin/kubeadm alpha phase controlplane all --config {{.KubeadmConfigFile}} &&
sudo /usr/bin/kubeadm alpha phase etcd local --config {{.KubeadmConfigFile}}
`))

var kubeadmInitTemplate = template.Must(template.New("kubeadmInitTemplate").Parse(`
sudo /usr/bin/kubeadm init --config {{.KubeadmConfigFile}} {{if .SkipPreflightChecks}}--skip-preflight-checks{{else}}{{range .Preflights}}--ignore-preflight-errors={{.}} {{end}}{{end}} &&
sudo /usr/bin/kubeadm alpha phase addon kube-dns
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
