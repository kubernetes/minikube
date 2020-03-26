---
title: "hyperv"
linkTitle: "hyperv"
weight: 2
date: 2017-01-05
date: 2018-08-05
description: >
  Microsoft Hyper-V driver
---
## Overview

Hyper-V is a native hypervisor built in to modern versions of Microsoft Windows.

{{% readfile file="/docs/Reference/Drivers/includes/hyperv_usage.inc" %}}

## Special features

The `minikube start` command supports additional hyperv specific flags:

* **`--hyperv-virtual-switch`**: The hyperv virtual switch name. Defaults to first found

## Issues

Also see [co/hyperv open issues](https://github.com/kubernetes/minikube/labels/co%2Fhyperv)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
