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

Remember to turn off the `imagePullPolicy:Always` (use `imagePullPolicy:IfNotPresent` or `imagePullPolicy:Never`), as otherwise Kubernetes won't use images you built locally.

### Example

```shell
$ cat Containerfile 
FROM busybox
CMD exec /bin/sh -c "trap : TERM INT; (while true; do sleep 1000; done) & wait"
$ eval $(minikube podman-env)
$ podman-remote build -t example.com/test:v1 .
STEP 1: FROM busybox
STEP 2: CMD exec /bin/sh -c "trap : TERM INT; (while true; do sleep 1000; done) & wait"
STEP 3: COMMIT example.com/test:v1
2881381f7b9675ea5a0e635605bc0c4c08857582990bcadf0685b9f8976de2d3
$ minikube ssh -- sudo crictl images example.com/test:v1
IMAGE               TAG                 IMAGE ID            SIZE
example.com/test    v1                  2881381f7b967       1.44MB
$ kubectl run test --image example.com/test:v1 --image-pull-policy=IfNotPresent
kubectl run --generator=deployment/apps.v1 is DEPRECATED and will be removed in a future version. Use kubectl run --generator=run-pod/v1 or kubectl create instead.
deployment.apps/test created
$ kubectl get pods
NAME                   READY   STATUS    RESTARTS   AGE
test-d98bdbfdd-lwnqz   1/1     Running   0          18s

```

##  Related Documentation

- [Using the Docker registry]({{< ref "/docs/tasks/docker_registry" >}})
