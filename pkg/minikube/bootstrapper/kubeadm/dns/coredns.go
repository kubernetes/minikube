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

// minikube-specific CoreDNS manifests based on the default kubeadm's embedded ones
// ref: https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/phases/addons/dns/manifests.go

package dns

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os/exec"
	"path"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/vmpath"
	kconst "k8s.io/minikube/third_party/kubeadm/app/constants"
)

const (
	// CoreDNSService is the CoreDNS Service manifest
	CoreDNSService = `
apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: "CoreDNS"
  name: kube-dns
  namespace: kube-system
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
  # Without this resourceVersion value, an update of the Service between versions will yield:
  #   Service "kube-dns" is invalid: metadata.resourceVersion: Invalid value: "": must be specified for an update
  resourceVersion: "0"
spec:
  clusterIP: {{ .DNSIP }}
  ports:
  - name: dns
    port: 53
    protocol: UDP
    targetPort: 53
  - name: dns-tcp
    port: 53
    protocol: TCP
    targetPort: 53
  - name: metrics
    port: 9153
    protocol: TCP
    targetPort: 9153
  selector:
    k8s-app: kube-dns
`

	// CoreDNSDeployment is the CoreDNS Deployment manifest
	CoreDNSDeployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .DeploymentName }}
  namespace: kube-system
  labels:
    k8s-app: kube-dns
spec:
  replicas: {{ .Replicas }}
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: kube-dns
  template:
    metadata:
      labels:
        k8s-app: kube-dns
    spec:
      priorityClassName: system-cluster-critical
      serviceAccountName: coredns
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: k8s-app
                  operator: In
                  values: ["kube-dns"]
              topologyKey: kubernetes.io/hostname
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      - key: {{ .ControlPlaneTaintKey }}
        effect: NoSchedule
      nodeSelector:
        kubernetes.io/os: linux
      containers:
      - name: coredns
        image: {{ .Image }}
        imagePullPolicy: IfNotPresent
        resources:
          limits:
            memory: 170Mi
          requests:
            cpu: 100m
            memory: 70Mi
        args: [ "-conf", "/etc/coredns/Corefile" ]
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
        ports:
        - containerPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          name: metrics
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
          initialDelaySeconds: 60
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: 8181
            scheme: HTTP
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
      dnsPolicy: Default
      volumes:
        - name: config-volume
          configMap:
            name: coredns
            items:
            - key: Corefile
              path: Corefile
`

	// CoreDNSConfigMap is the CoreDNS ConfigMap manifest
	CoreDNSConfigMap = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        log
        errors
        health {
           lameduck 5s
        }
        ready
        kubernetes {{ .DNSDomain }} in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        hosts {
          {{ .MinikubeHostIP }} {{ .MinikubeHostFQDN }}
          fallthrough
        }
        forward . /etc/resolv.conf {
           max_concurrent 1000
        }
        cache 30
        loop
        reload
        loadbalance
    }
`
	// CoreDNSClusterRole is the CoreDNS ClusterRole manifest
	CoreDNSClusterRole = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:coredns
rules:
- apiGroups:
  - ""
  resources:
  - endpoints
  - services
  - pods
  - namespaces
  verbs:
  - list
  - watch
- apiGroups:
  - discovery.k8s.io
  resources:
  - endpointslices
  verbs:
  - list
  - watch
`
	// CoreDNSClusterRoleBinding is the CoreDNS Clusterrolebinding manifest
	CoreDNSClusterRoleBinding = `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: system:coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:coredns
subjects:
- kind: ServiceAccount
  name: coredns
  namespace: kube-system
`
	// CoreDNSServiceAccount is the CoreDNS ServiceAccount manifest
	CoreDNSServiceAccount = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: coredns
  namespace: kube-system
`
)

func DeployCoreDNS(cc config.ClusterConfig, r command.Runner, hostIP, hostFQDN string) error {
	manifests, err := coreDNSManifests(cc, hostIP, hostFQDN)
	if err != nil {
		return fmt.Errorf("coredns manifests: %v", err)
	}
	klog.Infof("coredns manifests:\n%s\n", manifests)

	// copy over manifests file
	manifestPath := path.Join(vmpath.GuestAddonsDir, "coredns.yaml")
	m := assets.NewMemoryAssetTarget(manifests, manifestPath, "0640")
	if err := r.Copy(m); err != nil {
		return fmt.Errorf("coredns asset copy: %v", err)
	}

	// apply manifests file
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	kubectl := kapi.KubectlBinaryPath(cc.KubernetesConfig.KubernetesVersion)
	klog.Infof("applying CoreDNS manifests using %s ...", kubectl)
	cmd := exec.CommandContext(ctx, "sudo", kubectl, "apply", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")), "-f", manifestPath)
	if rr, err := r.RunCmd(cmd); err != nil {
		return fmt.Errorf("coredns apply cmd: %q output: %q error: %v", rr.Command(), rr.Output(), err)
	}

	return nil
}

func coreDNSManifests(cc config.ClusterConfig, hostIP, hostFQDN string) ([]byte, error) {
	toml := CoreDNSServiceAccount + "---" +
		CoreDNSClusterRole + "---" +
		CoreDNSClusterRoleBinding + "---" +
		CoreDNSConfigMap + "---" +
		CoreDNSDeployment + "---" +
		CoreDNSService

	dnsip, err := kconst.GetDNSIP(cc.KubernetesConfig.ServiceCIDR, true)
	if err != nil {
		return nil, err
	}

	image := ""
	imgs, err := images.Kubeadm(cc.KubernetesConfig.ImageRepository, cc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("kubeadm images: %v", err)
	}
	for _, img := range imgs {
		if strings.Contains(img, kconst.CoreDNSImageName) {
			image = img
			break
		}
	}
	if image == "" {
		return nil, fmt.Errorf("coredns image not found")
	}

	var replicas int32 = 1

	params := struct {
		DNSDomain, MinikubeHostIP, MinikubeHostFQDN string
		DeploymentName, ControlPlaneTaintKey, Image string
		Replicas                                    *int32
		DNSIP                                       string
	}{
		DNSDomain:            cc.KubernetesConfig.DNSDomain,
		MinikubeHostIP:       hostIP,
		MinikubeHostFQDN:     hostFQDN,
		DeploymentName:       kconst.CoreDNSDeploymentName,
		ControlPlaneTaintKey: kconst.LabelNodeRoleControlPlane,
		Image:                image,
		Replicas:             &replicas,
		DNSIP:                dnsip.String(),
	}

	t := template.Must(template.New("coredns").Parse(toml))
	var manifests bytes.Buffer
	if err = t.Execute(&manifests, params); err != nil {
		return nil, fmt.Errorf("executing CoreDNS template: %v", err)
	}

	return manifests.Bytes(), nil
}
