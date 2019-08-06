---
title: "kvm2"
linkTitle: "kvm2"
weight: 1
date: 2017-01-05
date: 2018-08-05
description: >
  Linux KVM (Kernel-based Virtual Machine Driver
---

## Overview

[KVM (Kernel-based Virtual Machine)](https://www.linux-kvm.org/page/Main_Page) is a full virtualization solution for Linux on x86 hardware containing virtualization extensions. To work with KVM, minikube uses the [libvirt virtualization API](https://libvirt.org/)

## Requirements

- libvirt v1.3.1 or higher
- qemu-kvm v2.0 or higher

## Prerequisites Installation

{{% readfile file="/docs/Reference/Drivers/_kvm2_prereqs_install.md" %}}

## Driver Installation

{{% readfile file="/docs/Reference/Drivers/_kvm2_driver_install.md" %}}

## Using the kvm2 driver

```shell
minikube start --vm-driver=kvm2
```
To make kvm2 the default for future invocations, run:

```shell
minikube config set vm-driver kvm2
```

## Driver Differences

The `minikube start` command supports 3 additional kvm specific flags:

* **\--gpu**: Enable experimental NVIDIA GPU support in minikube
* **\--hidden**: Hide the hypervisor signature from the guest in minikube
* **\--kvm-network**:  The KVM network name

## Known Issues

* `minikube` will repeatedly for root password if user is not in the correct `libvirt` group [#3467](https://github.com/kubernetes/minikube/issues/3467)
* `Machine didn't return an IP after 120 seconds` when firewall prevents VM network access [#3566](https://github.com/kubernetes/minikube/issues/3566)
* `unable to set user and group to '65534:992` when `dynamic ownership = 1` in `qemu.conf` [#4467](https://github.com/kubernetes/minikube/issues/4467) 
* KVM VM's cannot be used simultaneously with VirtualBox  [#4913](https://github.com/kubernetes/minikube/issues/4913)
* On some distributions, libvirt bridge networking may fail until the host reboots

Also see [co/kvm2 open issues](https://github.com/kubernetes/minikube/labels/co%2Fkvm2)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes related to kvm
* Run `docker-machine-kvm2 version` to verify the kvm2 driver executes properly.
* Read [How to debug Virtualization problems](https://fedoraproject.org/wiki/How_to_debug_Virtualization_problems)
 