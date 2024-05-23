---
title: "Testing"
date: 2019-07-31
weight: 3
description: >
  How to run tests
---

## Prerequisites

- Go distribution
  - Specific version depends on minikube version.
  - The current dependency version can be found here : [master branch's go.mod file](https://github.com/kubernetes/minikube/blob/master/go.mod).
- If you are on Linux, you will need to install `libvirt-dev`, since unit tests need kvm2 driver:

```shell
# For Debian based
sudo apt-get install libvirt-dev

# For Centos
yum install libvirt-devel

# For Fedora
dnf install libvirt-devel
```

## Unit Tests

Unit tests are run on Travis before code is merged. To run as part of a development cycle:

```shell
make test
```

## Integration Tests

### The basics

From the minikube root directory, build the binary and run the tests:

```shell
make integration
```

You may find it useful to set various options to test only a particular test against a non-default driver. For instance:

```shell
 env TEST_ARGS="-minikube-start-args=--driver=hyperkit -test.run TestStartStop" make integration
```

### Quickly iterating on a single test

Run a single test on an active cluster:

```shell
make integration -e TEST_ARGS="-test.run TestFunctional/parallel/MountCmd --profile=minikube --cleanup=false"
```

WARNING: For this to work repeatedly, the test must be written so that it cleans up after itself.

The `--cleanup=false` test arg ensures that the cluster will not be deleted after the test is run.

See [main_test.go](https://github.com/kubernetes/minikube/blob/master/test/integration/main_test.go) for details.

### Disabling parallelism

```shell
make integration -e TEST_ARGS="-test.parallel=1"
```

### Testing philosophy

- Tests should be so simple as to be correct by inspection
- Readers should need to read only the test body to understand the test
- Top-to-bottom readability is more important than code de-duplication

Tests are typically read with a great air of skepticism, because chances are they are being read only when things are broken.

## Conformance Tests

These are Kubernetes tests that run against an arbitrary cluster and exercise a wide range of Kubernetes features.
You can run these against minikube by following these steps:

Install [docker](https://docs.docker.com/engine/install/)
Install [kubectl](https://v1-18.docs.kubernetes.io/docs/tasks/tools/install-kubectl/)
Clone the [minikube repo](https://github.com/kubernetes/minikube)

## Compile the latest minikube binary

```console
% cd <minikube dir>
% make
```

## Trigger the tests and get back the results

```console
% cd <minikube dir>
./hack/conformance_tests.sh out/minikube --driver=docker --container-runtime=docker --kubernetes-version=stable
```

This script will run the latest sonobuoy against a minikube cluster with two nodes and the provided parameters.
