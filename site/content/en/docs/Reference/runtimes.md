---
title: "Container Runtimes"
linkTitle: "Container Runtimes"
weight: 6
date: 2019-08-01
description: >
  Available container runtimes
---

### Docker

The default container runtime in minikube is Docker. You can select it explicitly by using:

```shell
minikube start --container-runtime=docker
```

### CRI-O

To use [CRI-O](https://github.com/kubernetes-sigs/cri-o):

```shell
minikube start --container-runtime=cri-o
```

## containerd

To use [containerd](https://github.com/containerd/containerd):

```shell
minikube start --container-runtime=containerd
```

## gvisor

To use [gvisor](https://gvisor.dev):

```shell
minikube start --container-runtime=containerd
minikube addons enable gvisor
```

## Kata

Native support for [Kata containers](https://katacontainers.io) is a work-in-progress. See [#4347](https://github.com/kubernetes/minikube/issues/4347) for details.

In the mean time, it's possible to make Kata containers work within minikube using a bit of [elbow grease](https://gist.github.com/olberger/0413cfb0769dcdc34c83788ced583fa9).
