# Reusing the Docker daemon

When using a single VM of Kubernetes it's really handy to reuse the Docker daemon inside the VM; as this means you don't have to build on your host machine and push the image into a docker registry - you can just build inside the same docker daemon as minikube which speeds up local experiments.

To be able to work with the docker daemon on your mac/linux host use the docker-env command in your shell:

```shell
eval $(minikube docker-env)
```

You should now be able to use docker on the command line on your host mac/linux machine talking to the docker daemon inside the minikube VM:

```shell
docker ps
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
