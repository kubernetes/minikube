---
title: "podman"
linkTitle: "podman"
weight: 3
date: 2020-03-26
description: >
  Podman driver 
---

## Overview

The podman driver is another kubernetes in container driver for minikube. simmilar to [docker](https://minikube.sigs.k8s.io/docs/reference/drivers/docker/) driver.
podman driver is currently experimental. 
and only supported on Linux and MacOs (with a remote podman server)


## Try it with CRI-O container runtime.
```shell
minikube start --driver=podman --container-runtime=cri-o
```


{{% readfile file="/docs/Reference/Drivers/includes/podman_usage.inc" %}}



