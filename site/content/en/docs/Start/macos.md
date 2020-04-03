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

```
brew cask uninstall minikube

brew install minikube

```

If `which minikube` fails after installing with brew, you may need to uninstall the cask and reinstall by running:

You can't have a cask and a formula of the same thing installed since they will install the same files. If you uninstall the cask it will automatically link again. (and you only need one install anyway).

To fix this run:
`brew cask uninstall minikube` and then either `brew link minikube` or `brew reinstall minikube`.


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
{{% readfile file="/docs/Reference/Drivers/includes/docker_usage.inc" %}}
{{% /tab %}}
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
{{% tab "Podman (experimental)" %}}
{{% readfile file="/docs/Reference/Drivers/includes/podman_usage.inc" %}}
{{% /tab %}}

{{% /tabs %}}

{{% readfile file="/docs/Start/includes/post_install.inc" %}}
