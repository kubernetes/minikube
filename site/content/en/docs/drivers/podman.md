---
title: "podman"
weight: 3
aliases:
    - /docs/reference/drivers/podman
---

## Overview

{{% pageinfo %}}
This driver is experimental and in active development. Help wanted!
{{% /pageinfo %}}

The podman driver is another kubernetes in container driver for minikube. similar to [docker](https://minikube.sigs.k8s.io/Drivers/docker/) driver. The podman driver is  experimental, and only supported on Linux and macOS (with a remote podman server).



## Requirements
- amd64 system 
- Latest version of [Podman](https://podman.io/getting-started/installation) 


## Try it with CRI-O container runtime.

```shell
minikube start --driver=podman --container-runtime=cri-o
```

{{% readfile file="/docs/drivers/includes/podman_usage.inc" %}}

## Known Issues
- Podman driver is not supported on non-amd64 architectures such as arm yet. For non-amd64 archs please use [other drivers]({{< ref "/docs/drivers/" >}} 
