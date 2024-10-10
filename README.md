# minikube

[![Actions Status](https://github.com/kubernetes/minikube/workflows/build/badge.svg)](https://github.com/kubernetes/minikube/actions)
[![GoReport Widget]][GoReport Status]
[![GitHub All Releases](https://img.shields.io/github/downloads/kubernetes/minikube/total.svg)](https://github.com/kubernetes/minikube/releases/latest)
[![Latest Release](https://img.shields.io/github/v/release/kubernetes/minikube?include_prereleases)](https://github.com/kubernetes/minikube/releases/latest)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/5015/badge)](https://www.bestpractices.dev/en/projects/5015)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/kubernetes/minikube/badge)](https://api.securityscorecards.dev/projects/github.com/kubernetes/minikube)
 

[GoReport Status]: https://goreportcard.com/report/github.com/kubernetes/minikube
[GoReport Widget]: https://goreportcard.com/badge/github.com/kubernetes/minikube

<img src="https://github.com/kubernetes/minikube/raw/master/images/logo/logo.png" width="100" alt="minikube logo">

minikube implements a local Kubernetes cluster on macOS, Linux, and Windows. minikube's [primary goals](https://minikube.sigs.k8s.io/docs/concepts/principles/) are to be the best tool for local Kubernetes application development and to support all Kubernetes features that fit. 

<img src="https://raw.githubusercontent.com/kubernetes/minikube/master/site/static/images/screenshot.png" width="575" height="322" alt="screenshot">

## Features

minikube runs the latest stable release of Kubernetes, with support for standard Kubernetes features like:

* [LoadBalancer](https://minikube.sigs.k8s.io/docs/handbook/accessing/#loadbalancer-access) - using `minikube tunnel`
* Multi-cluster - using `minikube start -p <name>`
* [NodePorts](https://minikube.sigs.k8s.io/docs/handbook/accessing/#nodeport-access) - using `minikube service`
* [Persistent Volumes](https://minikube.sigs.k8s.io/docs/handbook/persistent_volumes/)
* [Ingress](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/)
* [Dashboard](https://minikube.sigs.k8s.io/docs/handbook/dashboard/) - `minikube dashboard`
* [Container runtimes](https://minikube.sigs.k8s.io/docs/handbook/config/#runtime-configuration) - `minikube start --container-runtime`
* [Configure apiserver and kubelet options](https://minikube.sigs.k8s.io/docs/handbook/config/#modifying-kubernetes-defaults) via command-line flags
* Supports common [CI environments](https://github.com/minikube-ci/examples)

As well as developer-friendly features:

* [Addons](https://minikube.sigs.k8s.io/docs/handbook/deploying/#addons) - a marketplace for developers to share configurations for running services on minikube
* [NVIDIA GPU support](https://minikube.sigs.k8s.io/docs/tutorials/nvidia/) - for machine learning
* [AMD GPU support](https://minikube.sigs.k8s.io/docs/tutorials/amd/) - for machine learning
* [Filesystem mounts](https://minikube.sigs.k8s.io/docs/handbook/mount/)

**For more information, see the official [minikube website](https://minikube.sigs.k8s.io)**

## Installation

See the [Getting Started Guide](https://minikube.sigs.k8s.io/docs/start/)

:mega: **Please fill out our [fast 5-question survey](https://forms.gle/Gg3hG5ZySw8c1C24A)** so that we can learn how & why you use minikube, and what improvements we should make. Thank you! :dancers:

## Documentation

See https://minikube.sigs.k8s.io/docs/

## More Examples

See minikube in action [here](https://minikube.sigs.k8s.io/docs/handbook/controls/)

## Governance

Kubernetes project is governed by a framework of principles, values, policies and processes to help our community and constituents towards our shared goals.

The [Kubernetes Community](https://github.com/kubernetes/community/blob/master/governance.md) is the launching point for learning about how we organize ourselves.

The [Kubernetes Steering community repo](https://github.com/kubernetes/steering) is used by the Kubernetes Steering Committee, which oversees governance of the Kubernetes project.

## Community

minikube is a Kubernetes [#sig-cluster-lifecycle](https://github.com/kubernetes/community/tree/master/sig-cluster-lifecycle)  project.

* [**#minikube on Kubernetes Slack**](https://kubernetes.slack.com/messages/minikube) - Live chat with minikube developers!
* [minikube-users mailing list](https://groups.google.com/g/minikube-users)
* [minikube-dev mailing list](https://groups.google.com/g/minikube-dev)

* [Contributing](https://minikube.sigs.k8s.io/docs/contrib/)
* [Development Roadmap](https://minikube.sigs.k8s.io/docs/contrib/roadmap/)

Join our community meetings:
* [Bi-weekly office hours, Mondays @ 11am PST](https://tinyurl.com/minikube-oh)
* [Triage Party](https://minikube.sigs.k8s.io/docs/contrib/triage/)
