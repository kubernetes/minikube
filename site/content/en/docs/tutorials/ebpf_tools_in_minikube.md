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

## Prerequisites

- Latest minikube binary

## Tutorial

First, start minikube:

```
# FIXME: what is the new URL for minikube.iso that includes https://github.com/kubernetes/minikube/pull/8582?
$ ./out/minikube start --iso-url file://$(pwd)/out/minikube.iso --driver=kvm2
```

You can now run [BCC tools](https://github.com/iovisor/bcc) as a Docker container in minikube:

```shell
$ minikube ssh -- docker run --rm --privileged -ti --workdir /usr/share/bcc/tools kinvolk/bcc ./execsnoop
Unable to find image 'kinvolk/bcc:latest' locally
latest: Pulling from kinvolk/bcc
23884877105a: Pull complete 
bc38caa0f5b9: Pull complete 
2910811b6c42: Pull complete 
36505266dcc6: Pull complete 
7cce3becaab1: Pull complete 
900cb26d625f: Pull complete 
Digest: sha256:f49695286ac5e896e98d9e9165e062e62212b6e2fa0195f577aa2409eb40a6cd
Status: Downloaded newer image for kinvolk/bcc:latest
PCOMM            PID    PPID   RET ARGS
runc             3321   869      0 /usr/bin/runc --version
docker-init      3327   869      0 /usr/bin/docker-init --version
^C
```

You can also run [Inspektor Gadget](https://github.com/kinvolk/inspektor-gadget) in minikube:

```shell
$ kubectl krew install gadget
$ kubectl gadget deploy | kubectl apply -f -
```

When the gadget deployment is ready, you can run the execsnoop gadget in two terminals:

```shell
$ kubectl gadget execsnoop -n default -l run=shell
Node numbers: 0 = minikube
NODE PCOMM            PID    PPID   RET ARGS
[ 0] date             4799   4783     0 /bin/date
[ 0] sleep            4800   4783     0 /bin/sleep 1
[ 0] date             4801   4783     0 /bin/date
[ 0] sleep            4802   4783     0 /bin/sleep 1
^C
```

```shell
$ kubectl run --image=busybox shell -- sh -c 'while sleep 1 ; do date ; done'
```
