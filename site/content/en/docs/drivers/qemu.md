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

## Special features

minikube start supports some qemu specific flags:

* **`--qemu-firmware-path`**: The path to the firmware image to be used.
  * Note: if this flag does not take effect, try removing the file `~/.minikube` which may have a reference to this setting.

## Issues

* [Full list of open 'qemu' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fqemu-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=4` to debug crashes
