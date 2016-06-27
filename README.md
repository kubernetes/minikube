# Minikube

[![Build Status](https://travis-ci.org/kubernetes/minikube.svg?branch=master)](https://travis-ci.org/kubernetes/minikube)

## What is Minikube?

Minikube is a tool that makes it easy to run Kubernetes locally. Minikube runs a single-node Kubernetes cluster inside a VM on your laptop for users looking to try out Kubernetes or develop with it day-to-day.

### Features

* Minikube packages and configures a Linux VM, Docker and all Kubernetes components, optimized for local development.
* Minikube supports Kubernetes features such as:
  * DNS
  * NodePorts
  * ConfigMaps and Secrets
  * Dashboards

## Installation

### Requirements

* [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion) installation
* VT-x/AMD-v virtualization must be enabled in BIOS

### Instructions

See the installation instructions for the [latest release](https://github.com/kubernetes/minikube/releases).

## Quickstart

Here's a brief demo of minikube usage.
If you want to change the VM driver to VMware Fusion add the `--vm-driver=vmwarefusion` flag to `minikube start`.

Note that the IP below is dynamic and can change. It can be retrieved with `minikube ip`.

```shell
$ minikube start
Starting local Kubernetes cluster...
Running pre-create checks...
Creating machine...
Starting local Kubernetes cluster...
Kubernetes is available at https://192.168.99.100:443.

$ kubectl run hello-minikube --image=gcr.io/google_containers/echoserver:1.4 --hostport=8000 --port=8080
deployment "hello-minikube" created
$ curl http://$(minikube ip):8000
CLIENT VALUES:
client_address=192.168.99.1
command=GET
real path=/
...
$ minikube stop
Stopping local Kubernetes cluster...
Stopping "minikubeVM"...
```

### Reusing the Docker daemon

If you wish to interact with the docker daemon in the VM (such as to build images which can be used directly by the single node cluster without having to push to a docker registry), type the following into your shell:

```
export DOCKER_HOST=tcp://192.168.99.100:2376
export DOCKER_CERT_PATH=~/.minikube
export DOCKER_TLS_VERIFY=1
```
you can now run commands line the following on your host mac/windows machine:
```
docker ps
```

## Managing your Cluster

### Starting a Cluster

The [minikube start](./docs/minikube_start.md) command can be used to start your cluster.
This command creates and configures a virtual machine that runs a single-node Kubernetes cluster.
This command also configures your [kubectl](http://kubernetes.io/docs/user-guide/kubectl-overview/) installation to communicate with this cluster.

### Stopping a Cluster
The [minikube stop](./docs/minikube_stop.md) command can be used to stop your cluster.
This command shuts down the minikube virtual machine, but preserves all cluster state and data.
Starting the cluster again will restore it to it's previous state.

### Deleting a Cluster
The [minikube delete](./docs/minikube_delete.md) command can be used to delete your cluster.
This command shuts down and deletes the minikube virtual machine. No data or state is preserved.

## Interacting With your Cluster

### Kubectl

The `minikube start` command creates a "[kubectl context](http://kubernetes.io/docs/user-guide/kubectl/kubectl_config_set-context/)" called "minikube".
This context contains the configuration to communicate with your minikube cluster.

Minikube sets this context to default automatically, but if you need to switch back to it in the future, run:

`kubectl config set-context minikube`,

or pass the context on each command like this: `kubectl get pods --context=minikube`.

### Dashboard

To access the [Kubernetes Dashboard](http://kubernetes.io/docs/user-guide/ui/), run this command in a shell after starting minikube to get the address:
```shell
minikube dashboard
```

## Networking

The minikube VM is exposed to the host system via a host-only IP address, that can be obtained with the `minikube ip` command.
Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this:

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].NodePort}"'`

## Persistent Volumes

Minikube supports [PersistentVolumes](http://kubernetes.io/docs/user-guide/persistent-volumes/) of type `hostPath`.
These PersistentVolumes are mapped to a directory inside the minikube VM.

## Documentation
For a list of minikube's available commands see the [full CLI docs](https://github.com/kubernetes/minikube/blob/master/docs/minikube.md).

## Known Issues
* Features that require a Cloud Provider will not work in Minikube. These include:
  * LoadBalancers
  * PersistentVolumes
  * Ingress
* Features that require multiple nodes. These include:
  * Advanced scheduling policies
* Alternate runtimes, like rkt.

## Design

Minikube uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for provisioning VMs, and [localkube](https://github.com/kubernetes/minikube/tree/master/pkg/localkube) (originally written and donated to this project by [RedSpread](https://redspread.com/)) for running the cluster.

For more information about minikube, see the [proposal](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/local-cluster-ux.md).

## Goals and Non-Goals
For the goals and non-goals of the minikube project, please see our [roadmap](ROADMAP.md).

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

## Community

Contributions, questions, and comments are all welcomed and encouraged! minkube developers hang out on [Slack](https://kubernetes.slack.com) in the #minikube channel (get an invitation [here](http://slack.kubernetes.io/)). We also have the [kubernetes-dev Google Groups mailing list](https://groups.google.com/forum/#!forum/kubernetes-dev). If you are posting to the list please prefix your subject with "minikube: ".
