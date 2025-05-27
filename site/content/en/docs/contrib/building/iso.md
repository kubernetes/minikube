---
linkTitle: "ISO Build"
title: "Building the minikube ISO"
date: 2019-08-09
weight: 4
---
## Overview

The minikube ISO is booted by each hypervisor to provide a stable minimal Linux environment to start Kubernetes from. It is based on coreboot, uses systemd, and includes all necessary container runtimes and hypervisor guest drivers.

## Prerequisites

* Machine with x86\_64 CPU
* Ubuntu 22.04.5 LTS (Jammy Jellyfish)
* docker
* A recent GNU Make distribution (>=4.0)
* A recent Go distribution (>=1.22.0)
* If you are on Windows or Mac, you'll need Docker to be installed.
* 4GB of RAM

## Downloading the source

```shell
git clone https://github.com/kubernetes/minikube.git
cd minikube
```

## Building

### Building in Docker

To build for x86:

```shell
$ make buildroot-image
$ make minikube-iso-x86_64
```

To build for ARM:

```shell
$ make buildroot-image
$ make minikube-iso-aarch64
```

The build will occur inside a docker container.
The bootable ISO image will be available in `out/minikube-<arch>.iso`.

### Building without docker

Install required tools:

```shell
sudo apt-get install \
    automake \
    bc \
    build-essential \
    cpio \
    gcc-multilib \
    genisoimage \
    git \
    gnupg2 \
    libtool \
    locales \
    p7zip-full \
    python2 \
    unzip \
    wget \
```

Install Go using these instructions:
https://go.dev/doc/install

To build without docker run:

```shell
IN_DOCKER=1 make minikube-iso-<arch>
```

> [!IMPORTANT]
> Some external projects will try to use docker even when building
> without docker. You must install docker on the build host.

## Using a local ISO image

```shell
$ ./out/minikube start --iso-url=file://$(pwd)/out/minikube-<arch>.iso
```

## Modifying buildroot components

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

## Adding kernel modules

To make kernel configuration changes and save them, execute:

```shell
$ make linux-menuconfig
```

This will open the kernel configuration menu, and then save your changes to our
iso directory after they've been selected.

## Adding third-party packages

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

