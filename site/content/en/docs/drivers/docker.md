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

- The following Docker runtime security options are currently *unsupported and will not work* with the Docker driver:
  - [userns-remap](https://docs.docker.com/engine/security/userns-remap/)
  - [rootless](https://docs.docker.com/engine/security/rootless/)

- Docker driver is not supported on non-amd64 architectures such as arm yet. For non-amd64 archs please use [other drivers]({{< ref "/docs/drivers/_index.md" >}}) 

- On macOS, containers might get hung and require a restart of Docker for Desktop. See [docker/for-mac#1835](https://github.com/docker/for-mac/issues/1835)

- The `ingress`, and `ingress-dns` addons are currently only supported on Linux. See [#7332](https://github.com/kubernetes/minikube/issues/7332)

- On WSL2 (experimental - see [#5392](https://github.com/kubernetes/minikube/issues/5392)), you may need to run:

   `sudo mkdir /sys/fs/cgroup/systemd && sudo mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd`.

## Troubleshooting

[comment]: <> (this title is used in the docs links, don't change)
### Verify Docker container type is Linux

- On Windows, make sure Docker Desktop's container type setting is Linux and not windows. see docker docs on [switching container type](https://docs.docker.com/docker-for-windows/#switch-between-windows-and-linux-containers). 
You can verify your Docker container type by running:
   ```shell
   docker info --format '{{.OSType}}'
   ```

### Run with logs

- Run `--alsologtostderr -v=1` for extra debugging information

### Deploying MySql on a linux with AppArmor

- On Linux, if you want to run MySQL pod, you need to disable AppArmor for mysql profile

   If your docker has [AppArmor](https://wiki.ubuntu.com/AppArmor) enabled, running mysql in privileged mode with docker driver will have the issue [#7401](https://github.com/kubernetes/minikube/issues/7401).
   There is a workaround - see [moby/moby#7512](https://github.com/moby/moby/issues/7512#issuecomment-61787845).
