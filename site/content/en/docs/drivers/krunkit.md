---
title: "krunkit"
weight: 2
aliases:
    - /docs/reference/drivers/krunkit
---

## Overview

[krunkit](https://github.com/libkrun/krunkit) is an open-source program for
macOS virtualization, optimized for GPU accelerated virtual machines and AI
workloads.

## Requirements

- Available only on Apple silicon.
- Requires macOS 14 or later.
- Requires minikube version 1.37.0 or later.
- Requires krunkit version 1.0.0 or later.
- Requires [vmnet-helper](https://github.com/nirs/vmnet-helper).

## Installing krunkit

To install krunkit run:

```shell
brew tap libkrun/krun
brew trust libkrun/krun # trust the tap only required if using Homebrew 6.0 or later
brew install krunkit
```

{{% alert title="Note" color="primary" %}}
If you've ever installed krunkit from the old tap `slp/krunkit` or `slp/krun`, to upgrade to the latest version you'll need to follow the [Removing the old Homebrew tap](https://github.com/libkrun/krunkit#removing-the-old-homebrew-tap) instructions first before installing as shown above.
{{% /alert %}}

## Networking

To use the krunkit driver you must install
[vmnet-helper](https://github.com/nirs/vmnet-helper), see installation
instructions below.

{{% readfile file="/docs/drivers/includes/vmnet_helper.inc" %}}

### Usage

```shell
minikube start --driver krunkit
```

### Vmnet offloading

The `--vmnet-offloading` flag enables vmnet checksum and TSO offloading,
which can improve host network throughput by up to 5x (e.g. pulling images
from a local registry, RBD mirroring):

```shell
minikube start --driver krunkit --vmnet-offloading
```

{{% alert title="Warning" color="warning" %}}
Offloading may not work for all workloads. Review the limitations below
and verify that this flag works for your workload before relying on it.
{{% /alert %}}

#### Limitations

- All VMs connected to the same vmnet bridge must enable offloading.
  If any VM on the bridge does not use offloading, the bridge disables
  TSO and TX performance drops significantly.
- Pod-to-VM network RX performance is severely degraded when offloading
  is enabled, making it almost unusable. Pod-to-host network works well
  and is faster than without offloading.

## Issues

### Other

* [Full list of open 'krunkit' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fkrunkit)

## Troubleshooting

### Run with logs

Run `minikube start --driver krunkit --alsologtostderr -v=7` to debug crashes

### Troubleshooting vmnet-helper

Check for errors in vment-helper log:

```shell
$MINIKUBE_HOME/.minikube/machines/MACHINE-NAME/vmnet-helper.log
```

Check that the `vmnet-helper` process is running:

```shell
ps au | grep vmnet-helper | grep -v grep
```

If the helper is not running restart the minikube cluster.

For help with vment-helper please use the
[discussions](https://github.com/nirs/vmnet-helper/discussions).
