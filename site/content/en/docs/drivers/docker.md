---
title: "docker"
weight: 3
aliases:
    - /docs/reference/drivers/docker
---

## Overview

The Docker driver allows you to install Kubernetes into an existing Docker install. On Linux, this does not require virtualization to be enabled.

{{% tabs %}}
{{% tab "Standard Docker" %}}
## Requirements

- [Install Docker](https://docs.docker.com/engine/install/) 18.09 or higher (20.10 or higher is recommended)
- amd64 or arm64 system.
- If using WSL complete [these steps]({{<ref "/docs/tutorials/wsl_docker_driver">}}) first
- Don't forget to follow this [step](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user) to manage Docker as a non-root user.

## Usage

Start a cluster using the docker driver:

```shell
minikube start --driver=docker
```
To make docker the default driver:

```shell
minikube config set driver docker
```
{{% /tab %}}
{{% tab "Rootless Docker" %}}
## Requirements
- Docker 20.10 or higher, see https://rootlesscontaine.rs/getting-started/docker/
- Cgroup v2 delegation, see https://rootlesscontaine.rs/getting-started/common/cgroup2/
- Kernel 5.11 or later (5.13 or later is recommended when SELinux is enabled), see https://rootlesscontaine.rs/how-it-works/overlayfs/

## Usage

Start a cluster using the rootless docker driver:

```shell
dockerd-rootless-setuptool.sh install -f
docker context use rootless
minikube start --driver=docker --container-runtime=containerd
```

Unlike Podman driver, it is not necessary to set the `rootless` property of minikube (`minikube config set rootless true`).
When the `rootless` property is explicitly set but the current Docker host is not rootless, minikube fails with an error.

It is recommended to set the `--container-runtime` flag to "containerd".
{{% /tab %}}
{{% /tabs %}}

## Special features

- Cross platform (linux, macOS, Windows)
- No hypervisor required when run on Linux
- Experimental support for [WSL2](https://docs.microsoft.com/en-us/windows/wsl/wsl2-install) on Windows 10

## Known Issues

- The following Docker runtime security options are currently *unsupported and will not work* with the Docker driver (see [#9607](https://github.com/kubernetes/minikube/issues/9607)):
  - [userns-remap](https://docs.docker.com/engine/security/userns-remap/)

- On macOS, containers might get hung and require a restart of Docker for Desktop. See [docker/for-mac#1835](https://github.com/docker/for-mac/issues/1835)

- The `ingress`, and `ingress-dns` addons are currently only supported on Linux. See [#7332](https://github.com/kubernetes/minikube/issues/7332)

- On WSL2 (experimental - see [#5392](https://github.com/kubernetes/minikube/issues/5392)), you may need to run:

   `sudo mkdir /sys/fs/cgroup/systemd && sudo mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd`.

Also see [co/docker-driver open issues](https://github.com/kubernetes/minikube/labels/co%2Fdocker-driver).

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
