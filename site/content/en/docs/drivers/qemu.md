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

## Known Issues

### 1. Start stuck with `user` network on corp machine or custom DNS

When using the `user` network (default) the guest uses **only** the first `nameserver` entry in the hosts `/etc/resolv.conf` for DNS lookup. If your first `nameserver` entry is a corporate/internal DNS it's likely it will cause an issue. If you see the warning `❗ This VM is having trouble accessing https://registry.k8s.io` on `minikube start` you are likely being affected by this. This may prevent your cluster from starting entirely and you won't be able to pull remote images. More details can be found at: [#15021](https://github.com/kubernetes/minikube/issues/15021)

##### Workarounds:
1. If possible, reorder your `/etc/resolv.conf` to have a general `nameserver` entry first (eg. `8.8.8.8`) and reboot your machine.
2. (Coming soon) Use `--network=socket_vmnet`

[Full list of open 'qemu' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fqemu-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=4` to debug crashes
