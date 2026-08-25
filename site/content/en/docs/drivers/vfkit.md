---
title: "vfkit"
weight: 2
aliases:
    - /docs/reference/drivers/vfkit
---

## Overview

[VFKit](https://github.com/crc-org/vfkit) is an open-source program for
macOS virtualization, optimized for lightweight virtual machines and
container deployment.

## Requirements

- Requires macOS 14 or later.
- Requires minikube version 1.36.0 or later.

## Networking

The vfkit driver has two networking options: `nat` and `vmnet-shared`.
The `nat` network is always available, but it does not provide access
between minikube clusters. To access other clusters or run multi-node
cluster, you need the `vmnet-shared` network. The `vmnet-shared` network
requires [vmnet-helper](https://github.com/nirs/vmnet-helper), see
installation instructions below.

{{% tabs %}}
{{% tab vmnet-shared %}}

### Requirements

- Requires [vmnet-helper](https://github.com/nirs/vmnet-helper).

{{% readfile file="/docs/drivers/includes/vmnet_helper.inc" %}}

### Usage

```shell
minikube start --driver vfkit --network vmnet-shared
```

{{% /tab %}}
{{% tab builtin %}}
### Usage

```shell
minikube start --driver vfkit [--network nat]
````

The `nat` network is used by default if the `--network` option is not
specified.

{{% /tab %}}
{{% /tabs %}}

## Issues

### Other

* [Full list of open 'vfkit' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fvfkit)

## Troubleshooting

### Run with logs

Run `minikube start --driver vfkit --alsologtostderr -v=7` to debug crashes

### Upgrade VFKit

```shell
brew update
brew upgrade vfkit
```

### Troubleshooting the vmnet-shared network

Check for errors in vmnet-helper log:

```shell
$MINIKUBE_HOME/.minikube/machines/MACHINE-NAME/vmnet-helper.log
```

Check that the `vmnet-helper` process is running:

```shell
ps au | grep vmnet-helper | grep -v grep
```

If the helper is not running restart the minikube cluster.

For help with vmnet-helper please use the
[discussions](https://github.com/nirs/vmnet-helper/discussions).
