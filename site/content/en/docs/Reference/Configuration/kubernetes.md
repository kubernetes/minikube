---
title: "Kubernetes"
linkTitle: "Kubernetes"
weight: 3
date: 2019-08-01
description: >
  Kubernetes configuration reference
---

minikube allows users to configure the Kubernetes components with arbitrary values. To use this feature, you can use the `--extra-config` flag on the `minikube start` command.

This flag is repeated, so you can pass it several times with several different values to set multiple options.

## Selecting a Kubernetes version

By default, minikube installs the latest stable version of Kubernetes that was available at the time of the minikube release. You may select a different Kubernetes release by using the `--kubernetes-version` flag, for example:

`minikube start --kubernetes-version=v1.11.10`
  
If you omit this flag, minikube will upgrade your cluster to the default version. If you would like to pin to a specific Kubernetes version across clusters, restarts, and upgrades to minikube, use:

`minikube config set kubernetes-version v1.11.0`

minikube follows the [Kubernetes Version and Version Skew Support Policy](https://kubernetes.io/docs/setup/version-skew-policy/), so we guarantee support for the latest build for the last 3 minor Kubernetes releases. When practical, minikube aims for the last 6 minor releases so that users can emulate legacy environments.

As of September 2019, this means that minikube supports and actively tests against the latest builds of:

* v1.16 (default)
* v1.15
* v1.14
* v1.13
* v1.12
* v1.11 (best effort)

For more up to date information, see `OldestKubernetesVersion` and `NewestKubernetesVersion` in [constants.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/constants/constants.go)

## Modifying Kubernetes defaults

The kubeadm bootstrapper can be configured by the `--extra-config` flag on the `minikube start` command.  It takes a string of the form `component.key=value` where `component` is one of the strings

* kubeadm
* kubelet
* apiserver
* controller-manager
* scheduler

and `key=value` is a flag=value pair for the component being configured.  For example,

```shell
minikube start --extra-config=apiserver.v=10 --extra-config=kubelet.max-pods=100
```

For instance, to allow Kubernetes to launch on an unsupported Docker release:

```shell
minikube start --extra-config=kubeadm.ignore-preflight-errors=SystemVerification
```
