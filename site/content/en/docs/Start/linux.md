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

## Hypervisor Setup

Verify that your system has virtualization support enabled:

```shell
egrep -q 'vmx|svm' /proc/cpuinfo && echo yes || echo no
```

If the above command outputs "no":

- If you are running within a VM, your hypervisor does not allow nested virtualization. You will need to use the *None (bare-metal)* driver
- If you are running on a physical machine, ensure that your BIOS has hardware virtualization enabled

{{% tabs %}}

{{% tab "KVM" %}}
{{% readfile file="/docs/Reference/Drivers/includes/kvm2_usage.inc" %}}
{{% /tab %}}
{{% tab "VirtualBox" %}}
{{% readfile file="/docs/Reference/Drivers/includes/virtualbox_usage.inc" %}}
{{% /tab %}}
{{% tab "None (bare-metal)" %}}
If you are already running minikube from inside a VM, it is possible to skip the creation of an additional VM layer by using the `none` driver.

{{% readfile file="/docs/Reference/Drivers/includes/none_usage.inc" %}}
{{% /tab %}}
{{% /tabs %}}

{{% readfile file="/docs/Start/includes/post_install.inc" %}}
