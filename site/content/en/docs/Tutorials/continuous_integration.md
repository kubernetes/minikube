---
title: "Continuous Integration"
linkTitle: "Continuous Integration"
weight: 1
date: 2018-01-02
description: >
  Using minikube for Continuous Integration
---

## Overview

Most continuous integration environments are already running inside a VM, and may not support nested virtualization. The `none` driver was designed for this use case. or you could alternatively use the [Docker](https://minikube.sigs.k8s.io/docs/reference/drivers/docker).

## Prerequisites

- VM running a systemd based Linux distribution

## using none driver

 Here is an example, that runs minikube from a non-root user, and ensures that the latest stable kubectl is installed:

```shell
curl -Lo minikube \
  https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
  && sudo install minikube /usr/local/bin/

kv=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
curl -Lo kubectl \
  https://storage.googleapis.com/kubernetes-release/release/$kv/bin/linux/amd64/kubectl \
  && sudo install kubectl /usr/local/bin/

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
export KUBECONFIG=$HOME/.kube/config

mkdir -p $HOME/.kube $HOME/.minikube
touch $KUBECONFIG

sudo -E minikube start --driver=none
```

##  Alternative ways

you could alternatively use minikube's container drivers such as [Docker](https://minikube.sigs.k8s.io/docs/reference/drivers/docker) or [Podman](https://minikube.sigs.k8s.io/docs/reference/drivers/podman).