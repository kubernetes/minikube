---
title: "Getting Started"
linkTitle: "Getting Started"
weight: 2
description: >
  How to install and start minikube.
---

{{% pageinfo %}}
This is a new page. As an alternative, see [Kubernetes: minikube installation guide](https://kubernetes.io/docs/tasks/tools/install-minikube/).
{{% /pageinfo %}}

{{% tabs %}}
{{% tab "Linux" %}}

### Prerequisites

Verify that your BIOS and kernel has virtualization support enabled:

```shell
egrep -q 'vmx|svm' /proc/cpuinfo && echo yes || echo no
```

### Installation

```shell
 curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
   && sudo install minikube-linux-amd64 /usr/local/bin/minikube
 curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-kvm2 \
  && sudo install docker-machine-driver-kvm2 /usr/local/bin/
```

## 

{{% /tab %}}

{{% tab "macOS" %}}

### Prerequisites

* macOS 10.12 (Sierra)
* A hypervisor such as VirtualBox, VMWare, or hyperkit

### Installation

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64 \
  && sudo install minikube-darwin-amd64 /usr/local/bin/minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit \
  && sudo install docker-machine-driver-hyperkit /usr/local/bin/
sudo chown root:wheel /usr/local/bin/docker-machine-driver-hyperkit
sudo chmod u+s /usr/local/bin/docker-machine-driver-hyperkit
```

{{% /tab %}}

{{% tab "Windows" %}}

### Prerequisites

  * A hypervisor, such as VirtualBox (recommended) or HyperV
  * VT-x/AMD-v virtualization must be enabled in BIOS

### Installation

Download and run the [installer](https://storage.googleapis.com/minikube/releases/latest/minikube-installer.exe)

{{% /tab %}}

{{% /tabs %}}

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
