# vm-driver=none

## Overview

This document is written for system integrators who are familiar with minikube, and wish to run it within a customized VM environment.

The `none` driver allows advanced minikube users to skip VM creation, allowing minikube to be run on a user-supplied VM.

## What operating systems are supported?

The `none` driver supports releases of Debian, Ubuntu, and Fedora that are less than 2 years old. In practice, any systemd-based modern distribution is likely to work, and we will accept pull requests which improve compatibility with other systems.

## Can the none driver be used outside of a VM?

Yes, *but please avoid doing so if at all possible.*

minikube was designed to run Kubernetes within a dedicated VM, and assumes that it has complete control over the machine it is executing on.  With the `none` driver, minikube and Kubernetes run in an environment with very limited isolation, which could result in:

* Decreased security
* Decreased reliability
* Data loss

We'll cover these in detail below:

### Decreased security

* minikube starts services that may be available on the Internet. Please ensure that you have a firewall to protect your host from unexpected access. For instance:
  * apiserver listens on TCP *:8443
  * kubelet listens on TCP *:10250 and *:10255
  * kube-scheduler listens on TCP *:10259
  * kube-controller listens on TCP *:10257
* Containers may have full access to your filesystem.
* Containers may be able to execute arbitrary code on your host, by using container escape vulnerabilities such as [CVE-2019-5736](https://access.redhat.com/security/vulnerabilities/runcescape). Please keep your release of minikube up to date.

### Decreased reliability

* minikube with the none driver may be tricky to configure correctly at first, because there are many more chances for interference with other locally run services, such as dnsmasq.

* When run in `none` mode, minikube has no built-in resource limit mechanism, which means you could deploy pods which would consume all of the hosts resources.

* minikube and the Kubernetes services it starts may interfere with other running software on the system. For instance, minikube will start and stop container runtimes via systemd, such as docker, rkt, containerd, cri-o.

### Data loss

With the `none` driver, minikube will overwrite the following system paths:

* /usr/local/bin/kubeadm - Updated to match the exact version of Kubernetes selected
* /usr/local/bin/kubectl - Updated to match the exact version of Kubernetes selected
* /etc/kubernetes - configuration files

These paths will be erased when running `minikube delete`:

* /data/minikube
* /etc/kubernetes/manifests
* /var/lib/minikube

As Kubernetes has full access to both your filesystem as well as your docker images, it is possible that other unexpected data loss issues may arise.

## Known Issues

* You cannot run more than one `--vm-driver=none` instance on a single host
* Many `minikube` commands are not supported, such as: `dashboard`, `mount`, `ssh`
* minikube with the `none` driver has a confusing permissions model, as some commands need to be run as root ("start"), and others by a regular user ("dashboard")
* CoreDNS detects resolver loop, goes into CrashloopBackoff - [#3511](https://github.com/kubernetes/minikube/issues/3511)
* [Full list of issues](https://github.com/kubernetes/minikube/labels/co%2Fnone-driver)

