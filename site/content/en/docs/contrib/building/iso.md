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
    git \
    gnupg2 \
    libtool \
    locales \
    p7zip-full \
    python2 \
    unzip \
    wget \
    xorriso
```

Install Go using these instructions:
https://go.dev/doc/install

To build without docker run:

```shell
IN_DOCKER=1 make minikube-iso-<arch>
```

{{% alert title="Note" color="primary" %}}
Some external projects will try to use docker even when building
without docker. You must install docker on the build host.
{{% /alert %}}

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

## Updating the kernel

The ISO tracks a Linux longterm series (currently 6.6.x) as a **custom**
Buildroot kernel version. Buildroot's "Latest version" menu item is a different
stable series and is not what minikube uses.

Repeat the following for both `x86_64` and `aarch64`. Both
architectures share `patches/linux/linux.hash`; the hash script updates
that file only when the tarball line is missing.

1. Set the kernel version in Buildroot kconfig:

   ```shell
   make iso-menuconfig-<arch>
   ```

   Under **Kernel**:

   - **Kernel version** → **Custom version**
   - **Kernel version** string → the current longterm patch from
     [kernel.org](https://www.kernel.org/) (for example `6.6.152`)

   Under **Toolchain**, confirm **Kernel Headers** is the same series as the
   kernel (`6.6.x` for `6.6.152`). The exact headers patch (6.6.141 vs 6.6.152)
   does not matter and should be left as Buildroot set it. Only change headers
   when moving to a new series (for example 6.6 → 6.12).

   Save and exit. The Makefile runs `savedefconfig`, which writes
   `deploy/iso/minikube-iso/configs/minikube_<arch>_defconfig`
   (`BR2_LINUX_KERNEL_CUSTOM_VERSION_VALUE`).

2. Update the kernel tarball hash:

   ```shell
   hack/iso/update-kernel-hash.sh <arch>
   ```

   This reads `BR2_LINUX_KERNEL_CUSTOM_VERSION_VALUE` from that arch's
   defconfig. If `linux.hash` already has that tarball, it exits. Otherwise
   it fetches the sha256 from kernel.org and replaces the previous tarball
   line. The script is Linux-only (GNU `sed -i`), like the ISO build.

3. Refresh the board kernel config against the new tarball:

   ```shell
   make linux-menuconfig-<arch>
   ```

   Save and exit (accept new defaults). That runs `linux-savedefconfig` and
   `linux-update-defconfig`, updating
   `deploy/iso/minikube-iso/board/minikube/<arch>/linux_<arch>_defconfig`.

The kernel bump is a source change (`defconfig`, `linux.hash`, and the board
kernel defconfig). Rebuild the ISO to verify, then send a PR.

If that rebuild is incremental (reusing `out/buildroot/output-*`),
remove leftover kernel trees first. Buildroot does not delete files from
`target/` when the version changes, so old `/lib/modules/*` directories
are packed into the ISO and unpacked onto tmpfs at boot.

```shell
rm -rf out/buildroot/output-*/build/linux-[0-9]*
rm -rf out/buildroot/output-*/target/lib/modules/*
```

`linux-[0-9]*` matches `linux-6.6.95` and not `linux-headers-*`. A clean
`output-*` does not need this.

## Adding kernel modules

To change kernel `CONFIG_*` options without bumping the version:

```shell
$ make linux-menuconfig-<arch>
```

This opens the kernel configuration menu and saves the result to
`deploy/iso/minikube-iso/board/minikube/<arch>/linux_<arch>_defconfig`.

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

