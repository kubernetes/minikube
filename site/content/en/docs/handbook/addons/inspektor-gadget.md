---
title: "Using the Inspektor Gadget Addon"
linkTitle: "Inspektor Gadget"
weight: 1
date: 2023-02-16
---

## Inspektor Gadget Addon

[Inspektor Gadget](https://github.com/inspektor-gadget/inspektor-gadget)  is a collection of tools (or gadgets) to debug and inspect Kubernetes resources and applications. It manages the packaging, deployment and execution of [eBPF](https://ebpf.io/) programs in a Kubernetes cluster, including many based on [BCC](https://github.com/iovisor/bcc) tools, as well as some developed specifically for use in Inspektor Gadget. It automatically maps low-level kernel primitives to high-level Kubernetes resources, making it easier and quicker to find the relevant information.

### Enable Inspektor Gadget on minikube

To enable this addon, simply run:
```shell script
minikube addons enable inspektor-gadget
```

### Testing installation

```shell script
kubectl get pods -n gadget
```

If everything went well, there should be no errors about Inspektor Gadget's installation in your minikube cluster.

### Disable Inspektor Gadget

To disable this addon, simply run:

```shell script
minikube addons disable inspektor-gadget
```
