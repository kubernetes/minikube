## Requirements

VM running a systemd-based Linux distribution ([see #2704](https://github.com/kubernetes/minikube/issues/2704))

## Usage

The none driver requires minikube to be run as root, until [#3760](https://github.com/kubernetes/minikube/issues/3760) can be addressed.

```shell
sudo minikube start --driver=none
```

To make `none` the default driver:

```shell
sudo minikube config set driver none
```
