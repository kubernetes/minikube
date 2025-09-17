---
title: "none"
weight: 3
description: >
  Linux none (bare-metal) driver
aliases:
    - /docs/reference/drivers/none
---

## Overview

{{% pageinfo %}}
Most users of this driver should consider the newer [Docker driver]({{< ref "docker.md" >}}), as it is
significantly easier to configure and does not require root access. The 'none' driver is recommended for advanced users only.
{{% /pageinfo %}}

This document is written for system integrators who wish to run minikube within a customized VM environment. The `none` driver allows advanced minikube users to skip VM creation, allowing minikube to be run on a user-supplied VM.

{{% readfile file="/docs/drivers/includes/none_usage.inc" %}}

## Issues

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

* minikube and the Kubernetes services it starts may interfere with other running software on the system. For instance, minikube will start and stop container runtimes via systemd, such as docker, containerd, cri-o.

### Persistent storage

* minikube expects that some mount points used for volumes are bind-mounted or symlinked to a persistent location:

   * `/data`
   * `/tmp/hostpath_pv`
   * `/tmp/hostpath-provisioner`

If you don't have a dedicated disk to use for these, you can use the `/var` partition which is _usually_ persistent.

### Data loss

With the `none` driver, minikube will overwrite the following system paths:

* /etc/kubernetes - configuration files

These paths will be erased when running `minikube delete`:

* /data/minikube
* /etc/kubernetes/manifests
* /var/lib/minikube

As Kubernetes has full access to both your filesystem as well as your docker images, it is possible that other unexpected data loss issues may arise.

### Other

* `-p` (profiles) are unsupported: It is not possible to run more than one `--driver=none` instance
* Many `minikube` commands are not supported, such as: `dashboard`, `mount`, `ssh`
* minikube with the `none` driver has a confusing permissions model, as some commands need to be run as root ("start"), and others by a regular user ("dashboard")
* CoreDNS detects resolver loop, goes into CrashLoopBackOff - [#3511](https://github.com/kubernetes/minikube/issues/3511)
* Some versions of Linux have a version of docker that is newer than what Kubernetes expects. To overwrite this, run minikube with the following parameters: `minikube start --driver=none --kubernetes-version v1.34.0 --extra-config kubeadm.ignore-preflight-errors=SystemVerification`

* [Full list of open 'none' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fnone-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=4` to debug crashes
