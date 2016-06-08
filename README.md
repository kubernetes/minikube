# Minikube

Run Kubernetes locally

[![Build Status](https://travis-ci.org/kubernetes/minikube.svg?branch=master)](https://travis-ci.org/kubernetes/minikube)

## Background

Minikube is a tool that makes it easy to run Kubernetes locally. Minikube runs
a single-node Kubernetes cluster inside a VM on your laptop for users looking
to try out Kubernetes or develop with it day-to-day.

## Requirements For Running Minikube
* VirtualBox installation
* VT-x/AMD-v virtualization must be enabled in BIOS

## Installation
See the installation instructions for the [latest release](https://github.com/kubernetes/minikube/releases).

## Usage

Here's a brief demo of minikube usage. We're using the code from this [Kubernetes tutorial](http://kubernetes.io/docs/hellonode/).
Note that the IP below is dynamic and can change. It can be retrieved with `minikube ip`.

```shell
$ minikube start
Starting local Kubernetes cluster...
Running pre-create checks...
Creating machine...
Starting local Kubernetes cluster...
Kubernetes is available at https://192.168.99.100:443.

$ eval $(minikube docker-env)
$ docker build -t helloworld .
Successfully built d16fe85e1abe
$ kubectl run hello-minikube --image=helloworld --hostport=8000 --port=8080 --generator=run-pod/v1
pod "hello-minikube" created
$ curl http://$(minikube ip):8000
Hello World!
$ minikube stop
Stopping local Kubernetes cluster...
Stopping "minikubeVM"...
```
### Documentation
For a list of minikube's available commands see: [minikube docs](https://github.com/kubernetes/minikube/blob/master/docs/minikube.md)

### Dashboard

To access the dashboard, run this command in a shell after starting minikube to get the address:
```shell
echo $(minikube ip):$(kubectl get service kubernetes-dashboard --namespace=kube-system -o=jsonpath='{.spec.ports[0].nodePort}{"\n"}')
```
And then copy/paste that into your browser.

## Features
 * Minikube packages and configures a Linux VM, Docker and all Kubernetes components, optimized for local development.
 * Minikube supports Kubernetes features such as:
   * DNS
   * NodePorts
   * ConfigMaps and Secrets
   * Dashboards

## Known Issues
 * Features that require a Cloud Provider will not work in Minikube. These include:
  * LoadBalancers
  * PersistentVolumes
  * Ingress
 * Features that require multiple nodes. These include:
  * Advanced scheduling policies
  * DaemonSets
 * Alternate runtimes, like rkt.

If you need these features, don't worry! We're planning to add these to minikube over time. Please leave a note in the
issue tracker about how you'd like to use minikube!

## Design

Minikube uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for provisioning VMs, and [localkube](https://github.com/kubernetes/minikube/tree/master/pkg/localkube) (originally written and donated to this project by [RedSpread](https://redspread.com/)) for running the cluster.

For more information about minikube, see the [proposal](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/local-cluster-ux.md).

## Goals

* Works across multiple OSes - OS X, Linux and Windows primarily.
* Single command setup and teardown UX.
* Unified UX across OSes
* Minimal dependencies on third party software.
* Minimal resource overhead.
* Replace any other alternatives to local cluster deployment.

## Non Goals

* Simplifying kubernetes production deployment experience. Kube-deploy is attempting to tackle this problem.
* Supporting all possible deployment configurations of Kubernetes like various types of storage, networking, etc.

## Development Guide

See [CONTRIBUTING.md](CONTRIBUTING.md) for an overview of how to send pull requests.

### Build Requirements

* A recent Go distribution (>1.6)
* If you're not on Linux, you'll need a Docker installation
* Minikube requires at least 4GB of RAM to compile, which can be problematic when using docker-machine

### Build Instructions

```shell
make out/minikube
```

### Run Instructions

Start the cluster using your built minikube with:

```shell
$ ./out/minikube start
```

### Running Tests

#### Unit Tests

Unit tests are run on Travis before code is merged. To run as part of a development cycle:

```shell
make test
```

#### Integration Tests

Integration tests are currently run manually. 
To run them, build the binary and run the tests:

```shell
make integration
```

#### Conformance Tests

These are kubernetes tests that run against an arbitrary cluster and exercise a wide range of kubernetes features.
You can run these against minikube by following these steps:

* Clone the kubernetes repo somewhere on your system.
* Run `make quick-release` in the k8s repo.
* Start up a minikube cluster with: `minikube start`.
* Set these two environment variables:
```shell
export KUBECONFIG=$HOME/.kube/config
export KUBERNETES_CONFORMANCE_TEST=y
```
* Run the tests (from the k8s repo):
```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=\[Conformance\]" --check_version_skew=false --check_node_count=false
```
