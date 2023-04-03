---
title: "Running Minikube with a Specific Network"                                 
linkTitle: "Running Minikube with a Specific Network"
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
    --gateway=172.28.5.254 \
    br0
```

## Tutorial

Use the `--network` flag on `minikube start` to set the specific network.

**Note:** You cannot add a specific network to an existing cluster, you have to delete and recreate the cluster with the flag.

```
$ minikube start --driver=docker --network br0 --static-ip 172.28.5.199
😄  minikube v1.29.0 on Darwin 13.3 (arm64)
🎉  minikube 1.30.0 is available! Download it: https://github.com/kubernetes/minikube/releases/tag/v1.30.0
💡  To disable this notice, run: 'minikube config set WantUpdateNotification false'
✨  Using the docker driver based on user configuration
📌  Using Docker Desktop driver with root privileges
👍  Starting control plane node minikube in cluster minikube
🚜  Pulling base image ...
💾  Downloading Kubernetes v1.26.1 preload ...
    > preloaded-images-k8s-v18-v1...:  330.51 MiB / 330.51 MiB  100.00% 11.93 M
🔥  Creating docker container (CPUs=2, Memory=4000MB) ...
🐳  Preparing Kubernetes v1.26.1 on Docker 20.10.23 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔗  Configuring bridge CNI (Container Networking Interface) ...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🔎  Verifying Kubernetes components...
🌟  Enabled addons: storage-provisioner, default-storageclass
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default

$ docker inspect $CONTAINER_ID --format "{{json .NetworkSettings.Networks }}" | jq .
{
  "br0": {
    "IPAMConfig": {
      "IPv4Address": "172.28.5.199"
    },
    "Links": null,
    "Aliases": [
      "37ac060b925c",
      "minikube"
    ],
    "NetworkID": "c9dbf30c7d8df1ff998b1d4998d6c281649c2393c4acdbb4ec9ac33bbd82b2ad",
    "EndpointID": "d983457ba5a1ddc0b77be28ead21e665d345746ebb0a822aac8c82c625e91c00",
    "Gateway": "172.28.5.254",
    "IPAddress": "172.28.5.199",
    "IPPrefixLen": 16,
    ...
  }
}
```