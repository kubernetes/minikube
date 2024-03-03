/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package kubevip

import (
	"bytes"
	"html/template"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
)

const Manifest = "kube-vip.yaml"

// KubeVipTemplate is kube-vip static pod config template
// ref: https://kube-vip.io/docs/installation/static/
// update: regenerate with:
//
//	export KVVERSION=$(curl -sL https://api.github.com/repos/kube-vip/kube-vip/releases | jq -r ".[0].name")
//	docker run --rm ghcr.io/kube-vip/kube-vip:$KVVERSION manifest pod --interface eth0 --address 192.168.42.17 --controlplane --arp --leaderElection
var kubeVipTemplate = template.Must(template.New("kubeletSystemdTemplate").Parse(`apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: null
  name: kube-vip
  namespace: kube-system
spec:
  containers:
  - args:
    - manager
    env:
    - name: vip_arp
      value: "true"
    - name: port
      value: "{{ .Port }}"
    - name: vip_interface
      value: eth0
    - name: vip_cidr
      value: "32"
    - name: dns_mode
      value: first
    - name: cp_enable
      value: "true"
    - name: cp_namespace
      value: kube-system
    - name: vip_leaderelection
      value: "true"
    - name: vip_leasename
      value: plndr-cp-lock
    - name: vip_leaseduration
      value: "5"
    - name: vip_renewdeadline
      value: "3"
    - name: vip_retryperiod
      value: "1"
    - name: address
      value: {{ .VIP }}
    - name: prometheus_server
      value: :2112
    image: ghcr.io/kube-vip/kube-vip:v0.7.1
    imagePullPolicy: IfNotPresent
    name: kube-vip
    resources: {}
    securityContext:
      capabilities:
        add:
        - NET_ADMIN
        - NET_RAW
    volumeMounts:
    - mountPath: /etc/kubernetes/admin.conf
      name: kubeconfig
  hostAliases:
  - hostnames:
    - kubernetes
    ip: 127.0.0.1
  hostNetwork: true
  volumes:
  - hostPath:
      path: "{{ .AdminConf }}"
    name: kubeconfig
status: {}
`))

// Configure takes last client ip address in cluster nodes network subnet as vip address and generates kube-vip.yaml file.
func Configure(cc config.ClusterConfig, workaround bool) ([]byte, error) {
	klog.Info("generating kube-vip config ...")

	params := struct {
		VIP       string
		Port      int
		AdminConf string
	}{
		VIP:       cc.KubernetesConfig.APIServerHAVIP,
		Port:      cc.APIServerPort,
		AdminConf: "/etc/kubernetes/admin.conf",
	}
	if workaround {
		params.AdminConf = "/etc/kubernetes/super-admin.conf"
	}

	b := bytes.Buffer{}
	if err := kubeVipTemplate.Execute(&b, params); err != nil {
		return nil, errors.Wrapf(err, "parse template")
	}

	klog.Infof("kube-vip config:\n%s", b.String())

	return b.Bytes(), nil
}
