---
title: "Continuous Integration"
weight: 1
description: >
  Using minikube for Continuous Integration
---

## Overview

Most continuous integration environments are already running inside a VM, and may not support nested virtualization.

The `docker` driver was designed for this use case, as well as the older `none` driver.

## Example

 Here is an example, that runs minikube from a non-root user, and ensures that the latest stable kubectl is installed:

```shell
curl -LO \
  https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
  && install minikube-linux-amd64 /tmp/
  
kv=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
curl -LO \
  https://storage.googleapis.com/kubernetes-release/release/$kv/bin/linux/amd64/kubectl \
  && install kubectl /tmp/

export MINIKUBE_WANTUPDATENOTIFICATION=false
/tmp/minikube-linux-amd64 start --driver=docker
```
