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

// V1Alpha1 is for Kubernetes v1.11
var V1Alpha1 = template.Must(template.New("configTmpl-v1alpha1").Funcs(template.FuncMap{
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
apiServerCertSANs: ["127.0.0.1", "localhost", "{{.AdvertiseAddress}}"]
{{if .ImageRepository}}imageRepository: {{.ImageRepository}}
{{end}}{{if .CRISocket}}criSocket: {{.CRISocket}}
{{end}}{{range .ExtraArgs}}{{.Component}}ExtraArgs:{{range $i, $val := printMapInOrder .Options ": " }}
  {{$val}}{{end}}
{{end}}{{if .FeatureArgs}}featureGates: {{range $i, $val := .FeatureArgs}}
  {{$i}}: {{$val}}{{end}}
{{end}}`))
