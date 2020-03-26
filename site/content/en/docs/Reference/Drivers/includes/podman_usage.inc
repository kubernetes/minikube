## experimental

This is an experimental driver. please use it only for experimental reasons.
for a better kubernetes in container experience, use docker [driver](https://5e7c6ab90d754e000860cdfd--kubernetes-sigs-minikube.netlify.com/docs/reference/drivers/docker/).

## Install Podman

- [Podman](https://podman.io/getting-started/installation.html)

## Usage

Start a cluster using the docker driver:

```shell
minikube start --driver=podman
```
To make docker the default driver:

```shell
minikube config set driver podman
```
