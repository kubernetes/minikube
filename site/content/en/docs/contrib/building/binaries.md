---
linkTitle: "Building"
title: "Building the minikube binary"
date: 2019-07-31
weight: 2
---

## Prerequisites

* A recent Go distribution (>=1.12)
* If you are on Windows, you'll need Docker to be installed.
* 4GB of RAM

Additionally, if you are on Fedora, you will need to install `glibc-static`:

```shell
sudo dnf install -y glibc-static
```

## Downloading the source

```shell
git clone https://github.com/kubernetes/minikube.git
cd minikube
```

## Compiling minikube

```shell
make
```

Note: On Windows, this will only work in Git Bash or other terminals that support bash commands.

You can also build platform specific executables like below:
    1. `make windows` will build the binary for Windows platform
    2. `make linux` will build the binary for Linux platform
    3. `make darwin` will build the binary for Darwin/Mac platform

## Compiling minikube using Docker

To cross-compile to/from different operating systems:

```shell
MINIKUBE_BUILD_IN_DOCKER=y make cross
```

The resulting binaries for each platform will be located in the `out/` subdirectory.

## Using a source-built minikube binary

Start the cluster using your built minikube with:

```shell
./out/minikube start
```

## Building the ISO

See [Building the minikube ISO](../iso)

## Continuous Integration Builds

We publish CI builds of minikube, built at every Pull Request. Builds are available at (substitute in the relevant PR number):

- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-darwin-amd64>
- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-darwin-arm64>
- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-linux-amd64>
- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-linux-arm64>
- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-windows-amd64.exe>

We also publish CI builds of minikube-iso, built at every Pull Request that touches deploy/iso/minikube-iso.  Builds are available at:

- <https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-testing.iso>
