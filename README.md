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

<img src="https://github.com/kubernetes/minikube/raw/master/images/start.jpg" width="100">

Our [goal](https://github.com/kubernetes/minikube/blob/master/docs/contributors/principles.md) is to enable fast local development and to support all Kubernetes features that fit. We hope you enjoy it!

## News

* 2019-02-15 - v0.34.0 released! [[download](https://github.com/kubernetes/minikube/releases/tag/v0.34.0)] [[release notes](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md#version-0340---2019-02-15)]
* 2019-01-18 - v0.33.1 released to address [CVE-2019-5736](https://www.openwall.com/lists/oss-security/2019/02/11/2) [[download](https://github.com/kubernetes/minikube/releases/tag/v0.33.1)] [[release notes](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md#version-0331---2019-01-18)]
* 2019-01-17 - v0.33.0 released! [[download](https://github.com/kubernetes/minikube/releases/tag/v0.33.0)] [[release notes](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md#version-0330---2019-01-17)]

## Features

minikube runs the official stable release of Kubernetes, with support for standard Kubernetes features like:

* NodePorts - `minikube service`
* Ingress
* [LoadBalancer](https://github.com/kubernetes/minikube/blob/master/docs/tunnel.md) - `minikube tunnel` 
* [Persistent Volumes](https://github.com/kubernetes/minikube/blob/master/docs/persistent_volumes.md)
* ConfigMaps
* RBAC
* Secrets
* Dashboard - `minikube dashboard`
* [Multiple container runtimes](https://github.com/kubernetes/minikube/blob/master/docs/alternative_runtimes.md) - `start --container-runtime`

minikube also supports features for developer convenience:

* [Addons](https://github.com/kubernetes/minikube/blob/master/docs/addons.md) - a marketplace for developers to share configurations for running services on minikube
* [GPU support](https://github.com/kubernetes/minikube/blob/master/docs/gpu.md) - for machine learning
* [Filesystem mounts](https://github.com/kubernetes/minikube/blob/master/docs/host_folder_mount.md)
* Automatic failure analysis

## Community & Documentation

* [**#minikube on Kubernetes Slack**](https://kubernetes.slack.com) - Live chat with minikube developers!
* [minikube-users mailing list](https://groups.google.com/forum/#!forum/minikube-users)
* [minikube-dev mailing list](https://groups.google.com/forum/#!forum/minikube-dev)

* [**Advanced Topics and Tutorials**](https://github.com/kubernetes/minikube/blob/master/docs/README.md)
* [Contributing](https://github.com/kubernetes/minikube/blob/master/CONTRIBUTING.md)
* [Development Guide](https://github.com/kubernetes/minikube/blob/master/docs/contributors/README.md)
* [Development Roadmap](https://github.com/kubernetes/minikube/blob/master/docs/contributors/roadmap.md)

## Requirements

* 4GB of memory (VM reserves 2GB by default), 32GB of disk space
* An internet connection - preferably one that does not require a VPN or SSL proxy to access the internet
* macOS 10.12 (Sierra) or higher
  * Requires a hypervisor, such as:
     * hyperkit (recommended)
     * VirtualBox
* Linux
  * libvirt for the KVM driver, or VirtualBox
  * VT-x/AMD-v virtualization must be enabled in BIOS
* Windows 10
  * HyperV (Windows 10 Pro) or a 3rd party hypervisor, such as VirtualBox.
  * VT-x/AMD-v virtualization must be enabled in BIOS

## Installation

* *macOS* with [brew](https://brew.sh/): `brew cask install minikube`
* *macOS*: `curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64 && sudo install minikube-darwin-amd64 /usr/local/bin/minikube`

* *Windows with [Chocolatey](https://chocolatey.org/)* `choco install minikube`
* *Windows without Choco* - Download and run the [installer](https://storage.googleapis.com/minikube/releases/latest/minikube-installer.exe)

* *Linux* `curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && sudo install minikube-linux-amd64 /usr/local/bin/minikube`

For full installation instructions, please see https://kubernetes.io/docs/tasks/tools/install-minikube/

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

Start a cluster:

`minikube start`

Interact with it using `kubectl`, just like any other kubernetes cluster:


```
$ kubectl run hello-minikube --image=k8s.gcr.io/echoserver:1.4 --port=8080
deployment "hello-minikube" created

$ kubectl expose deployment hello-minikube --type=NodePort
service "hello-minikube" exposed
```

Open the endpoint in your browser:

`minikube service hello-minikube`

Start a second local cluster:

`minikube start -p cluster2`

Stop your local cluster:

`minikube stop`

Delete your local cluster:

`minikube delete`
