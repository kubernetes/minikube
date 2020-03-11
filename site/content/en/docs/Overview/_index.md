---
title: "Overview"
linkTitle: "Overview"
weight: 1
description: >
  What is it?
---

minikube implements a local Kubernetes cluster on macOS, Linux, and Windows.

minikube's [primary goals](https://minikube.sigs.k8s.io/docs/concepts/principles/) are to be the best tool for local Kubernetes application development and to support all Kubernetes features that fit.

minikube runs the latest stable release of Kubernetes, with support for standard Kubernetes features like:

* [LoadBalancer](https://minikube.sigs.k8s.io/docs/tasks/loadbalancer/) - using `minikube tunnel`
* Multi-cluster - using `minikube start -p <name>`
* NodePorts - using `minikube service`
* [Persistent Volumes](https://minikube.sigs.k8s.io/docs/reference/persistent_volumes/)
* Ingress
* RBAC
* [Dashboard](https://minikube.sigs.k8s.io/docs/tasks/dashboard/) - `minikube dashboard`
* [Container runtimes](https://minikube.sigs.k8s.io/docs/reference/runtimes/) - `start --container-runtime`
* [Configure apiserver and kubelet options](https://minikube.sigs.k8s.io/docs/reference/configuration/kubernetes/) via command-line flags

As well as developer-friendly features:

* [Addons](https://minikube.sigs.k8s.io/docs/tasks/addons/) - a marketplace for developers to share configurations for running services on minikube
* [GPU support](https://minikube.sigs.k8s.io/docs/tutorials/nvidia_gpu/) - for machine learning
* [Filesystem mounts](https://minikube.sigs.k8s.io/docs/tasks/mount/)
* Automatic failure analysis

## Why do I want it?

If you would like to develop Kubernetes applications:

* locally
* offline
* using the latest version of Kubernetes

Then minikube is for you.

* **What is it good for?** Developing local Kubernetes applications
* **What is it not good for?** Production Kubernetes deployments

## Where should I go next?

* [Getting Started](/docs/start/): Get started with minikube
* [Examples](/docs/examples/): Check out some minikube examples!

📣😀 **Please fill out our [fast 5-question survey](https://forms.gle/Gg3hG5ZySw8c1C24A)** so that we can learn how & why you use minikube, and what improvements we should make. Thank you! 💃🏽🎉
