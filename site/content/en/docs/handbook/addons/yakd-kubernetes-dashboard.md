---
title: "Using the YAKD - Kubernetes Dashboard Addon"
linkTitle: "YAKD - Kubernetes Dashboard"
weight: 1
date: 2023-12-12
---

## YAKD - Kubernetes Dashboard Addon

[YAKD - Kubernetes Dashboard](https://github.com/manusa/yakd) is a full-featured web-based Kubernetes Dashboard with special functionality for minikube.

The dashboard features a real-time Search pane that allows you to search for Kubernetes resources and see them update in real-time as you type.

### Enable YAKD - Kubernetes Dashboard on minikube

To enable this addon, simply run:

```shell script
minikube addons enable yakd
```

Once the addon is enabled, you can access the YAKD - Kubernetes Dashboard's web UI using the following command.

```shell script
minikube service yakd-dashboard -n yakd-dashboard
```

The dashboard will open in a new browser window and you should be able to start using it with no further hassle.

YAKD - Kubernetes Dashboard is also compatible with metrics-server. To install it, run:

```shell script
minikube addons enable metrics-server	
```

### Disable YAKD - Kubernetes Dashboard

To disable this addon, simply run:

```shell script
minikube addons disable yakd
```
