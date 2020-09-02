## Requirements

A Linux VM with the following:

* systemd or OpenRC
* a container runtime, such as Docker or CRIO

This VM must also meet the [kubeadm requirements](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/), such as:

* 2 CPU's
* 2GB RAM
* iptables (in legacy mode)
* conntrack
* crictl
* SELinux permissive
* cgroups v1 (v2 is not yet supported by Kubernetes)

## Usage

The none driver requires minikube to be run as root, until [#3760](https://github.com/kubernetes/minikube/issues/3760) can be addressed.

```shell
sudo -E minikube start --driver=none
```

To make `none` the default driver:

```shell
sudo minikube config set driver none
```
