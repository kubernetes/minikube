# Build Guide

## Build Requirements

* A recent Go distribution (>=1.12)
* If you're not on Linux, you'll need a Docker installation
* minikube requires at least 4GB of RAM to compile, which can be problematic when using docker-machine

### Prerequisites for different GNU/Linux distributions

#### Fedora

On Fedora you need to install _glibc-static_
```shell
$ sudo dnf install -y glibc-static
```

### Building from Source

Clone and build minikube:
```shell
$ git clone https://github.com/kubernetes/minikube.git
$ cd minikube
$ make
```

Note: Make sure that you uninstall any previous versions of minikube before building
from the source.

### Building from Source in Docker (using Debian stretch image with golang)

Clone minikube:
```shell
$ git clone https://github.com/kubernetes/minikube.git
```

Build (cross compile for linux / OS X and Windows) using make:
```shell
$ cd minikube
$ MINIKUBE_BUILD_IN_DOCKER=y make cross
```

Check "out" directory:
```shell
$ ls out/
minikube-darwin-amd64  minikube-linux-amd64  minikube-windows-amd64.exe
```

### Run Instructions

Start the cluster using your built minikube with:

```shell
$ ./out/minikube start
```

## Running Tests

### Unit Tests

Unit tests are run on Travis before code is merged. To run as part of a development cycle:

```shell
make test
```

### Integration Tests

Integration tests are currently run manually.
To run them, build the binary and run the tests:

```shell
make integration
```

You may find it useful to set various options to test only a particular test against a non-default driver. For instance:

```shell
 env TEST_ARGS="-minikube-start-args=--vm-driver=hyperkit -test.run TestStartStop" make integration
 ```

### Conformance Tests

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
