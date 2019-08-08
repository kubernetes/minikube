---
title: "VirtualBox"
linkTitle: "VirtualBox"
weight: 1
date: 2018-08-08
description: >
  VirtualBox driver
---

## Overview

VirtualBox is the oldest and most stable VM driver for minikube.


## Requirements

- [https://www.virtualbox.org/wiki/Downloads](VirtualBox) 5.2 or higher

## Usage

minikube currently uses VirtualBox by default, but it can also be explicitly set:

```shell
minikube start --vm-driver=virtualbox
```
To make virtualbox the default driver:

```shell
minikube config set vm-driver virtualbox
```

## Special features

minikube start supports some VirtualBox specific flags:

* **\--host-only-cidr**: The CIDR to be used for the minikube VM (default "192.168.99.1/24")
* **\--no-vtx-check**: Disable checking for the availability of hardware virtualization 

## Issues

* [Full list of open 'virtualbox' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fvirtualbox)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes