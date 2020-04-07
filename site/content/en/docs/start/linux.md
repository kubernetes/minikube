---
title: "Linux"
linkTitle: "Linux"
weight: 1
aliases:
  - /docs/start/linux/
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
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_{{< latest >}}-0_amd64.deb \
 && sudo dpkg -i minikube_{{< latest >}}-0_amd64.deb
 ```

{{% /tab %}}

{{% tab "Fedora/Red Hat (rpm)" %}}

Download and install minikube:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-{{< latest >}}-0.x86_64.rpm \
 && sudo rpm -ivh minikube-{{< latest >}}-0.x86_64.rpm
 ```

{{% /tab %}}
{{% /tabs %}}

## Driver Setup

{{% tabs %}}
{{% tab "Docker" %}}
## Check container support
{{% readfile file="/docs/drivers/includes/docker_usage.inc" %}}
{{% /tab %}}

{{% tab "KVM" %}}
## Check virtualization support
{{% readfile file="/docs/drivers/includes/check_virtualization_linux.inc" %}}

{{% readfile file="/docs/drivers/includes/kvm2_usage.inc" %}}
{{% /tab %}}
{{% tab "VirtualBox" %}}
## Check virtualization support
{{% readfile file="/docs/drivers/includes/check_virtualization_linux.inc" %}}

{{% readfile file="/docs/drivers/includes/virtualbox_usage.inc" %}}
{{% /tab %}}
{{% tab "None (bare-metal)" %}}
## Check baremetal support
{{% readfile file="/docs/drivers/includes/check_baremetal_linux.inc" %}}

If you are already running minikube from inside a VM, it is possible to skip the creation of an additional VM layer by using the `none` driver.

{{% readfile file="/docs/drivers/includes/none_usage.inc" %}}
{{% /tab %}}
{{% tab "Podman (experimental)" %}}
{{% readfile file="/docs/drivers/includes/podman_usage.inc" %}}
{{% /tab %}}


{{% /tabs %}}

{{% readfile file="/docs/start/includes/post_install.inc" %}}
