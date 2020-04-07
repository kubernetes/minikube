---
title: "Pushing images"
weight: 5
description: >
  There are many ways to push images into minikube.
aliases:
 - /docs/tasks/building
 - /docs/tasks/caching
 - /docs/tasks/podman_service
 - /docs/tasks/docker_daemon
---

# Cached Images

From the host, you can push a Docker image directly to minikube. It will also be cached for future cluster starts.

```shell
minikube cache add ubuntu:16.04
```

The add command will store the requested image to `$MINIKUBE_HOME/cache/images`, and load it into the VM's container runtime environment next time `minikube start` is called.

To display images you have added to the cache:

```shell
minikube cache list
```

This listing will not include the images which are built-in to minikube.


```shell
minikube cache delete <image name>
```

For more information, see:

* [Reference: cache command]({{< ref "/docs/commands/cache.md" >}})

You must be using minikube with the container runtime set to Docker. This is the default setting.

# Pushing directly to the in-cluster Docker daemon
When user a container or VM driver, it's really handy to reuse the Docker daemon inside minikube; as this means you don't have to build on your host machine and push the image into a docker registry - you can just build inside the same docker daemon as minikube which speeds up local experiments.

To point your terminal to use the docker daemon inside minikube run this:

```shell
eval $(minikube docker-env)
```

now any command you run in this current terminal will run against the docker inside minikube VM or Container.
Try it:

```shell
docker ps
```

now you can use same docker build command against the docker inside minikube. which is instantly accessible to kubernetes cluster.

'''
docker build -t myimage .
'''


Remember to turn off the `imagePullPolicy:Always` (use `imagePullPolicy:IfNotPresent` or `imagePullPolicy:Never`), as otherwise Kubernetes won't use images you built locally.

more information on [docker-env](https://minikube.sigs.k8s.io/docs/commands/docker-env/)

# Pushing directly to in-cluster CRIO

To push directly to CRIO, configure podman client on your mac/linux host using the podman-env command in your shell:

```shell
eval $(minikube podman-env)
```

You should now be able to use podman on the command line on your host mac/linux machine talking to the podman service inside the minikube VM:

```shell
podman-remote help
```

Remember to turn off the `imagePullPolicy:Always` (use `imagePullPolicy:IfNotPresent` or `imagePullPolicy:Never`), as otherwise Kubernetes won't use images you built locally.

# Pushing to an in-cluster Registry

For illustration purpose, we will assume that minikube VM has one of the ip from `192.168.39.0/24` subnet. If you have not overridden these subnets as per [networking guide](https://minikube.sigs.k8s.io/reference/networking/), you can find out default subnet being used by minikube for a specific OS and driver combination [here](https://github.com/kubernetes/minikube/blob/dfd9b6b83d0ca2eeab55588a16032688bc26c348/pkg/minikube/cluster/cluster.go#L408) which is subject to change. Replace `192.168.39.0/24` with appropriate values for your environment wherever applicable.

Ensure that docker is configured to use `192.168.39.0/24` as insecure registry. Refer [here](https://docs.docker.com/registry/insecure/) for instructions.

Ensure that `192.168.39.0/24` is enabled as insecure registry in minikube. Refer [here](https://minikube.sigs.k8s.io/Handbook/registry/insecure/) for instructions..

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

# Building images inside of minikube

Use `minikube ssh` to connect to the virtual machine, and run the `docker build` there:

```shell
docker build
```

For more information on the `docker build` command, read the [Docker documentation](https://docs.docker.com/engine/reference/commandline/build/) (docker.com).

For Podman, use:

```shell
sudo -E podman build
```

For more information on the `podman build` command, read the [Podman documentation](https://github.com/containers/libpod/blob/master/docs/source/markdown/podman-build.1.md) (podman.io).

