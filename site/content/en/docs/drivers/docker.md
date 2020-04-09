---
title: "docker"
weight: 3
aliases:
    - /docs/reference/drivers/docker
---

## Overview

The Docker driver allows you to install Kubernetes into an existing Docker install. On Linux, this does not require virtualization to be enabled.

{{% readfile file="/docs/drivers/includes/docker_usage.inc" %}}

## Special features

- Cross platform (linux, macOS, Windows)
- No hypervisor required when run on Linux
- Experimental support for [WSL2](https://docs.microsoft.com/en-us/windows/wsl/wsl2-install) on Windows 10

## Known Issues

- On macOS, containers might get hung and require a restart of Docker for Desktop. See [docker/for-mac#1835](https://github.com/docker/for-mac/issues/1835)

- The `ingress`, `ingress-dns` and `registry` addons are currently only supported on Linux. See [#7332](https://github.com/kubernetes/minikube/issues/7332) and [#7535](https://github.com/kubernetes/minikube/issues/7535)

- On WSL2 (experimental - see [#5392](https://github.com/kubernetes/minikube/issues/5392)), you may need to run:

   `sudo mkdir /sys/fs/cgroup/systemd && sudo mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd`.

- Addon 'registry' for mac and windows is not supported yet and it is [a work in progress](https://github.com/kubernetes/minikube/issues/7535).



## Troubleshooting

- On macOS or Windows, you may need to restart Docker for Desktop if a command gets hung
- Run `--alsologtostderr -v=1` for extra debugging information
