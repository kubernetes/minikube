---
linkTitle: "ISO Build"
title: "Building the minikube ISO"
date: 2019-08-09
weight: 4
---
## Overview

The minikube ISO is booted by each hypervisor to provide a stable minimal Linux environment to start Kubernetes from. It is based on coreboot, uses systemd, and includes all necessary container runtimes and hypervisor guest drivers.

## Prerequisites

* A recent Go distribution (>=1.16)
* If you are on Windows, you'll need Docker to be installed.
* 4GB of RAM
* Build tools:

```shell
sudo apt-get install build-essential gnupg2 p7zip-full git wget cpio python \
    unzip bc gcc-multilib automake libtool locales
```

Additionally, if you are on Fedora, you will need to install `glibc-static`:

```shell
sudo dnf install -y glibc-static
```

## Downloading the source

```shell
git clone https://github.com/kubernetes/minikube.git
cd minikube
```

### Building

```shell
$ make buildroot-image
$ make out/minikube.iso
```

The build will occur inside a docker container. If you want to do this on
baremetal, replace `make out/minikube.iso` with `IN_DOCKER=1 make out/minikube.iso`.
The bootable ISO image will be available in `out/minikube.iso`.

### Using a local ISO image

```shell
$ ./out/minikube start --iso-url=file://$(pwd)/out/minikube.iso
```

### Modifying buildroot components

To change which Linux userland components are included by the guest VM, use this to modify the buildroot configuration:

```shell
cd out/buildroot
make menuconfig
make
```

To save these configuration changes, execute:

```shell
make savedefconfig
```

The changes will be reflected in the `minikube-iso/configs/minikube_defconfig` file.

### Adding kernel modules

To make kernel configuration changes and save them, execute:

```shell
$ make linux-menuconfig
```

This will open the kernel configuration menu, and then save your changes to our
iso directory after they've been selected.

### Adding third-party packages

To add your own package to the minikube ISO, create a package directory under `iso/minikube-iso/package`.  This directory will require at least 3 files:

`<package name>.mk` - A Makefile describing how to download the source code and build the program  
`<package name>.hash` - Checksums to verify the downloaded source code  
`Config.in` - buildroot configuration

For a relatively simple example to start with, you may want to reference the `podman` package.

## Continuous Integration Builds

We publish CI builds of minikube, built at every Pull Request. Builds are available at (substitute in the relevant PR number):

- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-darwin-amd64>
- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-linux-amd64>
- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-windows-amd64.exe>

We also publish CI builds of minikube-iso, built at every Pull Request that touches deploy/iso/minikube-iso.  Builds are available at:

- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-testing.iso>
