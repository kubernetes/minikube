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

minikube's [primary goals](https://github.com/kubernetes/minikube/blob/master/docs/contributors/principles.md) are to be the best tool for local Kubernetes application development and to support all Kubernetes features that fit. We hope you enjoy it!

## News

* 2019-05-21 - v1.1.0 released! [[download](https://github.com/kubernetes/minikube/releases/tag/v1.1.0)] [[release notes](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md#version-110---2019-05-21)]
* 2019-04-29 - v1.0.1 released! [[download](https://github.com/kubernetes/minikube/releases/tag/v1.0.1)] [[release notes](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md#version-101---2019-04-29)]
* 2019-03-27 - v1.0.0 released! [[download](https://github.com/kubernetes/minikube/releases/tag/v1.0.0)] [[release notes](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md#version-1000---2019-03-27)]

## Features

minikube runs the official stable release of Kubernetes, with support for standard Kubernetes features like:

* [LoadBalancer](https://github.com/kubernetes/minikube/blob/master/docs/tunnel.md) - using `minikube tunnel`
* Multi-cluster - using `minikube start -p <name>`
* NodePorts - using `minikube service`
* [Persistent Volumes](https://github.com/kubernetes/minikube/blob/master/docs/persistent_volumes.md)
* Ingress
* RBAC
* Dashboard - `minikube dashboard`
* [Container runtimes](https://github.com/kubernetes/minikube/blob/master/docs/alternative_runtimes.md) - `start --container-runtime`
* [Configure apiserver and kubelet options](https://github.com/kubernetes/minikube/blob/master/docs/configuring_kubernetes.md) via command-line flags

As well as developer-friendly features:

* [Addons](https://github.com/kubernetes/minikube/blob/master/docs/addons.md) - a marketplace for developers to share configurations for running services on minikube
* [GPU support](https://github.com/kubernetes/minikube/blob/master/docs/gpu.md) - for machine learning
* [Filesystem mounts](https://github.com/kubernetes/minikube/blob/master/docs/host_folder_mount.md)
* Automatic failure analysis

## Documentation

* [**Installation**](https://kubernetes.io/docs/tasks/tools/install-minikube/)
* [Advanced Topics and Tutorials](https://github.com/kubernetes/minikube/blob/master/docs/README.md)
* [Contributors Guide](https://github.com/kubernetes/minikube/blob/master/docs/contributors/README.md)

## Community

![Help Wanted!](/images/help_wanted.jpg)

minikube is a Kubernetes [#sig-cluster-lifecycle](https://github.com/kubernetes/community/tree/master/sig-cluster-lifecycle)  project.

* [**#minikube on Kubernetes Slack**](https://kubernetes.slack.com) - Live chat with minikube developers!
* [minikube-users mailing list](https://groups.google.com/forum/#!forum/minikube-users)
* [minikube-dev mailing list](https://groups.google.com/forum/#!forum/minikube-dev)
* [Bi-weekly office hours, Mondays @ 10am PST](https://tinyurl.com/minikube-oh)

* [Contributing](https://github.com/kubernetes/minikube/blob/master/CONTRIBUTING.md)
* [Development Roadmap](https://github.com/kubernetes/minikube/blob/master/docs/contributors/roadmap.md)

## Installation

See the [installation guide](https://kubernetes.io/docs/tasks/tools/install-minikube/). For the impatient, here is the TL;DR:

* *macOS 10.12 (Sierra)*
  * Requires installing a hypervisor, such as [hyperkit](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) (recommended) or VirtualBox
  * using [brew](https://brew.sh/): `brew cask install minikube`
  * manually: `curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64 && sudo install minikube-darwin-amd64 /usr/local/bin/minikube`

* *Windows 10*
  * Requires a hypervisor, such as VirtualBox (recommended) or HyperV
  * VT-x/AMD-v virtualization must be enabled in BIOS
  * using [chocolatey](https://chocolatey.org/) `choco install minikube`
  * manually: Download and run the [installer](https://storage.googleapis.com/minikube/releases/latest/minikube-installer.exe)

* *Linux*
  * Requires either the [kvm2 driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver) (recommended), or VirtualBox
  * VT-x/AMD-v virtualization must be enabled in BIOS
  * manually:  `curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && sudo install minikube-linux-amd64 /usr/local/bin/minikube`

### Supported Hypervisors

`minikube start` defaults to virtualbox, but supports other drivers using the `--vm-driver` argument:

* [KVM2](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver) - Recommended Linux driver
* [hyperkit](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) - Recommended macOS driver
* virtualbox - Recommended Windows driver
* [none](https://github.com/kubernetes/minikube/blob/master/docs/vmdriver-none.md) - bare-metal execution on Linux, at the expense of system security and reliability

Other drivers which are not yet part of our continuous integration system are:

* [hyperv](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperV-driver)
* [vmware](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#vmware-unified-driver)

## Quick Start

Start a cluster by running:

`minikube start`

Access Kubernetes Dashboard within Minikube:

`minikube dashboard`

Once started, you can interact with your cluster using `kubectl`, just like any other Kubernetes cluster. For instance, starting a server:

`kubectl run hello-minikube --image=k8s.gcr.io/echoserver:1.4 --port=8080`

Exposing a service as a NodePort

`kubectl expose deployment hello-minikube --type=NodePort`

minikube makes it easy to open this exposed endpoint in your browser:

`minikube service hello-minikube`

Start a second local cluster:

`minikube start -p cluster2`

Stop your local cluster:

`minikube stop`

Delete your local cluster:

`minikube delete`
