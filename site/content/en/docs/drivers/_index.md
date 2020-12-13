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
Added a one liner description of bare metal environments.
There are various options through which minikube can be deployed. These include using a VM, a container, or bare metal. In the case of bare metal environments, a server can be used as a single tenant. 

Difference between Docker Desktop and Docker Engine. 
In the windows operating system, docker is sometimes referred to as Docker Engine. Docker desktop can be used to run containers of both Linux and Windows os. 

Difference between Podman and Docker
Docker and Podman are container engines and have almost similar commands for their operations. Podman directly interacts with the container content, however, Docker uses daemon for this.

Added a one liner explanation of Docker machine library
The [Docker Machine](https://github.com/docker/machine) library provides dynamic ways to interact with different environments. In simple words, it's a VM that runs docker and Linux simultaneously.   

Explanation about hypervisor and Hyper-V
A hypervisor supervises the creating and running of VMs hence allowing multiple guest VM for a single guest. Through this process, the host is provided access to the resources of the VMs.Different kinds of operating systems need a different kind of hypervisors. 

Hyper-V provides another alternative of running Linux containers on Windows using docker. In this approach, every Window containers share a single kernel but every Linux container has a separate Linux kernel. 

Intro to KVM2 and explanation of the use of podman 
KVM2 or a VirtualBox can be used as a hypervisor for linux. An alternative container runtime to docker driver is the podman driver.

Explanation of the different possible options for macOS
Macbook doesn't directly run based on any containers and a different set of hypervisor are required. HyperKit, VirtualBox, VMware Fusion are some of the possible hypervisors that can be used. 

Intro to parallels
Parallels driver is another driver option for users working with the Parallels Desktop. It provides an easy blending of windows and OS.  
In macbooks, the docker cli is preconfigured to interact with the docker daemon that’s running directly on our macbook.

A little detail about installtion in windows 
In the case of a windows system, a docker based native Kubernetes integration can be used. This would be ideal for Windows 10. However, a hypervisor can provide the best performance.
Details about Hyper-V and VirtualBox windows
Hyper-V and VirtualBox are two available hypervisor options for windows.
The only problem with the Hyper-V hypervisor is that it requires a restart while switching between the container and VM.

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

Docker can be used uptill versions v1.15.1. For new versions, it is recommended to use hypervisors discussed above. 
