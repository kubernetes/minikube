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

If you have the [Brew Package Manager](https://brew.sh/) installed, this will download and install minikube to /usr/local/bin:

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
{{% readfile file="/docs/Getting started/_virtualbox.md" %}}
{{% /tab %}}
{{% tab "Hyperkit" %}}

The easiest way to install hyperkit is using the [Brew Package Manager](https://brew.sh/):

```shell
brew install hyperkit
```

To use minikube and hyperkit reliably, please download and install our improved fork of the driver:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit \
  && sudo install docker-machine-driver-hyperkit /usr/local/bin/
```

This driver requires root privileges to create and destroy network interfaces:

```shell
sudo chown root:wheel /usr/local/bin 
docker-machine-driver-hyperkit \
&& sudo chmod u+s /usr/local/bin/docker-machine-driver-hyperkit
```

### Using the hyperkit driver

```shell
minikube start --vm-driver=hyperkit
```

To make hyperkit the default for future invocations:

```shell
minikube config set vm-driver hyperkit
```

{{% /tab %}}

{{% tab "Parallels" %}}

Start minikube with Parallels support using:

```shell
minikube start --vm-driver=parallels
```

To make parallels the default for future invocations:

```shell
minikube config set vm-driver parallels
```
{{% /tab %}}

{{% tab "VMware Fusion" %}}

Start minikube with VMware Fusion support using:

```shell
minikube start --vm-driver=vmwarefusion
```

To make vmwarefusion the default for future invocations:

```shell
minikube config set vm-driver vmwarefusion
```
{{% /tab %}}

{{% /tabs %}}

{{% readfile file="/docs/Getting started/_post_install.md" %}}