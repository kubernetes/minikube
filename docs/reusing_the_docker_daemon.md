# Reusing the Docker daemon

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

Docker may report following forbidden error if you are using http proxy and the `$(minikube ip)` is not added to `no_proxy`/`NO_PROXY`:

```shell
error during connect: Get https://192.168.39.98:2376/v1.39/containers/json: Forbidden
```

On Centos 7, docker may report the following error:

```shell
Could not read CA certificate "/etc/docker/ca.pem": open /etc/docker/ca.pem: no such file or directory
```

The fix is to update /etc/sysconfig/docker to ensure that minikube's environment changes are respected:

```diff
< DOCKER_CERT_PATH=/etc/docker
---
> if [ -z "${DOCKER_CERT_PATH}" ]; then
>   DOCKER_CERT_PATH=/etc/docker
> fi
```

Remember to turn off the _imagePullPolicy:Always_, as otherwise Kubernetes won't use images you built locally.

## Method 2: With minikube registry addon

Enable minikube registry addon and then push images directly into registry. Steps are as follows:

For illustration purpose, we will assume that minikube VM has one of the ip from `192.168.39.0/24` subnet. If you have not overridden these subnets as per [networking guide](https://github.com/kubernetes/minikube/blob/master/docs/networking.md), you can find out default subnet being used by minikube for a specific OS and driver combination [here](https://github.com/kubernetes/minikube/blob/dfd9b6b83d0ca2eeab55588a16032688bc26c348/pkg/minikube/cluster/cluster.go#L408) which is subject to change. Replace `192.168.39.0/24` with appropriate values for your environment wherever applicable.

Ensure that docker is configured to use `192.168.39.0/24` as insecure registry. Refer [here](https://docs.docker.com/registry/insecure/) for instructions.

Ensure that `192.168.39.0/24` is enabled as insecure registry in minikube. Refer [here](https://github.com/kubernetes/minikube/blob/master/docs/insecure_registry.md) for instructions..

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
