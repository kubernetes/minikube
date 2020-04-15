## Requirements

* VM running a systemd-based Linux distribution ([see #2704](https://github.com/kubernetes/minikube/issues/2704))
  - OpenRC systems are supported on master (to be released in minikube v1.10)
* Docker
* [Meets kubeadm requirements](hhttps://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#before-you-begin)

## Usage

The none driver requires minikube to be run as root, until [#3760](https://github.com/kubernetes/minikube/issues/3760) can be addressed.

```shell
sudo minikube start --driver=none
```

To make `none` the default driver:

```shell
sudo minikube config set driver none
```
