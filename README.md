# Minikube

[![Build Status](https://travis-ci.org/kubernetes/minikube.svg?branch=master)](https://travis-ci.org/kubernetes/minikube)
[![codecov](https://codecov.io/gh/kubernetes/minikube/branch/master/graph/badge.svg)](https://codecov.io/gh/kubernetes/minikube)

<img src="https://github.com/kubernetes/minikube/raw/master/logo/logo.png" width="100">

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
* Windows
    * [Hyper-V](./DRIVERS.md#hyperv-driver)
* VT-x/AMD-v virtualization must be enabled in BIOS
* `kubectl` must be on your path. Minikube currently supports any version of `kubectl` greater than 1.0, but we recommend using the most recent version.
  You can install kubectl with [these steps](http://kubernetes.io/docs/getting-started-guides/kubectl/).
* Internet connection
    * You will need a decent internet connection running `minikube start` for the first time for Minikube to pull its Docker images.
    * It might take Minikube sometime to start. It would be as fast as your internet connection with Minikube image pulling the Docker images needed.

### Instructions

See the installation instructions for the [latest release](https://github.com/kubernetes/minikube/releases).

### CI Builds

We publish CI builds of minikube, built at every Pull Request. Builds are available at (substitute in the relevant PR number):
- https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-darwin-amd64
- https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-linux-amd64
- https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-windows-amd64.exe

We also publish CI builds of minikube-iso, built at every Pull Request that touches deploy/iso/minikube-iso.  Builds are available at:
- https://storage.googleapis.com/minikube-builds/PR_NUMBER/minikube-testing.iso

## Quickstart

Here's a brief demo of minikube usage.
If you want to change the VM driver add the appropriate `--vm-driver=xxx` flag to `minikube start`. Minikube Supports
the following drivers:

* virtualbox
* vmwarefusion
* kvm ([driver installation](./DRIVERS.md#kvm-driver))
* xhyve ([driver installation](./DRIVERS.md#xhyve-driver))
* hyperv

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


### Debugging Issues With Minikube
To debug issues with minikube (not kubernetes but minikube itself), you can use the -v flag to see debug level info.  The specified values for v will do the following (the values are all encompassing in that higher values will give you all lower value outputs as well):
* --v=0 INFO level logs
* --v=1 WARNING level logs
* --v=2 ERROR level logs
* --v=3 libmachine logging
* --v=7 libmachine --debug level logging

### Using rkt container engine

To use [rkt](https://github.com/coreos/rkt) as the container runtime run:

```shell
$ minikube start \
    --network-plugin=cni \
    --container-runtime=rkt \
```

### Driver plugins

See [DRIVERS](./DRIVERS.md) for details on supported drivers and how to install
plugins, if required.

To start using the the driver, you'll need to configure minikube: `minikube config set vm-driver kvm`.

### Reusing the Docker daemon

When using a single VM of kubernetes it's really handy to reuse the Docker daemon inside the VM; as this means you don't have to build on your host machine and push the image into a docker registry - you can just build inside the same docker daemon as minikube which speeds up local experiments.

To be able to work with the docker daemon on your mac/linux host use the [docker-env command](./docs/minikube_docker-env.md) in your shell:

```
eval $(minikube docker-env)
```
you should now be able to use docker on the command line on your host mac/linux machine talking to the docker daemon inside the minikube VM:
```
docker ps
```

On Centos 7, docker may report the following error:

```
Could not read CA certificate "/etc/docker/ca.pem": open /etc/docker/ca.pem: no such file or directory
```

The fix is to update /etc/sysconfig/docker to ensure that minikube's environment changes are respected:

```
< DOCKER_CERT_PATH=/etc/docker
---
> if [ -z "${DOCKER_CERT_PATH}" ]; then
>   DOCKER_CERT_PATH=/etc/docker
> fi
```

Remember to turn off the imagePullPolicy:Always, as otherwise kubernetes won't use images you built locally.

#### Enabling Docker Insecure Registry

Minikube allows users to configure the docker engine's `--insecure-registry` flag. You can use the `--insecure-registry` flag on the
`minikube start` command to enable insecure communication between the docker engine and registries listening to requests from the CIDR range.

One nifty hack is to allow the kubelet running in minikube to talk to registries deployed inside a pod in the cluster without backing them
with TLS certificates. Because the default service cluster IP is known to be available at 10.0.0.1, users can pull images from registries
deployed inside the cluster by creating the cluster with `minikube start --insecure-registry "10.0.0.0/24"`.

## Managing your Cluster

### Starting a Cluster

The [minikube start](./docs/minikube_start.md) command can be used to start your cluster.
This command creates and configures a virtual machine that runs a single-node Kubernetes cluster.
This command also configures your [kubectl](http://kubernetes.io/docs/user-guide/kubectl-overview/) installation to communicate with this cluster.

### Configuring Kubernetes

Minikube has a "configurator" feature that allows users to configure the Kubernetes components with arbitrary values.
To use this feature, you can use the `--extra-config` flag on the `minikube start` command.

This flag is repeated, so you can pass it several times with several different values to set multiple options.

This flag takes a string of the form `component.key=value`, where `component` is one of the strings from the list below, `key` is a value on the
configuration struct and `value` is the value to set.

Valid `key`s can be found by examining the documentation for the Kubernetes `componentconfigs` for each component.
Here is the documentation for each supported configuration:

* [kubelet](https://godoc.org/k8s.io/kubernetes/pkg/apis/componentconfig#KubeletConfiguration)
* [apiserver](https://godoc.org/k8s.io/kubernetes/cmd/kube-apiserver/app/options#APIServer)
* [proxy](https://godoc.org/k8s.io/kubernetes/pkg/apis/componentconfig#KubeProxyConfiguration)
* [controller-manager](https://godoc.org/k8s.io/kubernetes/pkg/apis/componentconfig#KubeControllerManagerConfiguration)
* [etcd](https://godoc.org/github.com/coreos/etcd/etcdserver#ServerConfig)
* [scheduler](https://godoc.org/k8s.io/kubernetes/pkg/apis/componentconfig#KubeSchedulerConfiguration)

You can enable feature gates for alpha and experimental features with the `--feature-gates` flag on `minikube start`.  As of v1.5.1, the options are:

* AllAlpha=true|false (ALPHA - default=false)
* AllowExtTrafficLocalEndpoints=true|false (BETA - default=true)
* AppArmor=true|false (BETA - default=true)
* DynamicKubeletConfig=true|false (ALPHA - default=false)
* DynamicVolumeProvisioning=true|false (ALPHA - default=true)
* ExperimentalHostUserNamespaceDefaulting=true|false (ALPHA - default=false)
* StreamingProxyRedirects=true|false (ALPHA - default=false)

Note: All alpha and experimental features are not guaranteed to work with minikube.

#### Examples

To change the `MaxPods` setting to 5 on the Kubelet, pass this flag: `--extra-config=kubelet.MaxPods=5`.

This feature also supports nested structs. To change the `LeaderElection.LeaderElect` setting to `true` on the scheduler, pass this flag: `--extra-config=scheduler.LeaderElection.LeaderElect=true`.

To set the `AuthorizationMode` on the `apiserver` to `RBAC`, you can use: `--extra-config=apiserver.GenericServerRunOptions.AuthorizationMode=RBAC`.

To enable all alpha feature gates, you can use: `--feature-gates=AllAlpha=true`

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
However, Minikube is configured to persist files stored under the following directories in the minikube VM:

* `/data`
* `/var/lib/localkube`
* `/var/lib/docker`
* `/tmp/hostpath_pv`

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

You can also achieve persistence by creating a PV in a mounted host folder.

## Mounted Host Folders
Some drivers will mount a host folder within the VM so that you can easily share files between the VM and host.  These are not configurable at the moment and different for the driver and OS you are using.  Note: Host folder sharing is not implemented on Linux yet.

| Driver | OS | HostFolder | VM |
| --- | --- | --- | --- |
| Virtualbox | OSX | /Users | /Users |
| Virtualbox | Windows | C://Users | /c/Users |
| VMWare Fusion | OSX | /Users | /Users |
| Xhyve | OSX | /Users | /Users |


## Private Container Registries
**GCR/ECR**: Minikube has an addon, `registry-creds` which maps credentials into Minikube to support pulling from Google Container Registry (GCR) and Amazon's EC2 Container Registry (ECR).  To use the addon, you will need to enable it via the `addons enable registry-creds` command and then create the necessary secrets as defined here: https://github.com/upmc-enterprises/registry-creds

For other private container registries, follow the steps on [this page](http://kubernetes.io/docs/user-guide/images/).

We recommend you use ImagePullSecrets, but if you would like to configure access on the minikube VM you can place the `.dockercfg` in the `/home/docker` directory or the `config.json` in the `/home/docker/.docker` directory.

## Add-ons

Minikube has a set of built in addons that can be used enabled, disabled, and opened inside of the local k8s environment.  Below is an exampe of this functionality for the `heapster` addon:
```shell
$ minikube addons list
- addon-manager: enabled
- dashboard: enabled
- kube-dns: enabled
- heapster: disabled
- registry-creds: disabled

# minikube must be running for these commands to take effect
$ minikube addons enable heapster
heapster was successfully enabled

$ minikube addons open heapster # This will open grafana (interacting w/ heapster) in the browser
Waiting, endpoint for service is not ready yet...
Waiting, endpoint for service is not ready yet...
Created new window in existing browser session.
```
The currently supported addons include:

* [Kubernetes Dashboard](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dashboard)
* [Kube-dns](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dns)
* [Heapster](https://github.com/kubernetes/heapster): [Troubleshooting Guide](https://github.com/kubernetes/heapster/blob/master/docs/influxdb.md) Note:You will need to login to Grafana as admin/admin in order to access the console
* [Registry Credentials](https://github.com/upmc-enterprises/registry-creds)

If you would like to have minikube properly start/restart custom addons, place the addon(s) you wish to be launched with minikube in the `.minikube/addons` directory.  Addons in this folder will be moved to the minikubeVM and launched each time minikube is started/restarted.

If you have a request for an addon in minikube, please open an issue with the name and preferably a link to the addon with a description of its purpose and why it should be added.  You can also attempt to add the addon to minikube by following the guide at [ADD_ADDON.md](./ADD_ADDON.md)

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

## Minikube Environment Variables
Minikube supports passing environment variables instead of flags for every value listed in `minikube config list`.  This is done by passing an environment variable with the prefix `MINIKUBE_`For example the `minikube start --iso-url="$ISO_URL"` flag can also be set by setting the `MINIKUBE_ISO_URL="$ISO_URL"` environment variable.

Some features can only be accessed by environment variables, here is a list of these features:

* **MINIKUBE_HOME** - (string) sets the path for the .minikube directory that minikube uses for state/configuration

* **MINIKUBE_WANTUPDATENOTIFICATION** - (bool) sets whether the user wants an update notification for new minikube versions

* **MINIKUBE_REMINDERWAITPERIODINHOURS** - (int) sets the number of hours to check for an update notification
* **MINIKUBE_WANTREPORTERROR** - (bool) sets whether the user wants to send anonymous errors reports to help improve minikube

* **MINIKUBE_WANTREPORTERRORPROMPT** - (bool) sets whether the user wants to be prompted on an error that they can report them to help improve minikube

* **MINIKUBE_WANTKUBECTLDOWNLOADMSG** - (bool) sets whether minikube should tell a user that `kubectl` cannot be found on there path

* **MINIKUBE_ENABLE_PROFILING** - (int, `1` enables it) enables trace profiling to be generated for minikube which can be analyzed via:

```shell
# set env var and then run minikube
$ MINIKUBE_ENABLE_PROFILING=1 ./out/minikube start
2017/01/09 13:18:00 profile: cpu profiling enabled, /tmp/profile933201292/cpu.pprof
Starting local Kubernetes cluster...
Kubectl is now configured to use the cluster.
2017/01/09 13:19:06 profile: cpu profiling disabled, /tmp/profile933201292/cpu.pprof

# Then you can examine the profile with:
$ go tool pprof  /tmp/profile933201292/cpu.pprof
```
## Accessing Localkube Resources From Inside A Pod: Example etcd
In order to access localkube resources from inside a pod, localkube's host ip address must be used.  This can be obtained by running:
```shell
$ minikube ssh -- "sudo /usr/local/bin/localkube --host-ip"
localkube host ip:  10.0.2.15
```

You can use the host-ip:`10.0.2.15` to access localkube's resources, for example its etcd cluster.  In order to access etcd from within a pod, you can run the following command inside:
```shell
curl -L -X PUT http://10.0.2.15:2379/v2/keys/message -d value="Hello"
```

## KNOWN Issues
* Features that require a Cloud Provider will not work in Minikube. These include:
  * LoadBalancers
* Features that require multiple nodes. These include:
  * Advanced scheduling policies


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
