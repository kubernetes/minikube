---
title: "CSI Driver and Volume Snapshots"
linkTitle: "CSI Driver and Volume Snapshots"
weight: 1
date: 2020-08-06
description: >
  CSI Driver and Volume Snapshots
---

## Overview

This tutorial explains how to set up the CSI Hostpath Driver in minikube and create volume snapshots.

## Prerequisites

- latest version of minikube
- kubernetes v1.20 or later

## What you’ll need

Support for volume snapshots in minikube is provided through the `volumesnapshots` addon. This addon provisions the required
CRDs and deploys the Volume Snapshot Controller. It is <b>disabled by default</b>.

Furthermore, the default storage provider in minikube does not implement the CSI interface and thus is NOT capable of creating/handling
volume snapshots. For that, you must first deploy a CSI driver. To make this step easy, minikube offers the `csi-hostpath-driver` addon,
which deploys the [CSI Hostpath Driver](https://github.com/kubernetes-csi/csi-driver-host-path). This addon is <b>disabled</b>
by default as well.

Thus, to utilize the volume snapshots functionality, you must:

1\) enable the `volumesnapshots` addon AND\
2a\) either enable the `csi-hostpath-driver` addon OR\
2b\) deploy your own CSI driver

You can enable/disable either of the above-mentioned addons using
```shell script
minikube addons enable [ADDON_NAME]
minikube addons disable [ADDON_NAME]
```

The `csi-hostpath-driver` addon deploys its required resources into the `kube-system` namespace and sets up a dedicated
storage class called `csi-hostpath-sc` that you need to reference in your PVCs. The driver itself is created under the
name `hostpath.csi.k8s.io`. Use this wherever necessary (e.g. snapshot class definitions).

Once both addons are enabled, you can create persistent volumes and snapshots using standard ways (for a quick test of
volume snapshots, you can find some example yaml files along with a step-by-step [here](https://kubernetes-csi.github.io/docs/snapshot-restore-feature.html)).
The driver stores all persistent volumes in the `/var/lib/csi-hostpath-data/` directory of minikube's host.

## Tutorial

In this tutorial, you use `volumesnapshots` addon(1) and `csi-hostpath-driver` addon(2a).

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">1</strong></span>Start your cluster</h2>

```shell
minikube start
```

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">2</strong></span>Enable addons</h2>

Enable `volumesnapshots` and `csi-hostpath-driver` addons:

```shell
minikube addons enable volumesnapshots
minikube addons enable csi-hostpath-driver
```

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">3</strong></span>Check volume snapshot class</h2>

When you create the volume snapshot, you have to register [Volume Snapshot Classes](https://kubernetes.io/docs/concepts/storage/volume-snapshot-classes/) to your cluster.
The default `VolumeSnapshotClass` called `csi-hostpath-snapclass` is already registered by `csi-hostpath-driver` addon.
You can check the `VolumeSnapshotClass` by the following command:

```shell
kubectl get volumesnapshotclasses
NAME                     DRIVER                DELETIONPOLICY   AGE
csi-hostpath-snapclass   hostpath.csi.k8s.io   Delete           10s
```

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">4</strong></span>Prepare persistent volume</h2>

Create persistent volume claim to create persistent volume dynamically:

```yaml
# example-pvc.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: csi-hostpath-sc
```

```shell
kubectl apply -f example-pvc.yaml
```

You can confirm that persistent volume is created by the following command:

```shell
kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM             STORAGECLASS      REASON   AGE
pvc-388c33e2-de56-475c-8dfd-4990d5f7a640   1Gi        RWO            Delete           Bound    default/csi-pvc   csi-hostpath-sc            60s
```

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">5</strong></span>Take a volume snapshot</h2>

You can take a volume snapshot for persistent volume claim:

```yaml
# example-csi-snapshot.yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: snapshot-demo
spec:
  volumeSnapshotClassName: csi-hostpath-snapclass
  source:
    persistentVolumeClaimName: csi-pvc
```

```shell
kubectl apply -f example-csi-snapshot.yaml
```

You could get volume snapshot. You can confirm your volume snapshot by the following command:

```shell
kubectl get volumesnapshot
NAME            READYTOUSE   SOURCEPVC   SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS            SNAPSHOTCONTENT                                    CREATIONTIME   AGE
snapshot-demo   true         csi-pvc                             1Gi           csi-hostpath-snapclass   snapcontent-19730fcb-c34a-4f1a-abf2-6c5a9808076b   5s             5s
```

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">6</strong></span>Restore from volume snapshot</h2>

You can restore persistent volume from your volume snapshot:

```yaml
# example-csi-restore.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-pvc-restore
spec:
  storageClassName: csi-hostpath-sc
  dataSource:
    name: snapshot-demo
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```

```shell
kubectl apply -f example-csi-restore.yaml
```

You can confirm that persistent volume claim is created from `VolumeSnapshot`:

```shell
kubectl get pvc
NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS      AGE
csi-pvc           Bound    pvc-388c33e2-de56-475c-8dfd-4990d5f7a640   1Gi        RWO            csi-hostpath-sc   23m
csi-pvc-restore   Bound    pvc-496bab30-9bd6-4abb-94e9-d2e9e1c8f210   1Gi        RWO            csi-hostpath-sc   26s
```
