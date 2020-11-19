---
title: "Testing"
date: 2019-07-31
weight: 3
description: >
  How to run tests
---

## Prerequisites

* Go distribution
  * Specific version depends on minikube version.
  * The current dependency version can be found here : [master branch's go.mod file](https://github.com/kubernetes/minikube/blob/master/go.mod).
* If you are on Linux, you will need to install `libvirt-dev`, since unit tests need kvm2 driver:

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

* Clone the Kubernetes repo somewhere on your system.
* Run `make quick-release` in the k8s repo.
* Start up a minikube cluster with: `minikube start`.
* Set following two environment variables:

```shell
export KUBECONFIG=$HOME/.kube/config
export KUBERNETES_CONFORMANCE_TEST=y
```

* Run the tests (from the k8s repo):

```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=\[Conformance\]" --check-version-skew=false
```

To run a specific conformance test, you can use the `ginkgo.focus` flag to filter the set using a regular expression.
The `hack/e2e.go` wrapper and the `e2e.sh` wrappers have a little trouble with quoting spaces though, so use the `\s` regular expression character instead.
For example, to run the test `should update annotations on modification [Conformance]`, use following command:

```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=should\supdate\sannotations\son\smodification" --check-version-skew=false
```
