# minikube

[![BuildStatus Widget]][BuildStatus Result]
[![GoReport Widget]][GoReport Status]

[BuildStatus Result]: https://travis-ci.org/kubernetes/minikube
[BuildStatus Widget]: https://travis-ci.org/kubernetes/minikube.svg?branch=master

[GoReport Status]: https://goreportcard.com/report/github.com/kubernetes/minikube
[GoReport Widget]: https://goreportcard.com/badge/github.com/kubernetes/minikube

<img src="https://github.com/kubernetes/minikube/raw/master/images/logo/logo.png" width="100" alt="minikube logo">

minikube implements a local Kubernetes cluster on macOS, Linux, and Windows. minikube's [primary goals](https://minikube.sigs.k8s.io/docs/concepts/principles/) are to be the best tool for local Kubernetes application development and to support all Kubernetes features that fit. 

<img src="https://github.com/kubernetes/minikube/raw/master/site/content/en/start.png" width="738" alt="screenshot">

## Features

minikube runs the latest stable release of Kubernetes, with support for standard Kubernetes features like:

* [LoadBalancer](https://minikube.sigs.k8s.io/docs/tasks/loadbalancer/) - using `minikube tunnel`
* Multi-cluster - using `minikube start -p <name>`
* NodePorts - using `minikube service`
* [Persistent Volumes](https://minikube.sigs.k8s.io/docs/reference/persistent_volumes/)
* [Ingress](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/)
* [Dashboard](https://minikube.sigs.k8s.io/docs/tasks/dashboard/) - `minikube dashboard`
* [Container runtimes](https://minikube.sigs.k8s.io/docs/reference/runtimes/) - `start --container-runtime`
* [Configure apiserver and kubelet options](https://minikube.sigs.k8s.io/docs/reference/configuration/kubernetes/) via command-line flags

As well as developer-friendly features:

* [Addons](https://minikube.sigs.k8s.io/docs/tasks/addons/) - a marketplace for developers to share configurations for running services on minikube
* [NVIDIA GPU support](https://minikube.sigs.k8s.io/docs/tutorials/nvidia_gpu/) - for machine learning
* [Filesystem mounts](https://minikube.sigs.k8s.io/docs/tasks/mount/)

**For more information, see the official [minikube website](https://minikube.sigs.k8s.io)**

## Installation

See the [Getting Started Guide](https://minikube.sigs.k8s.io/docs/start/)

:mega: **Please fill out our [fast 5-question survey](https://forms.gle/Gg3hG5ZySw8c1C24A)** so that we can learn how & why you use minikube, and what improvements we should make. Thank you! :dancers:

## Documentation

See https://minikube.sigs.k8s.io/docs/

## More Examples

See our [examples page](https://minikube.sigs.k8s.io/docs/examples/)

## Community

minikube is a Kubernetes [#sig-cluster-lifecycle](https://github.com/kubernetes/community/tree/master/sig-cluster-lifecycle)  project.

* [**#minikube on Kubernetes Slack**](https://kubernetes.slack.com) - Live chat with minikube developers!
* [minikube-users mailing list](https://groups.google.com/forum/#!forum/minikube-users)
* [minikube-dev mailing list](https://groups.google.com/forum/#!forum/minikube-dev)
* [Bi-weekly office hours, Mondays @ 11am PST](https://tinyurl.com/minikube-oh)

* [Contributing](https://minikube.sigs.k8s.io/docs/contributing/)
* [Development Roadmap](https://minikube.sigs.k8s.io/docs/contributing/roadmap/)

