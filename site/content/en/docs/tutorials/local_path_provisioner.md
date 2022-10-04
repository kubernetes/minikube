---
title: "Using Local Path Provisioner"
linkTitle: "Using Local Path Provisioner"
weight: 1
date: 2022-10-05
description: >
  Using Local Path Provisioner
---

## Overview

[Local Path Provisioner](https://github.com/rancher/local-path-provisioner), provides a way for the Kubernetes users to utilize the local storage in each node. It supports multi-node setups. This tutorial will show you how to setup local-path-prvisioner on two node minikube cluster.

## Prerequisites

- Minikube version higher than v1.27.0
- kubectl

## Tutorial

- Start a cluster with 2 nodes:

```shell
$ minikube start -n 2
```

- Enable `storage-provisioner-rancher` addon:

```
$ minikube addons enable storage-provisioner-rancher
```

- You should be able to see Pod in the `local-path-storage` namespace:

```
$ kubectl get pods -n local-path-storage
NAME                                     READY   STATUS    RESTARTS   AGE
local-path-provisioner-7f58b4649-hcbk9   1/1     Running   0          38s
```

- The `local-path` StorageClass should be marked as `default`:

```
$ kubectl get sc
NAME                   PROVISIONER                RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path      Delete          WaitForFirstConsumer   false                  107s
standard               k8s.io/minikube-hostpath   Delete          Immediate              false                  4m27s
```

- The following `yaml` creates PVC and Pod that creates file with content on second node (minikube-m02):

```
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 64Mi
---
apiVersion: v1
kind: Pod
metadata:
  name: test-local-path
spec:
  restartPolicy: OnFailure
  nodeSelector:
    "kubernetes.io/hostname": "minikube-m02"
  containers:
    - name: busybox
      image: busybox:stable
      command: ["sh", "-c", "echo 'local-path-provisioner' > /test/file1"]
      volumeMounts:
      - name: data
        mountPath: /test
  volumes:
    - name: data
      persistentVolumeClaim:
        claimName: test-pvc
```

```
$ kubectl get pvc
NAME       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
test-pvc   Bound    pvc-f07e253b-fea7-433a-b0ac-1bcea3f77076   64Mi       RWO            local-path     5m19s
```

```
$ kubectl get pods -o wide
NAME              READY   STATUS      RESTARTS   AGE     IP           NODE           NOMINATED NODE   READINESS GATES
test-local-path   0/1     Completed   0          5m19s   10.244.1.5   minikube-m02   <none>           <none>
```

- On the second node we are able to see created file with content `local-path-provisioner`:

```
$ minikube ssh -n minikube-m02 "cat /opt/local-path-provisioner/pvc-f07e253b-fea7-433a-b0ac-1bcea3f77076_default_test-pvc/file1"
local-path-provisioner
```
