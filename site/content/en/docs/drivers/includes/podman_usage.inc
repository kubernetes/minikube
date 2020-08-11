## Experimental

This is an experimental driver. Please use it only for experimental reasons until it has reached maturity. For a more reliable minikube experience, use a non-experimental driver, like [Docker](https://minikube.sigs.k8s.io/docs/drivers/docker/).

## Usage

It's recommended to run minikube with the podman driver and [CRI-O container runtime](https://https://cri-o.io/): 

```shell
minikube start --driver=podman --container-runtime=cri-o
```

Alternatively, start minikube with the podman driver only: 

```shell
minikube start --driver=podman 
```

To make podman the default driver: 

```shell
minikube config set driver podman
```
