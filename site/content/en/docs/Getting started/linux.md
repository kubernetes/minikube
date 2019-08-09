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
{{% readfile file="/docs/Getting started/_virtualbox.md" %}}
{{% /tab %}}
{{% tab "KVM" %}}

The KVM driver requires libvirt and qemu-kvm to be installed:

- Debian or Ubuntu 18.x: `sudo apt install libvirt-clients libvirt-daemon-system qemu-kvm`
- Ubuntu 16.x or older: `sudo apt install libvirt-bin libvirt-daemon-system qemu-kvm`
- Fedora/CentOS/RHEL: `sudo yum install libvirt libvirt-daemon-kvm qemu-kvm`
- openSUSE/SLES: `sudo zypper install libvirt qemu-kvm`

Additionally, The KVM driver requires an additional binary to be installed:

```shell
 curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-kvm2 \
  && sudo install docker-machine-driver-kvm2 /usr/local/bin/
```

### Validate libvirt

Before trying minikube, assert that libvirt is in a healthy state:

```shell
virt-host-validate
```

If you see any errors, stop now and consult your distributions documentation on configuring libvirt.

### Using the kvm2 driver

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

{{% readfile file="/docs/Getting started/_post_install.md" %}}
