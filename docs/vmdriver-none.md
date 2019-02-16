# vm-driver=none

## Overview

This document is written for system integrators who are familiar with minikube, and wish to run it within a customized VM environment.

The `none` driver allows advanced minikube users to skip VM creation, allowing minikube to be run on a user-supplied VM.

## What operating systems are supported?

The `none` driver supports releases of Debian, Ubuntu, and Fedora that are less than 2 years old

In practice, any systemd-based modern distribution is likely to work, and we will happily accept pull requests which improve compatibility with other systems.

## Can vm-driver=none be used outside of a VM?

Not if you can avoid it.

minikube was designed to run Kubernetes within a dedicated VM, and assumes that it has complete control over the machine it is executing on. With the `none` driver, minikube will overwrite the following system paths:

* /usr/local/bin/kubeadm
* /usr/local/bin/kubectl
* /etc/kubernetes

## Security

With the `none` driver, minikube has only limited container isolation abilities, which can result in decreased host security and reliability. It is highly recommended that your host is isolated from the rest of the network using a firewall.

Additionally, minikube with the `none` driver has a very confusing permissions model, as some commands need to be run as root ("start"), and others by a regular user ("dashboard").

## Known Issues

* You cannot run more than one `--vm-driver=none` instance on a single host
* Many `minikube` commands are not supported, such as: `dashboard`, `mount`, `ssh`
