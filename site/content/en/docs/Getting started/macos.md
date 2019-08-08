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

## Hypervisor Setup

{{% tabs %}}
{{% tab "VirtualBox" %}}
{{% readfile file="/docs/Reference/Drivers/includes/virtualbox_usage.md" %}}
{{% /tab %}}
{{% tab "Hyperkit" %}}
{{% readfile file="/docs/Reference/Drivers/includes/hyperkit_usage.md" %}}
{{% /tab %}}
{{% tab "Parallels" %}}
{{% readfile file="/docs/Reference/Drivers/includes/parallels_usage.md" %}}
{{% /tab %}}
{{% tab "VMware" %}}
{{% readfile file="/docs/Reference/Drivers/includes/vmware_macos_usage.md" %}}
{{% /tab %}}

{{% /tabs %}}

{{% readfile file="/docs/Getting started/includes/post_install.md" %}}