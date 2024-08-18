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
minikube can be deployed as a VM, a container, or bare-metal.

To do so, we use the [Docker Machine](https://github.com/docker/machine) library to provide a consistent way to interact with different environments. Here is what's supported:

## Linux

* [Docker]({{<ref "docker.md">}}) - container-based (preferred)
* [KVM2]({{<ref "kvm2.md">}}) - VM-based (preferred)
* [VirtualBox]({{<ref "virtualbox.md">}}) - VM
* [QEMU]({{<ref "qemu.md">}}) - VM
* [None]({{<ref "none.md">}}) -  bare-metal
* [Podman]({{<ref "podman.md">}}) - container-based (experimental)
* [SSH]({{<ref "ssh.md">}}) - remote ssh


## macOS

* [Docker]({{<ref "docker.md">}}) - VM + Container (preferred)
* [Hyperkit]({{<ref "hyperkit.md">}}) - VM
* [VirtualBox]({{<ref "virtualbox.md">}}) - VM
* [Parallels]({{<ref "parallels.md">}}) - VM
* [VMware Fusion]({{<ref "vmware.md">}}) - VM
* [QEMU]({{<ref "qemu.md">}}) - VM
* [Podman]({{<ref "podman.md">}}) - VM + Container (experimental)
* [VFkit]({{<ref "vfkit.md">}}) - VM (experimental)
* [SSH]({{<ref "ssh.md">}}) - remote ssh

## Windows

* [Hyper-V]({{<ref "hyperv.md">}}) - VM (preferred)
* [Docker]({{<ref "docker.md">}}) - VM + Container (preferred)
* [VirtualBox]({{<ref "virtualbox.md">}}) - VM
* [VMware Workstation]({{<ref "vmware.md">}}) - VM
* [QEMU]({{<ref "qemu.md">}}) - VM (experimental)
* [Podman]({{<ref "podman.md">}}) - VM + Container (experimental)
* [SSH]({{<ref "ssh.md">}}) - remote ssh
