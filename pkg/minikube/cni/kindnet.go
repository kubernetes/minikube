/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package cni

import (
	"bytes"
	"os/exec"
	"text/template"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
)

var kindNetManifest = template.Must(template.New("kindnet").Parse(`---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kindnet
rules:
  - apiGroups:
    - policy
    resources:
    - podsecuritypolicies
    verbs:
    - use
    resourceNames: 
    - kindnet
  - apiGroups:
      - ""
    resources:
      - nodes
      - namespaces
      - pods
    verbs:
      - list
      - watch
      - patch
  - apiGroups:
     - ""
    resources:
      - configmaps
    verbs:
      - get
  - apiGroups:
      -  networking.k8s.io
    resources:
      - networkpolicies
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kindnet
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kindnet
subjects:
- kind: ServiceAccount
  name: kindnet
  namespace: kube-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kindnet
  namespace: kube-system
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: kindnet
  namespace: kube-system
  labels:
    tier: node
    app: kindnet
    k8s-app: kindnet
spec:
  selector:
    matchLabels:
      app: kindnet
  template:
    metadata:
      labels:
        tier: node
        app: kindnet
        k8s-app: kindnet
    spec:
      hostNetwork: true
      tolerations:
      - operator: Exists
        effect: NoSchedule
      serviceAccountName: kindnet
      containers:
      - name: kindnet-cni
        image: {{.ImageName}}
        env:
        - name: HOST_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: POD_SUBNET
          value: {{.PodCIDR}}
        volumeMounts:
        - name: cni-cfg
          mountPath: /etc/cni/net.d
        - name: xtables-lock
          mountPath: /run/xtables.lock
          readOnly: false
        - name: lib-modules
          mountPath: /lib/modules
          readOnly: true
        resources:
          requests:
            cpu: "100m"
            memory: "50Mi"
          limits:
            cpu: "100m"
            memory: "50Mi"
        securityContext:
          privileged: false
          capabilities:
            add: ["NET_RAW", "NET_ADMIN"]
      volumes:
      - name: cni-cfg
        hostPath:
          path: {{.CNIConfDir}}
          type: DirectoryOrCreate
      - name: xtables-lock
        hostPath:
          path: /run/xtables.lock
          type: FileOrCreate
      - name: lib-modules
        hostPath:
          path: /lib/modules

---
`))

// KindNet is the KindNet CNI manager
type KindNet struct {
	cc config.ClusterConfig
}

// String returns a string representation of this CNI
func (c KindNet) String() string {
	return "CNI"
}

// manifest returns a Kubernetes manifest for a CNI
func (c KindNet) manifest() (assets.CopyableFile, error) {
	input := &tmplInput{
		DefaultRoute: "0.0.0.0/0", // assumes IPv4
		PodCIDR:      DefaultPodCIDR,
		ImageName:    images.KindNet(c.cc.KubernetesConfig.ImageRepository),
		CNIConfDir:   DefaultConfDir,
	}

	b := bytes.Buffer{}
	if err := kindNetManifest.Execute(&b, input); err != nil {
		return nil, err
	}
	return manifestAsset(b.Bytes()), nil
}

// Apply enables the CNI
func (c KindNet) Apply(r Runner) error {
	// This is mostly applicable to the 'none' driver
	if _, err := r.RunCmd(exec.Command("stat", "/opt/cni/bin/portmap")); err != nil {
		return errors.Wrap(err, "required 'portmap' CNI plug-in not found")
	}

	m, err := c.manifest()
	if err != nil {
		return errors.Wrap(err, "manifest")
	}
	return applyManifest(c.cc, r, m)
}

// CIDR returns the default CIDR used by this CNI
func (c KindNet) CIDR() string {
	return DefaultPodCIDR
}
