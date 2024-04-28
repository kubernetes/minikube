---
title: "virtualbox"
weight: 5
aliases:
  - /docs/reference/drivers/virtualbox
---

## Overview

VirtualBox is minikube's original driver. It may not provide the fastest start-up time, but it is the most stable driver available for users of Microsoft Windows Home.

{{% readfile file="/docs/drivers/includes/virtualbox_usage.inc" %}}

## Special features

minikube start supports some VirtualBox specific flags:

* **`--host-only-cidr`**: The CIDR to be used for the minikube VM (default "192.168.59.1/24")
  * On Linux, Mac OS X and Oracle Solaris with VirtualBox >= 6.1.28, [only IP addresses in the 192.168.56.0/21 range are allowed for host-only networking by default](https://www.virtualbox.org/manual/ch06.html#network_hostonly). Passing a disallowed value to `--host-only-cidr` will result in a VirtualBox access denied error: `VBoxManage: error: Code E_ACCESSDENIED (0x80070005) - Access denied (extended info not available)`.
* **`--no-vtx-check`**: Disable checking for the availability of hardware virtualization

## Issues

* [Full list of open 'virtualbox' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fvirtualbox)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
* If you experience slow network performance with the VirtualBox driver, changing the Network Interface Card (NIC) type may improve speed. Use the following command to start minikube with the AMD PCNet FAST III (Am79C973) for both NAT and host-only network interfaces:

    ```shell
    minikube start --vm-driver=virtualbox --nat-nic-type=Am79C973 --host-only-nic-type=Am79C973
    ```
