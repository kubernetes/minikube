---
title: "Setting a Static IP for a Cluster"                                 
linkTitle: "Setting a Static IP for a Cluster"
weight: 1
date: 2023-01-04
--- 

## Overview

This tutorial will show you how to create a minikube cluster with a static IP.

## Prerequisites

- minikube v1.29.0 or higher
- Docker or Podman driver

## Selecting a static IP

The static IP must be IPv4, private, and the last octet must be between 2-254 (X.X.X.2 - X.X.X.254).

Valid static IPs:<br>
10.0.0.2 - 10.255.255.254<br>
172.16.0.2 - 172.31.255.254<br>
192.168.0.2 - 192.168.255.254

## Tutorial

Use the `--static-ip` flag on `minikube start` to set the static IP.

**Note:** You cannot add a static IP to an existing cluster, you have to delete and recreate the cluster with the flag.

```
$ minikube start --driver docker --static-ip 192.168.200.200
😄  minikube v1.28.0 on Darwin 13.1 (arm64)
✨  Using the docker driver based on user configuration
📌  Using Docker Desktop driver with root privileges
👍  Starting control plane node minikube in cluster minikube
🚜  Pulling base image ...
🔥  Creating docker container (CPUs=2, Memory=4000MB) ...
🐳  Preparing Kubernetes v1.25.3 on Docker 20.10.21 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: default-storageclass
🏄  Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default

$ minikube ip
192.168.200.200
```
