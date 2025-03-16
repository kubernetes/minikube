## Experimental

This is an experimental driver. Please use it only for experimental reasons until it has reached maturity. For a more reliable minikube experience, use a non-experimental driver, like [Docker](https://minikube.sigs.k8s.io/docs/drivers/docker/).

## Usage

It's recommended to run minikube with the podman driver and [CRI-O container runtime](https://cri-o.io/) (except when using Rootless Podman):

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

## Rootless Podman

By default, minikube executes Podman with `sudo`.
To use Podman without `sudo` (i.e., Rootless Podman), set the `rootless` property to `true`:

```shell
minikube config set rootless true
```

For Rootless Podman, it is recommended to set `--container-runtime` to `containerd`:

```shell
minikube start --driver=podman --container-runtime=containerd
```

See the [Rootless Docker](https://minikube.sigs.k8s.io/docs/drivers/docker/#rootless-docker) section for the requirements and the restrictions.
