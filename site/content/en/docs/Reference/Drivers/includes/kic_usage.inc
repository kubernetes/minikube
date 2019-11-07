## Requirements

- Docker

## Docker installation

Please refer to https://docs.docker.com/install/


## Usage

Start a cluster using the kic driver:

```shell
minikube start --vm-driver=kic
```

To make kic the default driver:

```shell
minikube config set vm-driver kic
```
