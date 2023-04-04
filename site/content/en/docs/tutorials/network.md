---
title: "Create a minikube cluster with a custom docker network"                                 
linkTitle: "Create a minikube cluster with a custom docker network"
weight: 1
date: 2023-04-04
--- 

## Overview

This tutorial will show you how to create a minikube cluster with a specific network.

## Prerequisites

- minikube v1.29.0 or higher
- Docker driver

## Createing a custom network

Setting a custom Docker network which is going be used by a new Minikube cluster.

```bash
$ docker network create \
    --driver=bridge \
    --subnet=172.28.0.0/16 \
    --ip-range=172.28.5.0/24 \
    --gateway=172.28.5.1 \
    br0
```

## Tutorial

Use the `--network` and `--nodes` flags on `minikube start` to run multi-node cluster with a specific network.

**Note:** You cannot add a specific network to an existing cluster, you have to delete and recreate the cluster with the flag.

```
$ minikube start --nodes 2 --driver=docker --network br0
😄  minikube v1.26.1 on Darwin 13.3 (arm64)
🎉  minikube 1.30.0 is available! Download it: https://github.com/kubernetes/minikube/releases/tag/v1.30.0
💡  To disable this notice, run: 'minikube config set WantUpdateNotification false'
✨  Using the docker driver based on user configuration
📌  Using Docker Desktop driver with root privileges
👍  Starting control plane node minikube in cluster minikube
🚜  Pulling base image ...
💾  Downloading Kubernetes v1.24.3 preload ...
    > preloaded-images-k8s-v18-v1...:  342.82 MiB / 342.82 MiB  100.00% 28.24 M
🔥  Creating docker container (CPUs=2, Memory=2200MB) ...
🐳  Preparing Kubernetes v1.24.3 on Docker 20.10.17 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔗  Configuring CNI (Container Networking Interface) ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass
👍  Starting worker node minikube-m02 in cluster minikube
🚜  Pulling base image ...
🔥  Creating docker container (CPUs=2, Memory=2200MB) ...
🌐  Found network options:
    ▪ NO_PROXY=172.28.5.2
🐳  Preparing Kubernetes v1.24.3 on Docker 20.10.17 ...
    ▪ env NO_PROXY=172.28.5.2
🔎  Verifying Kubernetes components...
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default
```

Make sure that multi-node cluster with a particular network has been successfully created and attached

```
$ k get nodes                                                                                                                                     ─╯
NAME           STATUS     ROLES           AGE   VERSION
minikube       Ready      control-plane   45s   v1.24.3
minikube-m02   NotReady   <none>          8s    v1.24.3

$ docker inspect $(docker ps --format="{{.ID}}" --filter="name=minikube") --format "{{json .NetworkSettings.Networks }}" | jq .
{
  "br0": {
    "IPAMConfig": {
      "IPv4Address": "172.28.5.3"
    },
    "Links": null,
    "Aliases": [
      "4f7602b63300",
      "minikube-m02"
    ],
    "NetworkID": "97eae3489064e142776a169b248f35fd62da9773283d3f6e63eaf0cd23ab1aee",
    "EndpointID": "7b772be8e7efba68d6db762579569108ea2bbab7cd8bf98f9e9111c9cc762239",
    "Gateway": "172.28.5.1",
    "IPAddress": "172.28.5.3",
    "IPPrefixLen": 16,
    ...
  }
}
{
  "br0": {
    "IPAMConfig": {
      "IPv4Address": "172.28.5.2"
    },
    "Links": null,
    "Aliases": [
      "be4ab7fca150",
      "minikube"
    ],
    "NetworkID": "97eae3489064e142776a169b248f35fd62da9773283d3f6e63eaf0cd23ab1aee",
    "EndpointID": "e431c5a2ccb92194697b7e7f41c6a742f40e7b943ad531aba4cbad3d6988d041",
    "Gateway": "172.28.5.1",
    "IPAddress": "172.28.5.2",
    "IPPrefixLen": 16,
    "IPv6Gateway": "",
    ...
  }
}
```