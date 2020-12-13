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
There are various options through which minikube can be deployed. These include using a VM, a container, or bare metal. In the case of bare metal environments, a server can be used as a single tenant. 
In the windows operating system, docker is sometimes referred to as the Docker Engine. Docker desktop can be used to run containers of both Linux and Windows os. 
Docker and Podman are container engines and have almost similar commands for their operations. Podman directly interacts with the container content, however, Docker uses daemon for this.
To do so, we use the [Docker Machine](https://github.com/docker/machine) library to provide a consistent way to interact with different environments. In simple words, it's a VM that runs docker and Linux simultaneously.  

##Hypervisors
A hypervisor supervises the creating and running of VMs hence allowing multiple guest VM for a single guest. Through this process, the host is provided access to the resources of the VMs.Different kinds of operating systems need a different kind of hypervisors.

#Hyper-V 
Hyper-V provides another alternative of running Linux containers on Windows using docker. In this approach, every Window containers share a single kernel but every Linux container has a separate Linux kernel.

##KVM2 
KVM2 or a VirtualBox can be used as a hypervisor for Linux. An alternative container runtime to the docker driver is the podman driver.

MacBook doesn't directly run based on any containers and a different set of the hypervisor are required. HyperKit, VirtualBox, VMware Fusion are some of the possible hypervisors that can be used. Parallels driver is another driver option for users working with the Parallels Desktop. It provides an easy blending of windows and OS.
In MacBooks, the docker CLI is preconfigured to interact with the docker daemon that’s running directly on our MacBook.

In the case of a windows system, a docker based native Kubernetes integration can be used. This would be ideal for Windows 10. However, a hypervisor can provide the best performance. Hyper-V and VirtualBox are two available hypervisor options for windows. The only problem with the Hyper-V hypervisor is that it requires a restart while switching between the container and VM.

Here's a comprehensive list of all options for getting started on different os:

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
