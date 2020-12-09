---
title: "How to use KubeVirt with minikube"
linkTitle: "KubeVirt support"
weight: 1
date: 2020-05-26
description: >
  Using KubeVirt with minikube
---

## Prerequisites

- kvm2 driver

### Enable KubeVirt on minikube

Minikube can be started with default values and those will be enough to run a quick example, that being said, if you can spare a few more GiBs of RAM (by default it uses 2GiB), it’ll allow you to experiment further.

We’ll create a profile for KubeVirt so it gets its own settings without interfering what any configuration you might have already, let’s start by increasing the default memory to 4GiB:

```shell script
minikube config -p kubevirt set memory 4096
minikube config -p kubevirt set vm-driver kvm2
minikube start -p kubevirt
```

To enable this addon, simply run:
```shell script
minikube addons enable kubevirt
```

In a minute or so kubevirt's default components will be installed into your cluster.
You can run `kubectl get pods -n kubevirt` to see the progress of the kubevirt installation.

### Disable KubeVirt

To disable this addon, simply run:
```shell script
minikube addons disable kubevirt
```

### More Information

See official [Minikube Quickstart](https://kubevirt.io/quickstart_minikube/) documentation for more information and to install KubeVirt without using the addon.
