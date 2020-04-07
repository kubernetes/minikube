---
title: "hyperv"
weight: 2
aliases:
    - /docs/reference/drivers/hyperv
---
## Overview

Hyper-V is a native hypervisor built in to modern versions of Microsoft Windows.

{{% readfile file="/docs/drivers/includes/hyperv_usage.inc" %}}

## Special features

The `minikube start` command supports additional hyperv specific flags:

* **`--hyperv-virtual-switch`**: The hyperv virtual switch name. Defaults to first found

## Issues

Also see [co/hyperv open issues](https://github.com/kubernetes/minikube/labels/co%2Fhyperv)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
