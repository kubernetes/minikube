## Requirements

- Docker

## Docker installation

Please refer to https://docs.docker.com/install/


## Usage

Start a cluster using the kic's docker driver:

```shell
minikube start --vm-driver=docker
```

To make kic the default driver:

```shell
minikube config set vm-driver docker
```
