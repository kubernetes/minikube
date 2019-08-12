---
title: "Overview"
linkTitle: "Overview"
weight: 1
description: >
  What is it?
---

minikube implements a local Kubernetes cluster on macOS, Linux, and Windows.

minikube's [primary goals](https://github.com/kubernetes/minikube/blob/master/docs/contributors/principles.md) are to be the best tool for local Kubernetes application development and to support all Kubernetes features that fit.

minikube runs the latest stable release of Kubernetes, with support for standard Kubernetes features like:

* [LoadBalancer](https://github.com/kubernetes/minikube/blob/master/docs/tunnel.md) - using `minikube tunnel`
* Multi-cluster - using `minikube start -p <name>`
* NodePorts - using `minikube service`
* [Persistent Volumes](https://github.com/kubernetes/minikube/blob/master/docs/persistent_volumes.md)
* Ingress
* RBAC
* [Dashboard](https://github.com/kubernetes/minikube/blob/master/docs/dashboard.md) - `minikube dashboard`
* [Container runtimes](https://github.com/kubernetes/minikube/blob/master/docs/alternative_runtimes.md) - `start --container-runtime`
* [Configure apiserver and kubelet options](https://github.com/kubernetes/minikube/blob/master/docs/configuring_kubernetes.md) via command-line flags

As well as developer-friendly features:

* [Addons](https://github.com/kubernetes/minikube/blob/master/docs/addons.md) - a marketplace for developers to share configurations for running services on minikube
* [GPU support](https://github.com/kubernetes/minikube/blob/master/docs/gpu.md) - for machine learning
* [Filesystem mounts](https://github.com/kubernetes/minikube/blob/master/docs/host_folder_mount.md)
* Automatic failure analysis

## Why do I want it?

If you would like to develop Kubernetes applications:

* locally
* offline
* using the latest version of Kubernetes

Then minikube is for you.

* **What is it good for?** Developing local Kubernetes applications
* **What is it not good for?** Production Kubernetes deployments
* **What is it *not yet* good for?** Environments which do not allow VM's

## Where should I go next?

* [Getting Started](/start/): Get started with minikube
* [Examples](/examples/): Check out some minikube examples!
