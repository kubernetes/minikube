---
title: "Windows"
linkTitle: "Windows"
weight: 3
description: >
  How to install and start minikube on Windows.
---

### Prerequisites

  * A hypervisor, such as VirtualBox (recommended) or HyperV
  * VT-x/AMD-v virtualization must be enabled in BIOS

### Installation

Download and run the [installer](https://storage.googleapis.com/minikube/releases/latest/minikube-installer.exe)

## Try it out!

Start your first Kubernetes cluster:

```shell
minikube start
```

If `kubectl` is installed, this will show a status of all of the pods:

```shell
kubectl get po -A
```

## Further Setup

Once you have picked a hypervisor, you may set it as the default for future invocations:

```shell
minikube config set vm-driver <driver>
```

minikube only allocates a 2GB of RAM to Kubernetes, which is only enough for basic deployments. If you run into stability issues, increase this value if your system has the resources available:

```shell
minikube config set memory 4096
```
