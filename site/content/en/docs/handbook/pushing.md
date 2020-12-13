---
title: "Pushing images"
weight: 5
description: >
 comparing 5 ways to push your image into a minikiube cluster.
aliases:
 - /docs/tasks/building
 - /docs/tasks/caching
 - /docs/tasks/podman_service
 - /docs/tasks/docker_daemon
---


## Comparison table for different methods

The best method to push your image to minikube depends on the container-runtime you built your cluster with (the default is docker).
Here is a comparison table to help you choose:

| Method    | Supported Runtimes    |  |  Performance  |
|--- |--- |--- |--- |--- |
|  [docker-env command](/docs/handbook/pushing/#1pushing-directly-to-the-in-cluster-docker-daemon-docker-env) |   only docker |  good  |
|  [podman-env command](/docs/handbook/pushing/#3-pushing-directly-to-in-cluster-crio-podman-env) |   only cri-o |  good  |
|  [cache add command]({{< ref "/docs/commands/cache.md#minikube-cache-add" >}})  |  all  |  ok  |
|  [registry addon](/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon)   |   all |  ok  |
|  [minikube ssh](/docs/handbook/pushing/#5-building-images-inside-of-minikube-using-ssh)   |   all | best  |

* note1 : the default container-runtime on minikube is 'docker'.
* note2 : 'none' driver (bare metal) does not need pushing image to the cluster, as any image on your system is already available to the kuberentes.

---

## 1. Pushing directly to the in-cluster Docker daemon (docker-env)

This is similar to podman-env but only for Docker runtime.
When using a container or VM driver (all drivers except none), you can reuse the Docker daemon inside minikube cluster.
this means you don't have to build on your host machine and push the image into a docker registry. You can just build inside the same docker daemon as minikube which speeds up local experiments.

To point your terminal to use the docker daemon inside minikube run this:

```shell
eval $(minikube docker-env)
```

now any 'docker' command you run in this current terminal will run against the docker inside minikube cluster.

so if you do the following commands, it will show you the containers inside the minikube, inside minikube's VM or Container.

```shell
docker ps
```

now you can 'build' against the docker inside minikube. which is instantly accessible to kubernetes cluster.

```shell
docker build -t my_image .
```

To verify your terminal is using minikuber's docker-env you can check the value of the environment variable MINIKUBE_ACTIVE_DOCKERD to reflect the cluster name.

{{% pageinfo color="info" %}}
Tip 1:
Remember to turn off the `imagePullPolicy:Always` (use `imagePullPolicy:IfNotPresent` or `imagePullPolicy:Never`) in your yaml file. Otherwise Kubernetes won't use your locally build image and it will pull from the network.
{{% /pageinfo %}}

{{% pageinfo color="info" %}}
Tip 2:
Evaluating the docker-env is only valid for the current terminal.
By closing the terminal, you will go back to using your own system's docker daemon.
{{% /pageinfo %}}

{{% pageinfo color="info" %}}
Tip 3:
In container-based drivers such as Docker or Podman, you will need to re-do docker-env each time you restart your minikube cluster.
{{% /pageinfo %}}

more information on [docker-env](https://minikube.sigs.k8s.io/docs/commands/docker-env/)

---

## 2. Push images using 'cache' command.

From your host, you can push a Docker image directly to minikube. This image will be cached and automatically pulled into all future minikube clusters created on the machine

```shell
minikube cache add alpine:latest
```

The add command will store the requested image to `$MINIKUBE_HOME/cache/images`, and load it into the minikube cluster's container runtime environment automatically.

{{% pageinfo color="info" %}}
Tip 1 :
If your image changes after your cached it, you need to do 'cache reload'.
{{% /pageinfo %}}

minikube refreshes the cache images on each start. however to reload all the cached images on demand, run this command :
```shell
minikube cache reload
```

{{% pageinfo color="info" %}}
Tip 2 :
if you have multiple clusters, the cache command will load the image for all of them.
{{% /pageinfo %}}

To display images you have added to the cache:

```shell
minikube cache list
```

This listing will not include the images minikube's built-in system images.

```shell
minikube cache delete <image name>
```

For more information, see:

* [Reference: cache command]({{< ref "/docs/commands/cache.md" >}})

---

## 3. Pushing directly to in-cluster CRI-O. (podman-env)

This is similar to docker-env but only for CRI-O runtime.
To push directly to CRI-O, configure podman client on your host using the podman-env command in your shell:

```shell
eval $(minikube podman-env)
```

You should now be able to use podman client on the command line on your host machine talking to the podman service inside the minikube VM:

{{% tabs %}}
{{% linuxtab %}}

```shell
podman-remote help
```

now you can 'build' against the storage inside minikube. which is instantly accessible to kubernetes cluster.

```shell
podman-remote build -t my_image .
```

{{% pageinfo color="info" %}}
Note: On Linux the remote client is called "podman-remote", while the local program is called "podman".
{{% /pageinfo %}}

{{% /linuxtab %}}
{{% mactab %}}

```shell
podman help
```

now you can 'build' against the storage inside minikube. which is instantly accessible to kubernetes cluster.

```shell
podman build -t my_image .
```

{{% pageinfo color="info" %}}
Note: On macOS the remote client is called "podman", since there is no local "podman" program available.
{{% /pageinfo %}}

{{% /mactab %}}
{{% windowstab %}}

now you can 'build' against the storage inside minikube. which is instantly accessible to kubernetes cluster.

```shell
podman help
```

```shell
podman build -t my_image .
```

{{% pageinfo color="info" %}}
Note: On Windows the remote client is called "podman", since there is no local "podman" program available.
{{% /pageinfo %}}

{{% /windowstab %}}
{{% /tabs %}}

Remember to turn off the `imagePullPolicy:Always` (use `imagePullPolicy:IfNotPresent` or `imagePullPolicy:Never`), as otherwise Kubernetes won't use images you built locally.

---

## 4. Pushing to an in-cluster using Registry addon

For illustration purpose, we will assume that minikube VM has one of the ip from `192.168.39.0/24` subnet. If you have not overridden these subnets as per [networking guide](https://minikube.sigs.k8s.io/reference/networking/), you can find out default subnet being used by minikube for a specific OS and driver combination [here](https://github.com/kubernetes/minikube/blob/dfd9b6b83d0ca2eeab55588a16032688bc26c348/pkg/minikube/cluster/cluster.go#L408) which is subject to change. Replace `192.168.39.0/24` with appropriate values for your environment wherever applicable.

Ensure that docker is configured to use `192.168.39.0/24` as insecure registry. Refer [here](https://docs.docker.com/registry/insecure/) for instructions.

Ensure that `192.168.39.0/24` is enabled as insecure registry in minikube. Refer [here](https://minikube.sigs.k8s.io/docs/handbook/registry/#enabling-insecure-registries/) for instructions..

Enable minikube registry addon:

```shell
minikube addons enable registry
```

Build docker image and tag it appropriately:

```shell
docker build --tag $(minikube ip):5000/test-img .
```

Push docker image to minikube registry:

```shell
docker push $(minikube ip):5000/test-img
```

---

## 5. Building images inside of minikube using SSH

Use `minikube ssh` to run commands inside the minikube node, and run the build command directly there.
Any command you run there will run against the same daemon / storage that kubernetes cluster is using.

For Docker, use:

```shell
docker build
```

For more information on the `docker build` command, read the [Docker documentation](https://docs.docker.com/engine/reference/commandline/build/) (docker.com).

For CRI-O, use:

```shell
sudo podman build
```

For more information on the `podman build` command, read the [Podman documentation](https://github.com/containers/podman/blob/master/docs/source/markdown/podman-build.1.md) (podman.io).

For Containerd, use:

```shell
sudo buildctl build
```

For more information on the `buildctl build` command, read the [Buildkit documentation](https://github.com/moby/buildkit#quick-start) (mobyproject.org).

to exit minikube ssh and come back to your terminal type:

```shell
exit
```
