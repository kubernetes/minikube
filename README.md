# Minikube

[![BuildStatus Widget]][BuildStatus Result]
[![GoReport Widget]][GoReport Status]

[BuildStatus Result]: https://travis-ci.org/kubernetes/minikube
[BuildStatus Widget]: https://travis-ci.org/kubernetes/minikube.svg?branch=master

[GoReport Status]: https://goreportcard.com/report/github.com/kubernetes/minikube
[GoReport Widget]: https://goreportcard.com/badge/github.com/kubernetes/minikube

<img src="https://github.com/kubernetes/minikube/raw/master/logo/logo.png" width="100">

## What is minikube?

minikube makes running a local Kubernetes cluster fast and easy, supporting macOS, Linux, and Windows. Our design goals are fast local development, single command setup and teardown, and support for all Kubernetes features that fit.

## Kubernetes features

minikube runs the most recent stable release of Kubernetes, using the official release binaries. We explicitly support the following features:

* NodePorts - access via `minikube service`
* Ingress
* LoadBalancer - access via `minikube tunnel`
* DNS
* Persistent Volumes 
* ConfigMaps
* RBAC
* Secrets
* Dashboard - access via `minikube dashboard`
* Container runtimes - Docker, CRI-O, containerd 

## News

* 2019-02-15 - minikube v0.34.0 was released! See the [releases](https://github.com/kubernetes/minikube/releases) page for more.

## Community & Documentation

* [**#minikube on Kubernetes Slack**](https://kubernetes.slack.com) - Live chat with minikube developers!
* [**minikube-users mailing list](https://groups.google.com/forum/#!forum/minikube-users)
* [**minikube-dev mailing list](https://groups.google.com/forum/#!forum/minikube-dev)

* [**Advanced Topics and Tutorials**](https://github.com/kubernetes/minikube/blob/master/docs/README.md)
* [**Contributing**](https://github.com/kubernetes/minikube/blob/master/CONTRIBUTING.md)
* [**Development Guide**](https://github.com/kubernetes/minikube/blob/master/docs/contributors/README.md)


## Requirements

* 4GB of memory (VM reserves 2GB by default), 32GB of disk space
* An internet connection - preferably one that does not require a VPN or SSL proxy to access the internet
* macOS 10.12 (Sierra) or higher
  * Requires a hypervisor, such as:
     * hyperkit (recommended)
     * VirtualBox
* Linux
  * VT-x/AMD-v virtualization must be enabled in BIOS
  * libvirt for the KVM driver, or VirtualBox
* Windows 10
  * VT-x/AMD-v virtualization must be enabled in BIOS
  * HyperV (Pro) or a 3rd party hypervisor, such as VirtualBox.

## Installation

* *macOS* with brew: `brew cask install minikube` (if [Homebrew](https://brew.sh/ is installed)
* *macOS*: `curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64 \
  && sudo install minikube-darwin-amd64 /usr/local/bin/minikube`

* *Windows 10 with Choco* `choco install minikube` (if [Chocolatey](https://chocolatey.org/ is installed)
* *Windows 10 without Choco* - Download [minikube-windows-amd64.exe](https://storage.googleapis.com/minikube/releases/latest/minikube-windows-amd64.exe) file, rename it to `minikube.exe`, and add it to your path.

* *Generic Linux* `curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
  && sudo install minikube-linux-amd64 /usr/local/bin/minikube`
* [Arch Linux AUR](https://aur.archlinux.org/packages/minikube/)
* [Fedora/CentOS/Red Hat COPR](https://copr.fedorainfracloud.org/coprs/antonpatsev/minikube-rpm/)
* [Void Linux](https://github.com/void-linux/void-packages/tree/master/srcpkgs/minikube/template)
* [openSUSE/SUSE Linux Enterprise](https://build.opensuse.org/package/show/Virtualization:containers/minikube)

For full installation instructions, please see https://kubernetes.io/docs/tasks/tools/install-minikube/

### Supported Hypervisors

`minikube start` defaults to virtualbox, but supports other drivers using the `--vm-driver` argument:

* [KVM2](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#kvm2-driver) - Recommended Linux driver
* [hyperkit](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) - Recommended macOS driver
* virtualbox - Recommended Windows driver
* [hyperv](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperV-driver)
* vmwarefusion
* none (**Linux-only**) - run without a hypervisor, at the expense of system security or reliability.

## Quick Start

To start a cluster, run `minikube start`

You can then interact with it using `kubectl`, just like any other kubernetes cluster:

```
$ kubectl run hello-minikube --image=k8s.gcr.io/echoserver:1.4 --port=8080
deployment "hello-minikube" created

$ kubectl expose deployment hello-minikube --type=NodePort
service "hello-minikube" exposed
```

You can get the URL for the NodePort deployment by using:

`minikube service hello-minikube --url`

Start a second local cluster:

`minikube start -p cluster2`

Stop your local cluster:

`minikube stop`

Delete your local cluster:

`minikube delete`
