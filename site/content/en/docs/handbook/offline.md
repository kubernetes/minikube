---
title: "Offline usage"
linkTitle: "Offline usage"
weight: 8
date: 2019-08-01
description: >
  Cache Rules Everything Around minikube
---

minikube has built-in support for caching downloaded resources into `$MINIKUBE_HOME/cache`. Here are the important file locations:

* `~/.minikube/cache` - Top-level folder
* `~/.minikube/cache/iso/<arch>` - VM ISO image. Typically updated once per major minikube release.
* `~/.minikube/cache/kic/<arch>` - Docker base image. Typically updated once per major minikube release.
* `~/.minikube/cache/images/<arch>` - Images used by Kubernetes, only exists if preload doesn't exist.
* `~/.minikube/cache/<os>/<arch>/<version>` - Kubernetes binaries, such as `kubeadm` and `kubelet`
* `~/.minikube/cache/preloaded-tarball` - Tarball of preloaded images to improve start time

## Kubernetes image cache

NOTE: the `none` driver caches images directly into Docker rather than a separate disk cache.

`minikube start` caches all required Kubernetes images by default. This default may be changed by setting `--cache-images=false`. These images are not displayed by the `minikube cache` command.

## Sharing the minikube cache

For offline use on other hosts, one can copy the contents of `~/.minikube/cache`.

```text
cache/linux/amd64/v1.26.1/kubectl
cache/kic/amd64/kicbase_v0.0.37@sha256_8bf7a0e8a062bc5e2b71d28b35bfa9cc862d9220e234e86176b3785f685d8b15.tar
cache/preloaded-tarball/preloaded-images-k8s-v18-v1.26.1-docker-overlay2-amd64.tar.lz4
cache/preloaded-tarball/preloaded-images-k8s-v18-v1.26.1-docker-overlay2-amd64.tar.lz4.checksum
```

If any of these files exist, minikube will use copy them into the VM directly rather than pulling them from the internet.
