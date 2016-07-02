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

* [VirtualBox](https://www.virtualbox.org/wiki/Downloads), [VMware Fusion](https://www.vmware.com/products/fusion)
or [KVM](http://www.linux-kvm.org/) installation
* VT-x/AMD-v virtualization must be enabled in BIOS

### Instructions

See the installation instructions for the [latest release](https://github.com/kubernetes/minikube/releases).

## Quickstart

Here's a brief demo of minikube usage.
If you want to change the VM driver add the appropriate `--vm-driver=xxx` flag to `minikube start`. Minikube Supports
the following drivers:

* virtualbox
* vmwarefusion
* kvm ([driver installation](#kvm-driver))

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
# We have now launched an echoserver pod but we have to wait until the pod is up before curling/accessing it
# To check whether the pod is up and running we can use the following:
$ kubectl get pod
NAME                              READY     STATUS              RESTARTS   AGE
hello-minikube-3383150820-vctvh   1/1       ContainerCreating   0          3s
# We can see that the pod is still being created from the ContainerCreating status
$ kubectl get pod
NAME                              READY     STATUS    RESTARTS   AGE
hello-minikube-3383150820-vctvh   1/1       Running   0          13s
# We can see that the pod is now Running and we will now be able to curl it:
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

### Driver plugin installation

Minikube uses Docker Machine to manage the Kubernetes VM so it benefits from the
driver plugin architecture that Docker Machine uses to provide a consistent way to
manage various VM providers. Minikube embeds VirtualBox and VMware Fusion drivers
so there are no additional steps to use them. However, other drivers require an
extra binary to be present in the host PATH.

#### KVM driver

Download the `docker-machine-driver-kvm` binary from
https://github.com/dhiltgen/docker-machine-kvm/releases and put it somewhere in
your PATH. Minikube is currently tested against `docker-machine-driver-kvm` 0.7.0.

### Reusing the Docker daemon

When using a single VM of kubernetes its really handy to reuse the Docker daemon inside the VM; as this means you don't have to build on your host machine and push the image into a docker registry - you can just build inside the same docker daemon as minikube which speeds up local experiments.

To be able to work with the docker daemon on your mac/linux host use the [docker-env command](https://github.com/kubernetes/minikube/blob/master/docs/minikube_docker-env.md) in your shell:

```
eval $(minikube docker-env)
```
you should now be able to use docker on the command line on your host mac/linux machine talking to the docker daemon inside the minikube VM:
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

## Private Container Registries

To access a private container registry, follow the steps on [this page](http://kubernetes.io/docs/user-guide/images/).

We recommend you use ImagePullSecrets, but if you would like to configure access on the minikube VM you can place the `.dockercfg` in the `/home/docker` directory or the `config.json` in the `/home/docker/.docker` directory.

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

To run a specific Conformance Test, you can use the `ginkgo.focus` flag to filter the set using a regular expression.
The hack/e2e.go wrapper and the e2e.sh wrappers have a little trouble with quoting spaces though, so use the `\s` regular expression character instead.
For example, to run the test `should update annotations on modification [Conformance]`, use this command:

```shell
go run hack/e2e.go -v --test --test_args="--ginkgo.focus=should\supdate\sannotations\son\smodification" --check_version_skew=false --check_node_count=false
```

#### Adding a New Dependency

Minikube uses `Godep` to manage vendored dependencies.
`Godep` can be a bit finnicky with a project with this many dependencies.
Here is a rough set of steps that usually works to add a new dependency.

1. Make a clean GOPATH, with minikube in it.
This isn't strictly necessary, but it usually helps.

```shell
mkdir -p $HOME/newgopath/src/k8s.io
export GOPATH=$HOME/newgopath
cd $HOME/newgopath/src/k8s.io
git clone https://github.com/kubernetes/minikube.git
```

2. `go get` your new dependency.
```shell
go get mynewdepenency
```

3. Use it in code, build and test.

4. Import the dependency from GOPATH into vendor/
```shell
godep save ./...
```

If it is a large dependency, please commit the vendor/ directory changes separately.
This makes review easier in Github.

```shell
git add vendor/
git commit -m "Adding dependency foo"
git add --all
git commit -m "Adding cool feature"
```

#### Updating Kubernetes

To update Kubernetes, follow these steps:

1. Make a clean GOPATH, with minikube in it.
This isn't strictly necessary, but it usually helps.

 ```shell
 mkdir -p $HOME/newgopath/src/k8s.io
 export GOPATH=$HOME/newgopath
 cd $HOME/newgopath/src/k8s.io
 git clone https://github.com/kubernetes/minikube.git
 ```

2. Copy your vendor directory back out to the new GOPATH.

 ```shell
 cd minikube
 godep restore ./...
 ```

3. Kubernetes should now be on your GOPATH. Check it out to the right version.
Make sure to also fetch tags, as Godep relies on these.

 ```shell
 cd $GOPATH/src/k8s.io/kubernetes
 git fetch --tags
 ```
 
 Then list all available Kubernetes tags:

 ```shell
 git tag
 ...
 v1.2.4
 v1.2.4-beta.0
 v1.3.0-alpha.3
 v1.3.0-alpha.4
 v1.3.0-alpha.5
 ...
```

 Then checkout the correct one and update its dependencies with:
 
 ```shell
 git checkout $DESIREDTAG
 godep restore ./...
 ```

4. Build and test minikube, making any manual changes necessary to build.

5. Update godeps

 ```shell
 cd $GOPATH/src/k8s.io/minikube
 rm -rf Godeps/ vendor/
 godep save ./...
 ```

 6. Verify that the correct tag is marked in the Godeps.json file by running this script:

 ```shell
 python hack/get_k8s_version.py
 -X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.gitCommit=caf9a4d87700ba034a7b39cced19bd5628ca6aa3 -X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.gitVersion=v1.3.0-beta.2 -X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.gitTreeState=clean
```

The `-X k8s.io/minikube/vendor/k8s.io/kubernetes/pkg/version.gitVersion` flag should contain the right tag.

Once you've build and started minikube, you can also run:

```shell
kubectl version
Client Version: version.Info{Major:"1", Minor:"2", GitVersion:"v1.2.4", GitCommit:"3eed1e3be6848b877ff80a93da3785d9034d0a4f", GitTreeState:"clean"}
Server Version: version.Info{Major:"1", Minor:"3+", GitVersion:"v1.3.0-beta.2", GitCommit:"caf9a4d87700ba034a7b39cced19bd5628ca6aa3", GitTreeState:"clean"}
```

The Server Version should contain the right tag in `version.Info.GitVersion`.

If any manual changes were required, please commit the vendor changes separately.
This makes the change easier to view in Github.

```shell
git add vendor/
git commit -m "Updating Kubernetes to foo"
git add --all
git commit -m "Manual changes to update Kubernetes to foo"
```

## Community

Contributions, questions, and comments are all welcomed and encouraged! minkube developers hang out on [Slack](https://kubernetes.slack.com) in the #minikube channel (get an invitation [here](http://slack.kubernetes.io/)). We also have the [kubernetes-dev Google Groups mailing list](https://groups.google.com/forum/#!forum/kubernetes-dev). If you are posting to the list please prefix your subject with "minikube: ".
