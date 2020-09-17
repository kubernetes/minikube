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

## Tutorial

Support for volume snapshots in minikube is provided through the `volumesnapshots` addon. This addon provisions the required
CRDs and deploys the Volume Snapshot Controller. It is <b>disabled by default</b>.

Furthermore, the default storage provider in minikube does not implement the CSI interface and thus is NOT capable of creating/handling
volume snapshots. For that, you must first deploy a CSI driver. To make this step easy, minikube offers the `csi-hostpath-driver` addon,
which deploys the [CSI Hostpath Driver](https://github.com/kubernetes-csi/csi-driver-host-path). This addon is <b>disabled</b> 
by default as well.

Thus, to utilize the volume snapshots functionality, you must:

1\) enable the `volumesnapshots` addon AND\
2a\) either enable the `csi-hostpth-driver` addon OR\
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