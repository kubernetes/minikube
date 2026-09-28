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

This addon requires the installed autoscaler project, Docker, Helm and kubectl
(plus Bash and jq on macOS/Linux, or PowerShell on Windows). It supports
Kubernetes 1.35.x with QEMU and socket_vmnet on macOS, KVM2 on Linux amd64, or
Docker on Linux arm64 and Windows. Use Linux containers in Docker Desktop.

### Enable Cluster Autoscaler on minikube

```shell
minikube addons enable cluster-autoscaler
```

Minikube configures and starts the addon automatically. By default, it allows up
to two additional workers, limited by your host's memory. No separate terminal
or configuration file is needed.

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
