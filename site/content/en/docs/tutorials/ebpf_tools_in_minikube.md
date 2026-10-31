---
title: "Running eBPF Tools in Minikube"
linkTitle: "Running eBPF Tools in Minikube"
weight: 1
date: 2019-08-15
description: >
  Running eBPF Tools in Minikube
---

## Overview

eBPF tools are performance tools used for observing the Linux kernel.
These tools can be used to monitor your Kubernetes application in minikube.
This tutorial will cover how to set up your minikube cluster so that you can run eBPF tools from a Docker container within minikube.


## Requirement
- Any supported minikube driver (Docker, Podman, or VM)
- x86 (currently the [bcc image](https://github.com/iovisor/bcc) does not support arm64)
- Latest minikube version

## Tutorial

First, start minikube:

```
$ minikube start
```


You can now run [BCC tools](https://github.com/iovisor/bcc) as a Docker container in minikube:

```shell
$ minikube ssh -- docker run --rm   --privileged   -v /lib/modules:/lib/modules:ro   -v /usr/src:/usr/src:ro   -v /etc/localtime:/etc/localtime:ro   --workdir /usr/share/bcc/tools   ghcr.io/iovisor/bcc. ./execsnoop




