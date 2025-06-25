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

- Requires macOS 13 or later.
- Requires minikube version 1.36.0 or later.

## Networking

The vfkit driver has two networking options: `nat` and `vmnet-shared`.
The `nat` network is always available, but it does not provide access
between minikube clusters. To access other clusters or run multi-node
cluster, you need the `vmnet-shared` network. The `vmnet-shared` network
requires [vmnet-helper](https://github.com/nirs/vmnet-helper), see
installation instructions bellow.

{{% tabs %}}
{{% tab vmnet-shared %}}

### Requirements

- Requires [vmnet-helper](https://github.com/nirs/vmnet-helper).

### Install vment-helper

```shell
tag="$(curl -fsSL https://api.github.com/repos/nirs/vmnet-helper/releases/latest | jq -r .tag_name)"
machine="$(uname -m)"
archive="vmnet-helper-$tag-$machine.tar.gz"
curl -LOf "https://github.com/nirs/vmnet-helper/releases/download/$tag/$archive"
sudo tar xvf "$archive" -C / opt/vmnet-helper
rm "$archive"
```

The command downloads the latest release from github and installs it to
`/opt/vmnet-helper`.

**IMPORTANT**: The vmnet-helper executable and the directory where it is
installed must be owned by root and may not be modifiable by
unprivileged users.

### Grant permission to run vmnet-helper

The vment-helper process must run as root to create a vmnet interface.
To allow users in the `staff` group to run the vmnet helper without a
password, you can install the default sudoers rule:

```shell
sudo install -m 0640 /opt/vmnet-helper/share/doc/vmnet-helper/sudoers.d/vmnet-helper /etc/sudoers.d/
```

You can change the sudoers configuration to allow access to specific
users or other groups.

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
