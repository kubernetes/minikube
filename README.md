# minikube

[![BuildStatus Widget]][BuildStatus Result]
[![GoReport Widget]][GoReport Status]

[BuildStatus Result]: https://travis-ci.org/kubernetes/minikube
[BuildStatus Widget]: https://travis-ci.org/kubernetes/minikube.svg?branch=master

[GoReport Status]: https://goreportcard.com/report/github.com/kubernetes/minikube
[GoReport Widget]: https://goreportcard.com/badge/github.com/kubernetes/minikube

<img src="https://github.com/kubernetes/minikube/raw/master/images/logo/logo.png" width="100">

## What is minikube?

minikube implements a local Kubernetes cluster on macOS, Linux, and Windows.

![screenshot](/images/start.jpg)

minikube's [primary goals](https://minikube.sigs.k8s.io/docs/concepts/principles/) are to be the best tool for local Kubernetes application development and to support all Kubernetes features that fit. We hope you enjoy it!

## News

:mega: **Please fill out our [fast 5-question survey](https://forms.gle/Gg3hG5ZySw8c1C24A)** so that we can learn how & why you use minikube, and what improvements we should make. Thank you! :dancers:

* 2019-08-05 - v1.3.0 released! [[download](https://github.com/kubernetes/minikube/releases/tag/v1.3.0)] [[release notes](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md#version-130---2019-08-05)]

## Features

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
* [NVIDIA GPU support](https://minikube.sigs.k8s.io/docs/tutorials/nvidia_gpu/) - for machine learning
* [Filesystem mounts](https://minikube.sigs.k8s.io/docs/tasks/mount/)
* Automatic failure analysis

## Documentation

See https://minikube.sigs.k8s.io/docs/

## Community

![Help Wanted!](/images/help_wanted.jpg)

minikube is a Kubernetes [#sig-cluster-lifecycle](https://github.com/kubernetes/community/tree/master/sig-cluster-lifecycle)  project.

* [**#minikube on Kubernetes Slack**](https://kubernetes.slack.com) - Live chat with minikube developers!
* [minikube-users mailing list](https://groups.google.com/forum/#!forum/minikube-users)
* [minikube-dev mailing list](https://groups.google.com/forum/#!forum/minikube-dev)
* [Bi-weekly office hours, Mondays @ 10am PST](https://tinyurl.com/minikube-oh)

* [Contributing](https://minikube.sigs.k8s.io/docs/contributing/)
* [Development Roadmap](https://minikube.sigs.k8s.io/docs/contributing/roadmap/)

## Installation

See [getting started](https://minikube.sigs.k8s.io/docs/getting-started/)

## Examples

See [examples](https://minikube.sigs.k8s.io/docs/examples/)
