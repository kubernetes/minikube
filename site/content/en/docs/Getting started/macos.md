---
title: "macOS"
linkTitle: "macOS"
weight: 2
description: >
  How to install and start minikube on macOS.
---

### Prerequisites

* macOS 10.12 (Sierra)
* A hypervisor such as Hyperkit, Parallels, VirtualBox, or VMware Fusion

### Installation

{{% tabs %}}
{{% tab "Brew" %}}

Download and install minikube to /usr/local/bin:

```shell
brew install minikube
```

{{% /tab %}}
{{% tab "Manual" %}}

Download and install minikube to /usr/local/bin:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64 \
  && sudo install minikube-darwin-amd64 /usr/local/bin/minikube
```
{{% /tab %}}

{{% /tabs %}}

## Hypervisor Setup

{{% tabs %}}
{{% tab "VirtualBox" %}}
{{% readfile file="/docs/Getting started/_virtualbox.md" %}}
{{% /tab %}}
{{% tab "Hyperkit" %}}

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit \
  && sudo install docker-machine-driver-hyperkit /usr/local/bin/
sudo chown root:wheel /usr/local/bin/docker-machine-driver-hyperkit
sudo chmod u+s /usr/local/bin/docker-machine-driver-hyperkit
```

{{% /tab %}}

{{% tab "Parallels" %}}
{{% /tab %}}

{{% tab "VMWare Fusion" %}}
{{% /tab %}}

{{% /tabs %}}
