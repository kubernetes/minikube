## Requirements

A Linux VM with the following:

* Docker
* systemd (OpenRC based systems are also supported in v1.10+)

This VM must also meet the [kubeadm requirements.](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

Kubernetes v1.18+ needs conntract to be installed on debian based machines.

```shell
    sudo apt-get install conntrack -y
```

## Usage

The none driver requires minikube to be run as root, until [#3760](https://github.com/kubernetes/minikube/issues/3760) can be addressed.

```shell
sudo -E minikube start --driver=none
```

To make `none` the default driver:

```shell
sudo minikube config set driver none
```
