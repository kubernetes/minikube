---
title: "hyperkit"
linkTitle: "hyperkit"
weight: 1
date: 2018-08-08
description: >
  HyperKit driver
---

## Overview

[HyperKit](https://github.com/moby/hyperkit) is an open-source hypervisor for macOS hypervisor, optimized for lightweight virtual machines and container deployment.

{{% readfile file="/docs/Reference/Drivers/includes/hyperkit_usage.md" %}}

## Special features

minikube start supports some VirtualBox specific flags:

* **\--host-only-cidr**: The CIDR to be used for the minikube VM (default "192.168.99.1/24")
* **\--no-vtx-check**: Disable checking for the availability of hardware virtualization 

## Issues

### Local DNS server conflict

If you are using dnsmasq in your setup and cluster creation fails (stuck at kube-dns initialization) you might need to add listen-address=192.168.64.1 to dnsmasq.conf.

Note: If dnsmasq.conf contains listen-address=127.0.0.1 kubernetes discovers dns at 127.0.0.1:53 and tries to use it using bridge ip address, but dnsmasq replies only to requests from 127.0.0.1

### Other

* [Full list of open 'hyperkit' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fhyperkit)

## Troubleshooting

* Run `docker-machine-driver-hyperkit version` to verify that the version matches minikubes version. Example output:

```
version: v1.3.0
commit: 43969594266d77b555a207b0f3e9b3fa1dc92b1f
````

* Run `minikube start --alsologtostderr -v=7` to debug crashes
