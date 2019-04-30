# Configuring Kubernetes

Minikube has a "configurator" feature that allows users to configure the Kubernetes components with arbitrary values.
To use this feature, you can use the `--extra-config` flag on the `minikube start` command.

This flag is repeated, so you can pass it several times with several different values to set multiple options.


## Selecting a Kubernetes version

minikube defaults to the latest stable version of Kubernetes. If you need something different, you can use the `--kubernetes-version` flag, for instance:

  * `--kubernetes-version=v1.10.13'
  
minikube follows the [Kubernetes Version and Version Skew Support Policy](https://kubernetes.io/docs/setup/version-skew-policy/), though we also make also make every effort possible to support two additional minor releases. As of April 2019, minikube supports the latest builds of:

* v1.14 (default)
* v1.13
* v1.12
* v1.11 (best effort)
* v1.10 (best effort)

For more up to date information, see `OldestKubernetesVersion` and `NewestKubernetesVersion` in [constants.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/constants/constants.go)

## kubeadm

The kubeadm bootstrapper can be configured by the `--extra-config` flag on the `minikube start` command.  It takes a string of the form `component.key=value` where `component` is one of the strings

* kubeadm
* kubelet
* apiserver
* controller-manager
* scheduler

and `key=value` is a flag=value pair for the component being configured.  For example,

```shell
minikube start --extra-config=apiserver.v=10 --extra-config=kubelet.max-pods=100

minikube start --extra-config=kubeadm.ignore-preflight-errors=SystemVerification # allows any version of docker
```
