---
title: "Using the Docker registry"
linkTitle: "Using the Docker registry"
weight: 6
date: 2018-08-02
description: >
  How to access the Docker registry within minikube
---

As an alternative to [reusing the Docker daemon]({{< ref "/docs/tasks/docker_daemon.md" >}}), you may enable the registry addon to push images directly into registry.

Steps are as follows:

For illustration purpose, we will assume that minikube VM has one of the ip from `192.168.39.0/24` subnet. If you have not overridden these subnets as per [networking guide](https://minikube.sigs.k8s.io/docs/reference/networking/), you can find out default subnet being used by minikube for a specific OS and driver combination [here](https://github.com/kubernetes/minikube/blob/dfd9b6b83d0ca2eeab55588a16032688bc26c348/pkg/minikube/cluster/cluster.go#L408) which is subject to change. Replace `192.168.39.0/24` with appropriate values for your environment wherever applicable.

Ensure that docker is configured to use `192.168.39.0/24` as insecure registry. Refer [here](https://docs.docker.com/registry/insecure/) for instructions.

Ensure that `192.168.39.0/24` is enabled as insecure registry in minikube. Refer [here](https://minikube.sigs.k8s.io/docs/tasks/registry/insecure/) for instructions..

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

Now run it in minikube:

```shell
kubectl run test-img --image=$(minikube ip):5000/test-img
```

Or if `192.168.39.0/24` is not enabled as insecure registry in minikube, then:

```shell
kubectl run test-img --image=localhost:5000/test-img
```
