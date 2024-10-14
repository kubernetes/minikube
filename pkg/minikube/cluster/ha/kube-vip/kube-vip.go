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
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
)

const Manifest = "kube-vip.yaml"

// KubeVipTemplate is kube-vip static pod config template
// ref: https://kube-vip.io/docs/installation/static/#generating-a-manifest
// note: to check if the latest kube-vip version introduces any significant changes, compare the current one with a latest default manifest generated with:
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
    - name: vip_nodename
      valueFrom:
        fieldRef:
          fieldPath: spec.nodeName
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
    {{- if .EnableLB }}
    - name : lb_enable
      value: "true"
    - name: lb_port
      value: "{{ .Port }}"
    {{- end}}
    image: ghcr.io/kube-vip/kube-vip:v0.8.4
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
func Configure(cc config.ClusterConfig, r command.Runner, kubeadmCfg []byte, workaround bool) ([]byte, error) {
	klog.Info("generating kube-vip config ...")

	params := struct {
		VIP       string
		Port      int
		AdminConf string
		EnableLB  bool
	}{
		VIP:       cc.KubernetesConfig.APIServerHAVIP,
		Port:      cc.APIServerPort,
		AdminConf: "/etc/kubernetes/admin.conf",
		EnableLB:  enableCPLB(cc, r, kubeadmCfg),
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

// enableCPLB auto-enables control-plane load-balancing, if possible - currently only possible with ipvs.
// ref: https://kube-vip.io/docs/about/architecture/?query=ipvs#control-plane-load-balancing
func enableCPLB(cc config.ClusterConfig, r command.Runner, kubeadmCfg []byte) bool {
	// note known issue: "service lb with ipvs mode won't work with kubeproxy that is configured with ipvs mode"
	// ref: https://kube-vip.io/docs/about/architecture/?query=ipvs#known-issues
	// so we only want to enable control-plane load-balancing if kube-proxy mode is not set to ipvs
	// ref: https://kubernetes.io/docs/reference/networking/virtual-ips/#proxy-mode-ipvs
	if ipvs := strings.EqualFold(string(kubeadmCfg), "mode: ipvs"); ipvs {
		klog.Info("giving up enabling control-plane load-balancing as kube-proxy mode appears to be set to ipvs")
		return false
	}

	// for vm driver, ensure required ipvs kernel modules are loaded to enable kube-vip's control-plane load-balancing feature
	// ref: https://github.com/kubernetes/kubernetes/blob/f90461c43e881d320b78d48793db10c110d488d1/pkg/proxy/ipvs/README.md?plain=1#L257-L269
	if driver.IsVM(cc.Driver) {
		if _, err := r.RunCmd(exec.Command("sudo", "sh", "-c", "modprobe --all ip_vs ip_vs_rr ip_vs_wrr ip_vs_sh nf_conntrack")); err != nil {
			klog.Warningf("unable to load ipvs kernel modules: %v", err)
			return false
		}
		// for non-vm driver, only try to check if required ipvs kernel modules are already loaded
	} else if _, err := r.RunCmd(exec.Command("sudo", "sh", "-c", "lsmod | grep ip_vs")); err != nil {
		klog.Infof("giving up enabling control-plane load-balancing as ipvs kernel modules appears not to be available: %v", err)
		return false
	}

	klog.Info("auto-enabling control-plane load-balancing in kube-vip")
	return true
}
