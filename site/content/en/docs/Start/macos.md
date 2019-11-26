---
title: "macOS"
linkTitle: "macOS"
weight: 2
---

### Prerequisites

* macOS 10.12 (Sierra)
* A hypervisor such as Hyperkit, Parallels, VirtualBox, or VMware Fusion

### Installation

{{% tabs %}}
{{% tab "Brew" %}}

If the [Brew Package Manager](https://brew.sh/) is installed, use it to download and install minikube:

```shell
brew install minikube
```

{{% /tab %}}
{{% tab "Direct" %}}

Download and install minikube to /usr/local/bin:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64 \
  && sudo install minikube-darwin-amd64 /usr/local/bin/minikube
```
{{% /tab %}}
{{% /tabs %}}

### Upgrading minikube

{{% tabs %}}
{{% tab "Brew" %}}

If the [Brew Package Manager](https://brew.sh/) is installed, use it to download and upgrade minikube:

```shell
brew update
brew upgrade minikube
```

{{% /tab %}}
{{% /tabs %}}

## Hypervisor Setup

{{% tabs %}}
{{% tab "Hyperkit" %}}
{{% readfile file="/docs/Reference/Drivers/includes/hyperkit_usage.inc" %}}
{{% /tab %}}
{{% tab "VirtualBox" %}}
{{% readfile file="/docs/Reference/Drivers/includes/virtualbox_usage.inc" %}}
{{% /tab %}}
{{% tab "Parallels" %}}
{{% readfile file="/docs/Reference/Drivers/includes/parallels_usage.inc" %}}
{{% /tab %}}
{{% tab "VMware" %}}
{{% readfile file="/docs/Reference/Drivers/includes/vmware_macos_usage.inc" %}}
{{% /tab %}}

{{% /tabs %}}

{{% readfile file="/docs/Start/includes/post_install.inc" %}}
