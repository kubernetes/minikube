---
title: "kvm2"
weight: 2
description: >
  Linux KVM (Kernel-based Virtual Machine) driver
aliases:
    - /docs/reference/drivers/kvm2
---


## Overview

[KVM (Kernel-based Virtual Machine)](https://www.linux-kvm.org/page/Main_Page) is a full virtualization solution for Linux on x86 hardware containing virtualization extensions. To work with KVM, minikube uses the [libvirt virtualization API](https://libvirt.org/)

{{% readfile file="/docs/drivers/includes/kvm2_usage.inc" %}}

## Check virtualization support

{{% readfile file="/docs/drivers/includes/check_virtualization_linux.inc" %}}

## Special features

The `minikube start` command supports 3 additional kvm specific flags:

* **`--gpu`**: Enable experimental NVIDIA GPU support in minikube
* **`--hidden`**: Hide the hypervisor signature from the guest in minikube
* **`--kvm-network`**:  The KVM network name
* **`--kvm-qemu-uri`**: The KVM qemu uri, defaults to qemu:///system

## Issues

* `minikube` will repeatedly ask for the root password if user is not in the correct `libvirt` group [#3467](https://github.com/kubernetes/minikube/issues/3467)
* `Machine didn't return an IP after 120 seconds` when firewall prevents VM network access [#3566](https://github.com/kubernetes/minikube/issues/3566)
* `unable to set user and group to '65534:992` when `dynamic ownership = 1` in `qemu.conf` [#4467](https://github.com/kubernetes/minikube/issues/4467)
* KVM VM's cannot be used simultaneously with VirtualBox  [#4913](https://github.com/kubernetes/minikube/issues/4913)
* On some distributions, libvirt bridge networking may fail until the host reboots

Also see [co/kvm2 open issues](https://github.com/kubernetes/minikube/labels/co%2Fkvm2)

### Nested Virtulization

If you are running KVM in a nested virtualization environment ensure your config the kernel modules correctly follow either [this](https://stafwag.github.io/blog/blog/2018/06/04/nested-virtualization-in-kvm/) or [this](https://computingforgeeks.com/how-to-install-kvm-virtualization-on-debian/) tutorial.

## Troubleshooting
* Run `virt-host-validate` and check for the suggestions.
* Run `minikube start --alsologtostderr -v=7` to debug crashes
* Run `docker-machine-driver-kvm2 version` to verify the kvm2 driver executes properly.
* Read [How to debug Virtualization problems](https://fedoraproject.org/wiki/How_to_debug_Virtualization_problems)
