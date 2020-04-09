---
title: "Pushing images"
weight: 5
description: >
 comparing 3 ways to push your image into minikiube.
aliases:
 - /docs/tasks/building
 - /docs/tasks/caching
 - /docs/tasks/podman_service
 - /docs/tasks/docker_daemon
---

## What is best method to push image to minikiube ?
the answer depends on the container-runtime driver you choose. 
Here is a comparison table to help you choose:



| Method   	| Supported Runtimes   	|  Supported Drivers* 	|  Performance 	|
|---	|---	|---	|---	|---	|
|  [docker-env command](http://localhost:1313/docs/handbook/pushing/#pushing-directly-to-the-in-cluster-docker-daemon)	|   only docker	|  all	|  good 	|
|  [podman-env command](http://localhost:1313/docs/handbook/pushing/#pushing-directly-to-in-cluster-crio)	|   only cri-o	|   all  |  good 	|
|  [cache add command](http://localhost:1313/docs/handbook/pushing/#push-images-using-cache-command) 	|  all 	| all   	|  ok 	|
|  [registry addon](http://localhost:1313/docs/handbook/pushing/#pushing-to-an-in-cluster-using-a-registry-addon)   |   all	|   all but [docker on mac](https://github.com/kubernetes/minikube/issues/7535) |  ok 	|
|  [minikube ssh](http://localhost:1313/docs/handbook/pushing/#building-images-inside-of-minikube-using-ssh)   |   all	|   all |  best 	|


* note1 : the default container-runtime on minikube is 'docker'.
* note2 : 'none' driver (bare metal) does not need pushing image to the cluster, as any image on your system is already available to the kuberentes.


# Pushing directly to the in-cluster Docker daemon
When using a container or VM driver (all drivers except none), you can reuse the Docker daemon inside minikube cluster.
this means you don't have to build on your host machine and push the image into a docker registry. You can just build inside the same docker daemon as minikube which speeds up local experiments.

To point your terminal to use the docker daemon inside minikube run this:

```shell
eval $(minikube docker-env)
```

now any 'docker' command you run in this current terminal will run against the docker inside minikube VM or Container.
Try it:

```shell
docker ps
```

now you 'build' against the docker inside minikube. which is instantly accessible to kubernetes cluster.

'''
docker build -t myimage .
'''

Remember to turn off the `imagePullPolicy:Always` (use `imagePullPolicy:IfNotPresent` or `imagePullPolicy:Never`), as otherwise Kubernetes won't use images you built locally.

#### docker-env will be 
Please remember by closing the terminal, you will go back to using your own system's docker daemon.
to verify your terminal is using minikuber's docker-env you can check the value of the environment MINIKUBE_ACTIVE_DOCKERD

more information on [docker-env](https://minikube.sigs.k8s.io/docs/commands/docker-env/)

# Push images using Cache Command.

From the host, you can push a Docker image directly to minikube. It will also be cached for all minikube clusters .

```shell
minikube cache add ubuntu:16.04
```

The add command will store the requested image to `$MINIKUBE_HOME/cache/images`, and load it into the VM's container runtime environment next time `minikube start` is called.

To display images you have added to the cache:

```shell
minikube cache list
```

This listing will not include the images minikube's built-in system images.

to ensure your running cluster has cached the updated images use reload:

```shell
minikube cache reload
```


```shell
minikube cache delete <image name>
```

For more information, see:

* [Reference: cache command]({{< ref "/docs/commands/cache.md" >}})


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

# Pushing to an in-cluster using a Registry addon

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

# Building images inside of minikube using SSH

Use `minikube ssh` to go inside a minikube node, and run the `docker build` directly there.
any command you run there will run against the same daemon that kubernetes is using.

```shell
docker build
```

For more information on the `docker build` command, read the [Docker documentation](https://docs.docker.com/engine/reference/commandline/build/) (docker.com).

For Podman, use:

```shell
sudo -E podman build
```

For more information on the `podman build` command, read the [Podman documentation](https://github.com/containers/libpod/blob/master/docs/source/markdown/podman-build.1.md) (podman.io).

to exit minikube ssh and come back to your terminal type:
```shell
exit
```
