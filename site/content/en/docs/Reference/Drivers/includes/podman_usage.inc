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
