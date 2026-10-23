---
title: "podman"
weight: 3
aliases:
    - /docs/reference/drivers/podman
---

## Overview

The podman driver is an alternative container runtime to the [Docker]({{< ref "/docs/drivers/docker.md" >}}) driver.

## Requirements

- Install [podman](https://podman.io/getting-started/installation.html)

### Rootless host prerequisites

With `--driver=podman --rootless`, minikube runs as your non-root user, so the host
must grant that user subordinate UID/GID ranges in `/etc/subuid` and `/etc/subgid`.
Without them, `minikube start --driver=podman --rootless` fails.

Tested configuration (Ubuntu 24.04):

```shell
sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 $USER
podman system migrate
```

This was verified on Ubuntu 24.04; please validate and report results for other
distributions (Fedora, Debian, etc.). See the upstream
[podman rootless tutorial](https://github.com/podman-container-tools/podman/blob/main/docs/tutorials/rootless_tutorial.md)
for details on `/etc/subuid` and `/etc/subgid` configuration.

{{% readfile file="/docs/drivers/includes/podman_usage.inc" %}}

## Known Issues

- On Linux, Podman requires passwordless running of sudo. If you run into an error about sudo, do the following:

```shell
$ sudo visudo
```
Then append the following to the section *at the very bottom* of the file where `username` is your user account.

```shell
username ALL=(ALL) NOPASSWD: /usr/bin/podman
```

Be sure this text is *after* `#includedir /etc/sudoers.d`. To confirm it worked, try:

```shell
sudo -k -n podman version
```

- On all other operating systems, make sure to create and start the virtual machine that is needed for Podman.

```shell
podman machine init --cpus 2 --memory 2048 --disk-size 20
podman machine start
podman system connection default podman-machine-default-root
podman info
```

Also see [co/podman-driver open issues](https://github.com/kubernetes/minikube/labels/co%2Fpodman-driver).

## Troubleshooting

- Run `minikube start --alsologtostderr -v=7` to debug errors and crashes
