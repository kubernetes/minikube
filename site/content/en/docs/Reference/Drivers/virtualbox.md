---
title: "virtualbox"
linkTitle: "virtualbox"
weight: 5
date: 2018-08-08
description: >
  VirtualBox driver
---

## Overview

VirtualBox is the oldest and most stable VM driver for minikube.

{{% readfile file="/docs/Reference/Drivers/includes/virtualbox_usage.inc" %}}

## Special features

minikube start supports some VirtualBox specific flags:

* **`--host-only-cidr`**: The CIDR to be used for the minikube VM (default "192.168.99.1/24")
* **`--no-vtx-check`**: Disable checking for the availability of hardware virtualization

## Issues

* [Full list of open 'virtualbox' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fvirtualbox)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
