---
title: "Caching images"
date: 2019-08-05
weight: 1
description: >
  How to cache arbitrary Docker images
---

## Overview

For offline use and performance reasons, minikube caches required Docker images onto the local file system. Developers may find it useful to add their own images to this cache for local development.

## Adding an image

To add the ubuntu 16.04 image to minikube's image cache:

```shell
minikube cache add ubuntu:16.04
```

The add command will store the requested image to `$MINIKUBE_HOME/cache/images`, and load it into the VM's container runtime environment next time `minikube start` is called.

## Listing images

To display images you have added to the cache:

```shell
minikube cache list
```

This listing will not include the images which are built-in to minikube.

## Deleting an image

```shell
minikube cache delete <image name>
```

### Additional Information

* [Reference: Disk Cache]({{< ref "/docs/reference/disk_cache.md" >}})
* [Reference: cache command]({{< ref "/docs/reference/commands/cache.md" >}})