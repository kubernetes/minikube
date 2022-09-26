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

## Issues

* [Full list of open 'qemu' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fqemu-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=4` to debug crashes
