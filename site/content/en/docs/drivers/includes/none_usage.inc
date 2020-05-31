## Requirements

A Linux VM with the following:

* Docker
* systemd (OpenRC based systems are also supported in v1.10+)
* conntrack (Kubernetes v1.18+ needs conntrack to be installed on the machine)

    ```shell
    sudo apt-get install conntrack -y
    ```

This VM must also meet the [kubeadm requirements](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

## Usage

The none driver requires minikube to be run as root, until [#3760](https://github.com/kubernetes/minikube/issues/3760) can be addressed.

```shell
sudo minikube start --driver=none
```

To make `none` the default driver:

```shell
sudo minikube config set driver none
```
