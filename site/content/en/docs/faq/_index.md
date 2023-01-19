---
title: "FAQ"
linkTitle: "FAQ"
weight: 3
description: >
  Frequently Asked Questions
---


## Can I run an older Kubernetes version with minikube? Do I have to downgrade my minikube version?

You do not need to download an older minikube to run an older kubernetes version.
You can create a Kubernetes cluster with any version you desire using `--kubernetes-version` flag.

Example:

```bash
minikube start --kubernetes-version=v1.15.0
```

## How can I create more than one cluster with minikube?

By default, `minikube start` creates a cluster named "minikube". If you would like to create a different cluster or change its name, you can use the `--profile` (or `-p`) flag, which will create a cluster with the specified name. Please note that you can have multiple clusters on the same machine.

To see the list of your current clusters, run:
```
minikube profile list
```

minikube profiles are meant to be isolated from one another, with their own settings and drivers. If you want to create a single cluster with multiple nodes, try the [multi-node feature]({{< ref "/docs/tutorials/multi_node" >}}) instead.

## Can I use minikube as a Docker Desktop replacement?

Yes! Follow our tutorial on [Using minikube as a Docker Desktop Replacement]({{< ref "/docs/tutorials/docker_desktop_replacement" >}}).

## Can I start minikube without Kubernetes running?

Yes! If you want to use minikube only as a Docker Desktop replacement without starting Kubernetes itself, try:
```
minikube start --container-runtime=docker --no-kubernetes
```

Alternatively, if you want to temporarily turn off Kubernetes, you can pause and later unpause Kubernetes 
```
minikube pause
```

minikube also has an addon that automatically pauses Kubernetes after a certain amount of inactivity:

```
minikube addons enable auto-pause
```



## Docker Driver: How can I set minikube's cgroup manager?

For non-VM and non-SSH drivers, minikube will try to auto-detect your system's cgroups driver/manager and configure all other components accordingly.
For VM and SSH drivers, minikube will use cgroupfs cgroups driver/manager by default.
To force the `systemd` cgroup manager, run:

```bash
minikube start --force-systemd=true
```

## How can I run minikube with the Docker driver if I have an existing cluster with a VM driver?

First please ensure your Docker service is running. Then you need to either:  

(a) Delete the existing cluster and create a new one

```bash
minikube delete
minikube start --driver=docker
```

Alternatively, (b) Create a second cluster with a different profile name:

```bash
minikube start -p p1 --driver=docker 
```

## Does minikube support IPv6?

minikube currently doesn't support IPv6. However, it is on the [roadmap]({{< ref "/docs/contrib/roadmap.en.md" >}}). You can also refer to the [open issue](https://github.com/kubernetes/minikube/issues/8535).

## How can I prevent password prompts on Linux?

The easiest approach is to use the `docker` driver, as the backend service always runs as `root`.

`none` users may want to try `CHANGE_MINIKUBE_NONE_USER=true`, where kubectl and such will work without `sudo`. See [environment variables]({{< ref "/docs/handbook/config.md#environment-variables" >}}) for more details.  

Alternatively, you can configure `sudo` to never prompt for commands issued by minikube.

## How can I ignore system verification?

[kubeadm](https://github.com/kubernetes/kubeadm), minikube's bootstrapper, verifies a list of features on the host system before installing Kubernetes. In the case you get an error and still want to try minikube despite your system's limitation, you can skip verification by starting minikube with this extra option:

```shell
minikube start --extra-config kubeadm.ignore-preflight-errors=SystemVerification
```

## What is the minimum resource allocation necessary for a Knative setup using minikube?

Please allocate sufficient resources for Knative setup using minikube, especially when running minikube cluster on your local machine. We recommend allocating at least 3 CPUs and 3G memory.

```shell
minikube start --cpus 3 --memory 3072
```

## Do I need to install kubectl locally?

No, minikube comes with a built-in kubectl installation. See [minikube's kubectl documentation]({{< ref "docs/handbook/kubectl.md" >}}).

## How can I opt-in to beta release notifications?

Simply run the following command to be enrolled into beta notifications:
```
minikube config set WantBetaUpdateNotification true
```

## Can I get rid of the emoji in minikube's output?

Yes! If you prefer not having emoji in your minikube output 😔 , just set the `MINIKUBE_IN_STYLE` environment variable to `0` or `false`:

```
MINIKUBE_IN_STYLE=0 minikube start

```

## How can I access a minikube cluster from a remote network?

minikube's primary goal is to quickly set up local Kubernetes clusters, and therefore we strongly discourage using minikube in production or for listening to remote traffic. By design, minikube is meant to only listen on the local network.

However, it is possible to configure minikube to listen on a remote network. This will open your network to the outside world and is not recommended. If you are not fully aware of the security implications, please avoid using this.

For the docker and podman driver, use `--listen-address` flag:

```
minikube start --listen-address=0.0.0.0
```

## How can I allocate maximum resources to minikube?

Setting the `memory` and `cpus` flags on the start command to `max` will use maximum available resources:
```
minikube start --memory=max --cpus=max
```

## How can I run minikube on a different hard drive?

Set the `MINIKUBE_HOME` env to a path on the drive you want minikube to run, then run `minikube start`.

```
# Unix
export MINIKUBE_HOME=/otherdrive/.minikube

# Windows
$env:MINIKUBE_HOME = "D:\.minikube"

minikube start
```

## Can I set a static IP for the minikube cluster?

Currently a static IP can only be set when using the Docker or Podman driver.

For more details see the [static IP tutorial]({{< ref "docs/tutorials/static_ip.md" >}}).

## How to ignore the kubeadm requirements and pre-flight checks (such as minimum CPU count)?

Kubeadm has certain software and hardware requirements to maintain a stable Kubernetes cluster. However, these requirements can be ignored (such as when running minikube on a single CPU) by running the following:
```
minikube start --force --extra-config=kubeadm.skip-phases=preflight
```
This is not recommended, but for some users who are willing to accept potential performance or stability issues, this may be the only option.

