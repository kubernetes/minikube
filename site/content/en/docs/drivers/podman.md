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

The podman driver is an alternative container runtime to the [Docker]({{< ref "/docs/drivers/docker.md" >}}) driver.

## Requirements

- Linux or macOS operating systems on amd64 architecture
- Install [podman](https://podman.io/getting-started/installation.html)

{{% readfile file="/docs/drivers/includes/podman_usage.inc" %}}

## Known Issues

- Podman driver is not supported on non-amd64 architectures such as arm yet. For non-amd64 archs please use [other drivers]({{< ref "/docs/drivers/_index.md" >}})
- Podman v2 driver is not supported yet.
- Podman requirements passwordless running of sudo. If you run into an error about sudo, do the following:

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

## Troubleshooting

- Run `minikube start --alsologtostderr -v=7` to debug errors and crashes
