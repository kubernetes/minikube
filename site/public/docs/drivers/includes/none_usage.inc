## Requirements

A Linux VM with the following:

* systemd or OpenRC
* a Container Runtime Interface (CRI), such as Docker or CRI-O
* [containernetworking-plugins]({{<ref "/docs/faq/#how-do-i-install-containernetworking-plugins-for-none-driver" >}})
* [cri-dockerd](https://github.com/Mirantis/cri-dockerd) (Only if using Docker CRI)

This VM must also meet the [kubeadm requirements](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/), such as:

* 2 CPU's
* 2GB RAM
* iptables (in legacy mode)
* conntrack
* crictl
* cni-plugins
* SELinux permissive
* cgroups v1 (v2 is not yet supported by Kubernetes)

## Usage

```shell
minikube start --driver=none
```

To make `none` the default driver:

```shell
sudo minikube config set driver none
```
