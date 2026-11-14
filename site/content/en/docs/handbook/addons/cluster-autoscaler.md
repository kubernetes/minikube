---
title: "Using the Cluster Autoscaler Addon"
linkTitle: "Cluster Autoscaler"
weight: 1
date: 2026-09-27
---

## Cluster Autoscaler Addon

[Cluster Autoscaler](https://github.com/astrivant/minikube-cluster-autoscaler-addon)
adds worker nodes when Pods need more capacity and removes idle workers.
Existing nodes are kept.

### Prerequisites

- A running Kubernetes 1.35.x cluster with one control-plane node.
- Docker running, with Git, Go (the version in the addon's
  [`.go-version`](https://github.com/astrivant/minikube-cluster-autoscaler-addon/blob/main/.go-version)),
  Helm and kubectl installed.
- Bash and jq on macOS/Linux, or PowerShell on Windows.

Supported drivers are QEMU with socket_vmnet on macOS, KVM2 on Linux amd64,
and Docker on Linux arm64 and Windows. Use Linux containers in Docker Desktop.
See [host and cluster setup](https://github.com/astrivant/minikube-cluster-autoscaler-addon/blob/main/docs/minikube-integration.md#installation)
for driver setup and cluster creation.

### Enable Cluster Autoscaler on minikube

Clone the host components once, then enable the addon (works in Bash and PowerShell):

```shell
git clone --depth 1 https://github.com/astrivant/minikube-cluster-autoscaler-addon.git "$HOME/.minikube/addons/cluster-autoscaler"
minikube addons enable cluster-autoscaler
```

Minikube discovers this installation, builds the components, and configures and
starts the addon automatically. By default, it allows up to two additional
workers, limited by your host's memory. No separate terminal or configuration
file is needed.

These commands use the default Minikube home directory. For custom paths or
prebuilt release bundles, see the project's
[installation guide](https://github.com/astrivant/minikube-cluster-autoscaler-addon#quick-install).

### Testing installation

```shell
kubectl get pods -n kube-system -l app.kubernetes.io/instance=minikube-cluster-autoscaler-addon
kubectl get nodes -w
```

New workers appear when workloads cannot fit on the existing nodes.

### Disable Cluster Autoscaler

```shell
minikube addons disable cluster-autoscaler
```

This stops autoscaling and its background processes. Existing nodes are retained.
