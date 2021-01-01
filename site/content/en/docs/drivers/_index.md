---
title: "Drivers"
linkTitle: "Drivers"
weight: 8
no_list: true
description: >
  Configuring various minikube drivers
aliases:
  - /docs/reference/drivers
---
Minikube can be deployed as a VM, a container, or bare-metal where a server can be used as a single tenant.

Docker and Podman are container engines with almost similar operating commands. Podman interacts directly with the container, whereas Docker uses daemon for this.
The [Docker Machine](https://github.com/docker/machine) library is used to provide a consistent way to interact with different environments. It's a VM that runs docker and Linux simultaneously. 

Here is what's supported:

## Linux

* [Docker]({{<ref "docker.md">}}) - container-based (preferred)
* [KVM2]({{<ref "kvm2.md">}}) - VM-based (preferred)
* [VirtualBox]({{<ref "virtualbox.md">}}) - VM
* [None]({{<ref "none.md">}}) -  bare-metal
* [Podman]({{<ref "podman.md">}}) - container (experimental)

## macOS

* [Docker]({{<ref "docker.md">}}) - VM + Container (preferred)
* [Hyperkit]({{<ref "hyperkit.md">}}) - VM
* [VirtualBox]({{<ref "virtualbox.md">}}) - VM
* [Parallels]({{<ref "parallels.md">}}) - VM
* [VMware]({{<ref "vmware.md">}}) - VM

## Windows

* [Hyper-V]({{<ref "hyperv.md">}}) - VM (preferred)
* [Docker]({{<ref "docker.md">}}) - VM + Container (preferred)
* [VirtualBox]({{<ref "virtualbox.md">}}) - VM
