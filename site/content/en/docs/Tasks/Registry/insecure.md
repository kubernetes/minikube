---
title: "Insecure"
linkTitle: "Insecure"
weight: 6
date: 2019-08-1
description: >
  How to enable insecure registry support within minikube
---

minikube allows users to configure the docker engine's `--insecure-registry` flag. 

You can use the `--insecure-registry` flag on the
`minikube start` command to enable insecure communication between the docker engine and registries listening to requests from the CIDR range.

One nifty hack is to allow the kubelet running in minikube to talk to registries deployed inside a pod in the cluster without backing them
with TLS certificates. Because the default service cluster IP is known to be available at 10.0.0.1, users can pull images from registries
deployed inside the cluster by creating the cluster with `minikube start --insecure-registry "10.0.0.0/24"`.

### docker on macOS

Quick guide for configuring minikube and docker on macOS, enabling docker to push images to minikube's registry.

The first step is to enable the registry addon:

```
minikube addons enable registry
```

When enabled, the registry addon exposes its port 5000 on the minikube's virtual machine.

In order to make docker accept pushing images to this registry, we have to redirect port 5000 on the docker virtual machine over to port 5000 on the minikube machine. We can (ab)use docker's network configuration to instantiate a container on the docker's host, and run socat there:

```
docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000"
```

Once socat is running it's possible to push images to the minikube registry:

```
docker tag my/image localhost:5000/myimage
docker push localhost:5000/myimage
```

After the image is pushed, refer to it by `localhost:5000/{name}` in kubectl specs.
