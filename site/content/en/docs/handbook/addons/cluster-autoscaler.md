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

- A running stable Kubernetes cluster (v1.35 or newer) with one control-plane node. The installed provider validates the matching autoscaler image.
- A cluster-wide CNI, such as `minikube start --cni=kindnet`, even when initially starting with one node. A single-node bridge-only default does not initialize networking for later elastic workers.
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

Before first enable, set `MINIKUBE_AUTOSCALER_MAX_WORKERS` to an integer from
1 through 16 to choose a different ceiling for additional workers. For a
one-to-four-node cluster, start with one base node and set this value to `3`.
The explicit ceiling must fit the host memory budget; it is never silently
reduced. Enabled pools retain their saved configuration, and a conflicting limit
on re-enable fails until the existing pool configuration is reconciled.

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

### Restoring a managed control plane

A configured provider retains its control-plane namespace and Node identities.
If an existing control plane fails to restart, `minikube start` returns the
original error and preserves its state. This applies while the addon is enabled
and while its saved configuration is disabled.

When boot-local configuration is missing and the persistent etcd database remains,
restore the retained configuration before retrying. A failed data-presence probe
also holds initialization. A fresh managed profile without retained etcd data
continues through normal initialization.

### Restarting managed VM workers

An existing worker in a provider-managed VM profile retains its Kubernetes Node
identity during restart. Minikube restores a missing boot-local kubeconfig from the worker's retained
rotating client certificate, updates and starts its kubelet, waits for the
worker to become Ready, and leaves destructive teardown and rejoin outside this
managed resume path. The provider continues to validate its original UID and
ownership metadata before admitting scaling operations.

A profile whose workers already rejoined under the legacy VM restart behavior
retains its paused ownership journal for explicit reconciliation. Preserve the
original state and data while preparing that recovery.
