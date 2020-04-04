---
title: "docker"
weight: 3
aliases:
    - /docs/reference/drivers/docker
---

## Overview

The Docker driver is the newest minikube driver. which runs kubernetes in container VM-free ! with full feature parity with minikube in VM.

{{% readfile file="/docs/drivers/includes/docker_usage.inc" %}}

## Special features

- Cross platform (linux, macos, windows)
- No hypervisor required when run on Linux.

## Known Issues.

- The 'ingress' and 'ingress-dns' addons are only supported on Linux.
these addons are not supported for Docker Driver on MacOS and Windows yet. to get updates on the work in progress please check [issue page](https://github.com/kubernetes/minikube/issues/7332)

- a known [docker issue for macOS](https://github.com/docker/for-mac/issues/1835), a containers on Docker on MacOS might hang and get stuck while other containers can get created. The current workaround is restarting docker.
