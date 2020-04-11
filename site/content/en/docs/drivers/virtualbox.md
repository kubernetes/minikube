---
title: "virtualbox"
weight: 5
aliases:
  - /docs/reference/drivers/virtualbox
---

## Overview

VirtualBox is minikube's original driver. It may not provide the fastest start-up time, but it is the most stable driver available for users of Microsoft Windows Home.

{{% readfile file="/docs/drivers/includes/virtualbox_usage.inc" %}}

## Special features

minikube start supports some VirtualBox specific flags:

* **`--host-only-cidr`**: The CIDR to be used for the minikube VM (default "192.168.99.1/24")
* **`--no-vtx-check`**: Disable checking for the availability of hardware virtualization

## Issues

* [Full list of open 'virtualbox' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fvirtualbox)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
