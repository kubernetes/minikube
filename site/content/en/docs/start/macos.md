---
title: "macOS"
linkTitle: "macOS"
weight: 2
aliases:
  - /docs/start/macos/
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
{{% tab "Docker" %}}
{{% readfile file="/docs/drivers/includes/docker_usage.inc" %}}
{{% /tab %}}
{{% tab "Hyperkit" %}}
{{% readfile file="/docs/drivers/includes/hyperkit_usage.inc" %}}
{{% /tab %}}
{{% tab "VirtualBox" %}}
{{% readfile file="/docs/drivers/includes/virtualbox_usage.inc" %}}
{{% /tab %}}
{{% tab "Parallels" %}}
{{% readfile file="/docs/drivers/includes/parallels_usage.inc" %}}
{{% /tab %}}
{{% tab "VMware" %}}
{{% readfile file="/docs/drivers/includes/vmware_macos_usage.inc" %}}
{{% /tab %}}
{{% tab "Podman (experimental)" %}}
{{% readfile file="/docs/drivers/includes/podman_usage.inc" %}}
{{% /tab %}}

{{% /tabs %}}

{{% readfile file="/docs/start/includes/post_install.inc" %}}
