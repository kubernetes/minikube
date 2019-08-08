---
title: "Linux"
linkTitle: "Linux"
weight: 1
---

## Installation

{{% tabs %}}
{{% tab "Direct" %}}

Download and install minikube to /usr/local/bin:

```shell
 curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
   && sudo install minikube-linux-amd64 /usr/local/bin/minikube
```
{{% /tab %}}
{{% tab "Debian/Ubuntu (deb)" %}}

Download and install minikube:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_{{< latest >}}.deb \
 && sudo dpkg -i minikube_{{< latest >}}.deb
 ```

{{% /tab %}}

{{% tab "Fedora/Red Hat (rpm)" %}}

Download and install minikube:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-{{< latest >}}.rpm \
 && sudo rpm -ivh minikube-{{< latest >}}.rpm
 ```

{{% /tab %}}
{{% /tabs %}}

## Hypervisor Setup

Verify that your system has virtualization support enabled:

```shell
egrep -q 'vmx|svm' /proc/cpuinfo && echo yes || echo no
```

If the above command outputs "no":

- If you are running within a VM, your hypervisor does not allow nested virtualization. You will need to use the *None (bare-metal)* driver
- If you are running on a physical machine, ensure that your BIOS has hardware virtualization enabled

{{% tabs %}}

{{% tab "VirtualBox" %}}
{{% readfile file="/docs/Getting started/includes/virtualbox.md" %}}
{{% /tab %}}
{{% tab "KVM" %}}

### Prerequisites Installation

{{% readfile file="/docs/Reference/Drivers/includes/kvm2_prereqs_install.md" %}}

### Driver Installation

{{% readfile file="/docs/Reference/Drivers/includes/kvm2_driver_install.md" %}}

### Usage

```shell
minikube start --vm-driver=kvm2
```
To make kvm2 the default for future invocations, run:

```shell
minikube config set vm-driver kvm2
```

{{% /tab %}}
{{% tab "None (bare-metal)" %}}

If you are already running minikube from inside a VM, it is possible to skip the creation of an additional VM layer by using the `none` driver. 
This mode does come with additional requirements:

- docker
- systemd
- sudo access

```shell
sudo minikube start --vm-driver=none
```

Please see the [docs/reference/drivers/none](none driver) documentation for more information.
{{% /tab %}}
{{% /tabs %}}

{{% readfile file="/docs/Getting started/includes/post_install.md" %}}
