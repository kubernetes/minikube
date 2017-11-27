# Minikube Design Document

This document is intended to serve as a living, up-to-date design document for the `minikube` tool.

It should be useful to anyone interested in contributing to the `minikube` codebase.

## Overview

Minikube is a command-line tool that allows users to run a single-node Kubernetes cluster, on their workstation.

This goal translates to a few hard requirements for the tool that influence it's design:

* Minikube must run on the platforms most commonly used on developer workstations (Windows, Linux, OSX)
* Minikube must provide a platform Kubernetes can run in (Linux/amd64)

Minikube is logically structured into two main components, following these requirements: the host components and the guest components.

### Host Components

The first component is responsible for providing a linux/amd64 environment (in a virtual machine) on the workstation.
This component mainly consists of:

* The CLI for managing the lifecycle of the VM and cluster
* The `drivers` responsible for interfacing with hypervisors
* The `minikube` helper commands for interfacing with the running cluster (`minikube service`, `minikube dashboard`, etc.)

### Guest Components

The second component is responsible for bootstrapping a Kubernetes cluster in environment.
This component consists mainly of:

* The custom Linux distribution, packaged as an ISO, containing the dependencies required to run a Kubernetes cluster
* The code to bootstrap a cluster, using either `localkube` or `kubeadm`

## Life of a `minikube` cluster

This section details what happens when a user runs the `minikube start` command.

* Minikube parses the disk configuration and flags to determine which driver, k8s version and bootstrapper to use
* Minikube creates a docker-machine API client for the specified driver
* Minikube creates a `MachineConfig` struct with the specified configuration for disk size, memory, CPU, etc.
* Minikube instantiates the driver/cient and uses it to create the virtual machine
* Minikube creates a 'KubernetesConfig' struct with the specified configuration and `--extra-config` options.
* Minikube creates cluster certificates and uses these to setup cluster authentication
* Minikube instantiates the specified `Bootstrapper` and uses it to create the cluster

## Structs and Interfaces

Minikube is mostly written in `Go`, so it is useful to think about the major functionality of `minikube` in terms of Go's structs and interfaces.

### Interfaces

#### Drivers

Drivers the interface that provide hypervisor integration.
Drivers exist for hypervisors such as Virtualbox, KVM, Hyperkit and Hyper-v.

It is unlikely `minikube` will ever use only a single driver, so the `Driver` interface will be a part of `minikube` forever.

##### Overview

These drivers are repsonsible for exposing functions to manage the lifecycle of a VM using their respective hypervisor.
These methods include creating, starting, stopping and communicating with a virtual machine.

The full interface is defined TODO: HERE.

Minikube is designed to reuse some of the drivers implemented for [docker-machine](TODO).

##### In Process

To support `docker-machine` drivers and drivers that are distributed and installed separately, `minikube` supports two types of drivers: in-process and out-of-process.

In-process drivers are built and distributed in the same binary as the minikube CLI.
They are instantiated as Go structs and the methods are called as Go functions.

##### Out of Process

Out-of-process drivers are built and distributed as stand-alone binaries.
They are almost always implemented in Go, but do not necessarily need to be.
These drivers are instantiated as a separate process, and the methods are called via IPC.

#### Bootstrappers

##### Overview

The next major interface in `minikube` is the `Bootstrapper`.
Drivers control creation of the virtual machine, bootstrappers control creation of the cluster.

The `Bootstrapper` interface is hopefully temporary - once `localkube` is fully deprecated, the interface will be removed.

##### Localkube

Localkube was used before the original implementation of `minikube`, and predates the `bootstrapper` interface.

The localkube bootstrapper uses a special binary that contains all kubernetes cluster components and runs them as goroutines in a single process.

This bootstrapper is considered deprecated, and will be removed after the Kubernetes 1.9 release.

##### Kubeadm

Kubeadm will be the only bootstrapper supported after the Kubernetes 1.9 release.
Kubeadm is the standard Kubernetes cluster installation and configuration tool, and minikube uses this by default.

### Structs

#### MachineConfig

This object contains the virtual machine configuration that is standard to all drivers.
Settings like number of CPUs, amount of memory, the Docker version and settings all live on this struct.

#### KubernetesConfig

This object contains the cluster configuration.
Settings like the Kubernetes version, component flags, certificate and network information all live on this struct.

### Minikube.iso

TODO

## Other Interesting Pieces

### CLI Configuration

### Profiles

### Configurator

### Mount

### Networking

### CI

### Image Caching