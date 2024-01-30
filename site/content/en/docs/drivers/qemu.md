---
title: "qemu"
weight: 3
description: >
  QEMU driver
aliases:
    - /docs/reference/drivers/qemu
---

## Overview

The `qemu` driver uses QEMU (system) for VM creation.

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

The QEMU driver has two networking options: `socket_vmnet` and `builtin`. `socket_vmnet` will give you full minikube networking functionality, such as the `service` and `tunnel` commands. On the other hand, the `builtin` network is not a dedicated network and therefore commands such as `service` and `tunnel` are not available. [socket_vmnet](https://github.com/lima-vm/socket_vmnet) can be installed via brew or from source (instructions below).

{{% tabs %}}
{{% tab socket_vmnet %}}

### Requirements

Requires macOS 10.15 or later and [socket_vmnet](https://github.com/lima-vm/socket_vmnet).

### Install socket_vmnet via [brew](https://brew.sh/)
```shell
brew install socket_vmnet
brew tap homebrew/services
HOMEBREW=$(which brew) && sudo ${HOMEBREW} services start socket_vmnet
```

### Install socket_vmnet from source (requires [Go](https://go.dev/))
```shell
git clone https://github.com/lima-vm/socket_vmnet.git && cd socket_vmnet
sudo make install
```

### Usage

```shell
minikube start --driver qemu --network socket_vmnet
```

{{% /tab %}}
{{% tab builtin %}}
### Usage

```shell
minikube start --driver qemu --network builtin
````
{{% /tab %}}
{{% /tabs %}}

## Known Issues

{{% tabs %}}
{{% tab socket_vmnet %}}
##  `/var/db/dhcpd_leases` errors

If you're seeing errors related to `/var/db/dhcpd_leases` your firewall is likely blocking the bootpd process.

Run the following to unblock bootpd from the macOS builtin firewall:
```shell
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --add /usr/libexec/bootpd
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --unblock /usr/libexec/bootpd
```
{{% /tab %}}
{{% tab builtin %}}
## Start stuck on corp machine or with custom DNS

When using the `builtin` network (default) the guest uses **only** the first `nameserver` entry in the hosts `/etc/resolv.conf` for DNS lookup. If your first `nameserver` entry is a corporate/internal DNS it's likely it will cause an issue. If you see the warning `❗ This VM is having trouble accessing https://registry.k8s.io` on `minikube start` you are likely being affected by this. This may prevent your cluster from starting entirely and you won't be able to pull remote images. More details can be found at: [#15021](https://github.com/kubernetes/minikube/issues/15021)

#### Workarounds:

1. If possible, reorder your `/etc/resolv.conf` to have a general `nameserver` entry first (eg. `8.8.8.8`) and reboot your machine.
2. Use `--network=socket_vmnet`
{{% /tab %}}
{{% /tabs %}}

[Full list of open 'qemu' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fqemu-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=4` to debug crashes
