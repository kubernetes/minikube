---
title: "docker"
linkTitle: "docker"
weight: 3
date: 2020-02-05
description: >
  Docker driver 
---

## Overview

The Docker driver is the newest minikube driver. which runs kubernetes in container VM-free ! with full feature parity with minikube in VM.

{{% readfile file="/docs/Reference/Drivers/includes/docker_usage.inc" %}}


## Special features
- Cross platform (linux, macos, windows)
- No hypervisor required when run on Linux.

## Known Issues.

- The 'ingress' and 'ingress-dns' addons is only supported on Linux and they are  not supported in Docker Driver on MacOS and Windows yet. to get updates on the work in progress please check [issue page](https://github.com/kubernetes/minikube/issues/7332)

- a known [docker issue for MacOs](https://github.com/docker/for-mac/issues/1835), a containers on Docker on MacOS might hang and get stuck while other containers can get created. The current workaround is restarting docker.