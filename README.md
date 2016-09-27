# Minikube

[![Build Status](https://travis-ci.org/kubernetes/minikube.svg?branch=master)](https://travis-ci.org/kubernetes/minikube)
[![codecov](https://codecov.io/gh/kubernetes/minikube/branch/master/graph/badge.svg)](https://codecov.io/gh/kubernetes/minikube)

## What is Minikube?

Minikube is a tool that makes it easy to run Kubernetes locally. Minikube runs a single-node Kubernetes cluster inside a VM on your laptop for users looking to try out Kubernetes or develop with it day-to-day.

### Features

* Minikube packages and configures a Linux VM, the container runtime, and all Kubernetes components, optimized for local development.
* Minikube supports Kubernetes features such as:
  * DNS
  * NodePorts
  * ConfigMaps and Secrets
  * Dashboards
  * Container Runtime: Docker, and [rkt](https://github.com/coreos/rkt)
  * Enabling CNI (Container Network Interface)

## Installation

### Requirements

* OS X
    * [xhyve driver](./DRIVERS.md#xhyve-driver), [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Fusion](https://www.vmware.com/products/fusion) installation
* Linux
    * [VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [KVM](http://www.linux-kvm.org/) installation,
* VT-x/AMD-v virtualization must be enabled in BIOS
* `kubectl` must be on your path. Minikube currently supports any version of `kubectl` greater than 1.0, but we recommend using the most recent version.
  You can install kubectl with [these steps](http://kubernetes.io/docs/getting-started-guides/minikube/#download-kubectl).


### Instructions

See the installation instructions for the [latest release](https://github.com/kubernetes/minikube/releases).

## Quickstart

Here's a brief demo of minikube usage.
If you want to change the VM driver add the appropriate `--vm-driver=xxx` flag to `minikube start`. Minikube Supports
the following drivers:

* virtualbox
* vmwarefusion
* kvm ([driver installation](./DRIVERS.md#kvm-driver))
* xhyve ([driver installation](./DRIVERS.md#xhyve-driver))

Note that the IP below is dynamic and can change. It can be retrieved with `minikube ip`.

```shell
$ minikube start
Starting local Kubernetes cluster...
Running pre-create checks...
Creating machine...
Starting local Kubernetes cluster...

$ kubectl run hello-minikube --image=gcr.io/google_containers/echoserver:1.4 --port=8080
deployment "hello-minikube" created
$ kubectl expose deployment hello-minikube --type=NodePort
service "hello-minikube" exposed

# We have now launched an echoserver pod but we have to wait until the pod is up before curling/accessing it
# via the exposed service.
# To check whether the pod is up and running we can use the following:
$ kubectl get pod
NAME                              READY     STATUS              RESTARTS   AGE
hello-minikube-3383150820-vctvh   1/1       ContainerCreating   0          3s
# We can see that the pod is still being created from the ContainerCreating status
$ kubectl get pod
NAME                              READY     STATUS    RESTARTS   AGE
hello-minikube-3383150820-vctvh   1/1       Running   0          13s
# We can see that the pod is now Running and we will now be able to curl it:
$ curl $(minikube service hello-minikube --url)
CLIENT VALUES:
client_address=192.168.99.1
command=GET
real path=/
...
$ minikube stop
Stopping local Kubernetes cluster...
Stopping "minikube"...
```

To use [rkt](https://github.com/coreos/rkt) as the container runtime, execute:

```shell
$ minikube start \
    --network-plugin=cni \
    --container-runtime=rkt \
    --iso-url=https://github.com/coreos/minikube-iso/releases/download/v0.0.3/minikube-v0.0.3.iso
```

This will use an alternative minikube ISO image containing both rkt, and Docker, and enable CNI networking.

### Driver plugins

See [DRIVERS](./DRIVERS.md) for details on supported drivers and how to install
plugins, if required.

### Reusing the Docker daemon

When using a single VM of kubernetes its really handy to reuse the Docker daemon inside the VM; as this means you don't have to build on your host machine and push the image into a docker registry - you can just build inside the same docker daemon as minikube which speeds up local experiments.

To be able to work with the docker daemon on your mac/linux host use the [docker-env command](./docs/minikube_docker-env.md) in your shell:

```
eval $(minikube docker-env)
```
you should now be able to use docker on the command line on your host mac/linux machine talking to the docker daemon inside the minikube VM:
```
docker ps
```

Remember to turn off the imagePullPolicy:Always, as otherwise kubernetes won't use images you built locally.

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

`kubectl config use-context minikube`,

or pass the context on each command like this: `kubectl get pods --context=minikube`.

### Dashboard

To access the [Kubernetes Dashboard](http://kubernetes.io/docs/user-guide/ui/), run this command in a shell after starting minikube to get the address:
```shell
minikube dashboard
```

### Services

To access a service exposed via a node port, run this command in a shell after starting minikube to get the address:
```shell
minikube service [-n NAMESPACE] [--url] NAME
```

## Networking

The minikube VM is exposed to the host system via a host-only IP address, that can be obtained with the `minikube ip` command.
Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this:

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].NodePort}"'`

## Persistent Volumes
Minikube supports [PersistentVolumes](http://kubernetes.io/docs/user-guide/persistent-volumes/) of type `hostPath`.
These PersistentVolumes are mapped to a directory inside the minikube VM.

The Minikube VM boots into a tmpfs, so most directories will not be persisted across reboots (`minikube stop`).
However, Minikube is configured to persist files stored under the following host directories:

* `/data`
* `/var/lib/localkube`
* `/var/lib/docker`

Here is an example PersistentVolume config to persist data in the '/data' directory:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv0001
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: 5Gi
  hostPath:
    path: /data/pv0001/
```

## Private Container Registries

To access a private container registry, follow the steps on [this page](http://kubernetes.io/docs/user-guide/images/).

We recommend you use ImagePullSecrets, but if you would like to configure access on the minikube VM you can place the `.dockercfg` in the `/home/docker` directory or the `config.json` in the `/home/docker/.docker` directory.

## Add-ons

In order to have minikube properly start/restart custom addons, place the addon(s) you wish to be launched with minikube in the `.minikube/addons` directory.  Addons in this folder will be moved to the minikubeVM and launched each time minikube is started/restarted.

## Documentation

For a list of minikube's available commands see the [full CLI docs](./docs/minikube.md).

## Using Minikube with an HTTP Proxy

Minikube creates a Virtual Machine that includes Kubernetes and a Docker daemon.
When Kubernetes attempts to schedule containers using Docker, the Docker daemon may require external network access to pull containers.

If you are behind an HTTP proxy, you may need to supply Docker with the proxy settings.
To do this, pass the required environment variables as flags during `minikube start`.

For example:

```shell
$ minikube start --docker-env HTTP_PROXY=http://$YOURPROXY:PORT \
                 --docker-env HTTPS_PROXY=https://$YOURPROXY:PORT
```


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

## Additional Links:
* **Goals and Non-Goals**: For the goals and non-goals of the minikube project, please see our [roadmap](./ROADMAP.md).
* **Development Guide**: See [CONTRIBUTING.md](./CONTRIBUTING.md) for an overview of how to send pull requests.
* **Building Minikube**: For instructions on how to build/test minikube from source, see the [build guide](./BUILD_GUIDE.md)
* **Adding a New Dependency**: For instructions on how to add a new dependency to minikube see the [adding dependencies guide](./ADD_DEPENDENCY.md)
* **Updating Kubernetes**: For instructions on how to add a new dependency to minikube see the [updating kubernetes guide](./UPDATE_KUBERNETES.md)
* **Steps to Release Minikube**: For instructions on how to release a new version of minikube see the [release guide](./RELEASING.md)
* **Steps to Release Localkube**: For instructions on how to release a new version of localkube see the [localkube release guide](./LOCALKUBE_RELEASING.md)

## Community

Contributions, questions, and comments are all welcomed and encouraged! minkube developers hang out on [Slack](https://kubernetes.slack.com) in the #minikube channel (get an invitation [here](http://slack.kubernetes.io/)). We also have the [kubernetes-dev Google Groups mailing list](https://groups.google.com/forum/#!forum/kubernetes-dev). If you are posting to the list please prefix your subject with "minikube: ".
