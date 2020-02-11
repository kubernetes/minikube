---
title: "Using the Podman service"
linkTitle: "Using the Podman service"
weight: 6
date: 2020-01-20
description: >
  How to access the Podman service within minikube
---

## Prerequisites

You should be using minikube with the container runtime set to CRI-O. It uses the same storage as Podman.

## Method 1: Without minikube registry addon

When using a single VM of Kubernetes it's really handy to reuse the Podman service inside the VM; as this means you don't have to build on your host machine and push the image into a container registry - you can just build inside the same container storage as minikube which speeds up local experiments.

To be able to work with the podman client on your mac/linux host use the podman-env command in your shell:

```shell
eval $(minikube podman-env)
```

You should now be able to use podman on the command line on your host mac/linux machine talking to the podman service inside the minikube VM:

```shell
podman-remote help
```

Remember to turn off the _imagePullPolicy:Always_, as otherwise Kubernetes won't use images you built locally.

##  Related Documentation

- [docker_registry.md](Using the Docker registry)
