---
title: "qemu"
weight: 3
description: >
  QEMU driver
aliases:
    - /docs/reference/drivers/qemu
---

## Overview

The `qemu` driver users QEMU (system) for VM creation.

<https://www.qemu.org/>

## Usage

To start minikube with the qemu driver:

```shell
minikube start --driver=qemu
```

## Special features

minikube start supports some qemu specific flags:

* **`--qemu-firmware-path`**: The path to the firmware image to be used.
  * Note: while the flag should override the config, if the flag does not take effect try running `minikube delete`.
  * MacPorts: if you are installing [minikube](https://ports.macports.org/port/minikube/) and [qemu](https://ports.macports.org/port/qemu/) via MacPorts on a Mac with M1, use the following flag: `--qemu-firmware-path=/opt/local/share/qemu/edk2-aarch64-code.fd`

## Networking

The QEMU driver has two networking options, `user` & `socket_vmnet`.

{{% tabs %}}
{{% tab "user - limited functionality" %}}
The `user` network is not a dedicated network, it doesn't support some networking commands such as `minikube service` and `minikube tunnel`, and its IP address is not reachable from the host.
{{% /tab %}}
{{% tab "socket_vmnet - experimental/needs installation" %}}
##### Requirements

Requires macOS 10.15 or later and socket_vmnet.

[lima-vm/socket_vmnet](https://github.com/lima-vm/socket_vmnet) install instructions:
```shell
git clone https://github.com/lima-vm/socket_vmnet.git && cd socket_vmnet
sudo make PREFIX=/opt/socket_vmnet install
```

##### Usage
```shell
minikube start --driver qemu --network socket_vmnet
```

The `socket_vmnet` network is a dedicated network and supports the `minikube service` and `minikube tunnel` commands.
{{% /tab %}}
{{% /tabs %}}

## Known Issues

### 1. Start stuck with `user` network on corp machine or custom DNS

When using the `user` network (default) the guest uses **only** the first `nameserver` entry in the hosts `/etc/resolv.conf` for DNS lookup. If your first `nameserver` entry is a corporate/internal DNS it's likely it will cause an issue. If you see the warning `❗ This VM is having trouble accessing https://registry.k8s.io` on `minikube start` you are likely being affected by this. This may prevent your cluster from starting entirely and you won't be able to pull remote images. More details can be found at: [#15021](https://github.com/kubernetes/minikube/issues/15021)

##### Workarounds:
1. If possible, reorder your `/etc/resolv.conf` to have a general `nameserver` entry first (eg. `8.8.8.8`) and reboot your machine.
2. Use `--network=socket_vmnet`

### 2. `/var/db/dhcpd_leases` errors

If you're seeing errors related to `/var/db/dhcpd_leases` we recommend the following:
1. Uninstall `socket_vmnet`:
```shell
cd socket_vmnet
sudo make uninstll
sudo rm /var/run/socket_vmnet
```
2. Reboot
3. Reinsitall `socket_vmnet`:
```shell
cd socket_vmnet
sudo make install
```

[Full list of open 'qemu' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fqemu-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=4` to debug crashes
