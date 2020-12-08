---
title: "Using Minikube with Pod Security Policies"
linkTitle: "Using Minikube with Pod Security Policies"
weight: 1
date: 2019-11-24
description: >
  Using Minikube with Pod Security Policies
---

## Overview

This tutorial explains how to start minikube with Pod Security Policies (PSP) enabled.

## Prerequisites

- Minikube 1.11.1 with Kubernetes 1.16.x or higher

## Tutorial

Start minikube with the `PodSecurityPolicy` admission controller and the
`pod-security-policy` addon enabled.

```shell
minikube start --extra-config=apiserver.enable-admission-plugins=PodSecurityPolicy --addons=pod-security-policy
```

The `pod-security-policy` addon must be enabled along with the admission
controller to prevent issues during bootstrap.

## Older versions of minikube

Older versions of minikube do not ship with the `pod-security-policy` addon, so
the policies that addon enables must be separately applied to the cluster.

## Minikube 1.5.2 through 1.6.2

Before starting minikube, you need to give it the PSP YAMLs in order to allow minikube to bootstrap.

Create the directory:
```shell
mkdir -p ~/.minikube/files/etc/kubernetes/addons
```

Copy the YAML below into this file: `~/.minikube/files/etc/kubernetes/addons/psp.yaml`

Now start minikube:
```shell
minikube start --extra-config=apiserver.enable-admission-plugins=PodSecurityPolicy
```

```yaml
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: privileged
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: "*"
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
spec:
  privileged: true
  allowPrivilegeEscalation: true
  allowedCapabilities:
  - "*"
  volumes:
  - "*"
  hostNetwork: true
  hostPorts:
  - min: 0
    max: 65535
  hostIPC: true
  hostPID: true
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: restricted
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'MustRunAs'
    ranges:
      # Forbid adding the root group.
      - min: 1
        max: 65535
  fsGroup:
    rule: 'MustRunAs'
    ranges:
      # Forbid adding the root group.
      - min: 1
        max: 65535
  readOnlyRootFilesystem: false
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: psp:privileged
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
rules:
- apiGroups: ['policy']
  resources: ['podsecuritypolicies']
  verbs:     ['use']
  resourceNames:
  - privileged
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: psp:restricted
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
rules:
- apiGroups: ['policy']
  resources: ['podsecuritypolicies']
  verbs:     ['use']
  resourceNames:
  - restricted
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: default:restricted
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:restricted
subjects:
- kind: Group
  name: system:authenticated
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: default:privileged
  namespace: kube-system
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:privileged
subjects:
- kind: Group
  name: system:masters
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: system:nodes
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: system:serviceaccounts:kube-system
  apiGroup: rbac.authorization.k8s.io
```

### Minikube between 1.6.2 and 1.11.1

With minikube versions greater than 1.6.2 and less than 1.11.1, the YAML files
shown above will not be automatically applied to the cluster. You may have
errors during bootstrap of the cluster if the admission controller is enabled.

To use Pod Security Policies with these versions of minikube, first start a
cluster without the `PodSecurityPolicy` admission controller enabled.

Next, apply the YAML shown above to the cluster.

Finally, stop the cluster and then restart it with the admission controller
enabled.

```shell
minikube start
kubectl apply -f /path/to/psp.yaml
minikube stop
minikube start --extra-config=apiserver.enable-admission-plugins=PodSecurityPolicy
```
