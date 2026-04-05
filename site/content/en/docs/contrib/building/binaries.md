---
linkTitle: "Building"
title: "Building the minikube binary"
date: 2019-07-31
weight: 2
---

## Prerequisites

* A recent Go distribution (>=1.22.0)
* If you are on Windows, you'll need to have installed
  * Docker
  * GNU Make e.g. `winget install GnuWin32.make`, then add `C:\Program Files (x86)\GnuWin32\bin` to `PATH`
* 4GB of RAM

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

## Build minikube

To build and compile minikube on your machine simply run
```shell
make
```
and minikube binary will be available in ./out/minikube

## build minikube in docker

if you have issues running make due to tooling issue you can run the make in "docker"
```shell
MINIKUBE_BUILD_IN_DOCKER=y make
```

## build binaries for other platforms
if you wanted to build binaries for all platforms -linux,darwin,windows (cross-compile) to/from different operating systems:

```shell
MINIKUBE_BUILD_IN_DOCKER=y make cross
```

The resulting binaries for each platform will be located in the `out/` subdirectory.

## Using a source-built minikube binary

Start the cluster using your built minikube with:

```shell
./out/minikube start
```

## Unit Test and lint
```shell
make test
```

## clean and go mod tidy
```shell
make clean
make gomodtidy
```

## Run Short integration test (functional test)
```shell
make functional
```

To see HTML report of the functional test you can install [gopogh](https://github.com/medyagh/gopogh)
and run 
```shell
make html_report
```
This will produce an html report in `./out/` folder

### learn more about other make targets
```shell
make help
```

## Testing

See the [Testing Guide](../testing.en.md) for information on testing minikube.



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
