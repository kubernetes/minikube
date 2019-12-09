---
title: "Using the Docker daemon"
linkTitle: "Using the Docker daemon"
weight: 6
date: 2018-08-02
description: >
  How to access the Docker daemon within minikube
---

## Prerequisites

You must be using minikube with the container runtime set to Docker. This is the default setting.

## Method 1: Without minikube registry addon

When using a single VM of Kubernetes it's really handy to reuse the Docker daemon inside the VM; as this means you don't have to build on your host machine and push the image into a docker registry - you can just build inside the same docker daemon as minikube which speeds up local experiments.

To be able to work with the docker daemon on your mac/linux host use the docker-env command in your shell:

```shell
eval $(minikube docker-env)
```

You should now be able to use docker on the command line on your host mac/linux machine talking to the docker daemon inside the minikube VM:

```shell
docker ps
```

Remember to turn off the _imagePullPolicy:Always_, as otherwise Kubernetes won't use images you built locally.

### Possible errors and solutions

Docker may report following forbidden error if you are using http proxy and the `$(minikube ip)` is not added to `no_proxy`/`NO_PROXY`:

```
error during connect: Get https://192.168.39.98:2376/v1.39/containers/json: Forbidden
```

On Centos 7, docker may report the following error:

```
Could not read CA certificate "/etc/docker/ca.pem": open /etc/docker/ca.pem: no such file or directory
```

The fix is to update ``/etc/sysconfig/docker`` to ensure that minikube's environment changes are respected:

```diff
< DOCKER_CERT_PATH=/etc/docker
---
> if [ -z "${DOCKER_CERT_PATH}" ]; then
>   DOCKER_CERT_PATH=/etc/docker
> fi
```

When you're using a docker installed via `snap` on a distribution like Ubuntu that uses AppArmor profiles the following error may appear:

```
could not read CA certificate "/home/USERNAME/.minikube/certs/ca.pem": open /home/USERNAME/.minikube/certs/ca.pem: permission denied
```

The solution is to allow docker to read the minikube certificates by adding a line in ``/var/lib/snapd/apparmor/profiles/snap.docker.docker`` file:

```shell
# allow docker to read minikube certificates
owner @{HOME}/.minikube/certs/* r,
```

After that check for syntax errors and try again:

```shell
sudo apparmor_parser -r /var/lib/snapd/apparmor/profiles/snap.docker.docker
eval $(minikube docker-env)
docker ps
```

##  Related Documentation

- [docker_registry.md](Using the Docker registry)
