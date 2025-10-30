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

package ktmpl

import "text/template"

// V1Beta1 is kubeadm config template for Kubernetes v1.13+
var V1Beta1 = template.Must(template.New("configTmpl-v1beta1").Funcs(template.FuncMap{
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
  name: "{{.NodeName}}"
  kubeletExtraArgs:
    node-ip: {{.NodeIP}}
  taints: []
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
{{ if .ImageRepository}}imageRepository: {{.ImageRepository}}
{{end}}{{range .ComponentOptions}}{{.Component}}:
{{- range $k, $v := .Pairs }}
  {{$k}}: {{$v}}
{{- end}}
  extraArgs:
{{- range $i, $val := printMapInOrder .ExtraArgs ": " }}
    {{$val}}
{{- end}}
{{end -}}
{{if .FeatureArgs}}featureGates:
{{range $i, $val := .FeatureArgs}}{{$i}}: {{$val}}
{{end -}}{{end -}}
certificatesDir: {{.CertDir}}
clusterName: {{.ClusterName}}
controlPlaneEndpoint: {{.ControlPlaneAddress}}:{{.APIServerPort}}
dns:
  type: CoreDNS
etcd:
  local:
    dataDir: {{.EtcdDataDir}}
    extraArgs:
      listen-metrics-urls: http://127.0.0.1:2381,http://{{.AdvertiseAddress}}:2381
kubernetesVersion: {{.KubernetesVersion}}
networking:
  dnsDomain: {{if .DNSDomain}}{{.DNSDomain}}{{else}}cluster.local{{end}}
  podSubnet: "{{.PodSubnet }}"
  serviceSubnet: {{.ServiceCIDR}}
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
authentication:
  x509:
    clientCAFile: {{.ClientCAFile}}
cgroupDriver: {{.CgroupDriver}}
{{- range $key, $val := .KubeletConfigOpts}}
{{$key}}: {{$val}}
{{- end}}
clusterDomain: "{{if .DNSDomain}}{{.DNSDomain}}{{else}}cluster.local{{end}}"
# disable disk resource management by default
imageGCHighThresholdPercent: 100
evictionHard:
  nodefs.available: "0%"
  nodefs.inodesFree: "0%"
  imagefs.available: "0%"
failSwapOn: false
staticPodPath: {{.StaticPodPath}}
---
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
clusterCIDR: "{{.PodSubnet }}"
metricsBindAddress: {{.KubeProxyMetricsBindAddress}}
conntrack:
  maxPerCore: 0
# Skip setting "net.netfilter.nf_conntrack_tcp_timeout_established"
  tcpEstablishedTimeout: 0s
# Skip setting "net.netfilter.nf_conntrack_tcp_timeout_close"
  tcpCloseWaitTimeout: 0s
{{- range $i, $val := printMapInOrder .KubeProxyOptions ": " }}
{{$val}}
{{- end}}
`))
