---
linkTitle: "gVisor (DEPRECATED)"
title: "Releasing a gVisor image (DEPRECATED)"
date: 2019-09-25
weight: 10
---

> [!WARNING]
> **This guide is deprecated.**
>
> This document describes the outdated manual build process and refers to the legacy, internal Google infrastructure (`gcr.io/k8s-minikube`). For instructions on building, staging, and promoting helper/addon images (including gVisor) using modern Google Cloud Build and Prow workflows, please refer to the official [Publishing Container Images to registry.k8s.io](../../building/images/) guide.





## Prerequisites

* Credentials for `gcr.io/k8s-minikube`
* Docker
* Gcloud

## Background

gVisor support within minikube requires a special Docker image to be generated. After merging changes to `cmd/gvisor` or `pkg/gvisor`, this image will need to be updated.

The image is located at `gcr.io/k8s-minikube/gvisor-addon`

## Why is this image required?

`gvisor` requires changes to the guest VM in order to function. The `addons` feature in minikube does not normally allow for this, so to workaround it, a custom docker image is launched, containing a binary that makes the changes.

## What does the image do?

- Creates log directories
- Downloads and installs the latest stable `gvisor-containerd-shim` release
- Updates the containerd configuration
- Restarts containerd and rpc-statd

## Updating the gVisor image

```shell
make push-gvisor-addon-image
```
