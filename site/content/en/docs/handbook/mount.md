---
title: "Mounting filesystems"
date: 2017-01-05
weight: 12
description: >
  How to mount a host directory into a cluster
aliases:
  - /docs/tasks/mount
---

minikube supports the following methods to mount host directories into a cluster:

| Method                                    | Performance                  | Flexibility                |
|-------------------------------------------|------------------------------|----------------------------|
| [Mount during start](#mount-during-start) | Near-native with new drivers | Single mount, set at start |
| [Mount command](#mount-command)           | Limited (9p only)            | Multiple mounts, any time  |
| [Driver mounts](#driver-mounts)           | Varies                       | Legacy drivers only        |

## Mount During Start

The recommended way to mount a host directory. Best for directories that should
be available on every run. Uses the best method supported by the driver —
virtiofs or container volumes on modern drivers provide near-native performance,
while legacy drivers fall back to 9p.

```shell
minikube start --mount-string <host directory>:<guest directory>
```

For example, this would mount your `~/models` directory to appear as
`/mnt/models` within the cluster:

```shell
minikube start --mount-string ~/models:/mnt/models
```

The directory remains mounted while the cluster is running.

| Driver         | OS      | Method           | Notes                  |
|----------------|---------|------------------|------------------------|
| `docker`       | All     | container volume |                        |
| `podman`       | All     | container volume |                        |
| `vfkit`        | macOS   | virtiofs         | Since minikube 1.37.0  |
| `krunkit`      | macOS   | virtiofs         | Since minikube 1.37.0  |
| `hyperv`       | Windows | 9p               |                        |
| `kvm`          | Linux   | 9p               |                        |
| `qemu`         | macOS   | 9p               | Requires socket_vmnet  |

### Notes

- Some drivers do not support mounting (see
  [unsupported drivers](#unsupported-drivers)).
- Only a single mount is supported. If you need to mount additional directories,
  or mount and unmount directories after the cluster was started, use the
  [mount command](#mount-command).

## Mount Command

Mounts a host directory into a running cluster using the *9p* filesystem. Use
this when you need to mount multiple directories, or for temporary mounts whose
lifecycle is shorter than the cluster (e.g. mount a directory during a test run,
then unmount). Works with most drivers (see [unsupported drivers](#unsupported-drivers))
but suffers from performance and reliability issues with large directories (>600
files). Prefer [mount during start](#mount-during-start) when possible.

```shell
minikube mount <host directory>:<guest directory>
```

For example:

```shell
minikube mount ~/models:/mnt/models
```

The directory remains mounted while the mount command is running. To unmount,
terminate the command with `Ctrl+C`.

## Unsupported Drivers

The following drivers do not support mounting host directories:

- `none`
- `qemu` with the builtin network — use `--network=socket_vmnet` (macOS only,
  selected automatically when socket_vmnet is installed)

## Driver mounts

Some legacy drivers have built-in host folder sharing. These mounts are
automatic but the paths are not configurable.

{{% alert title="Warning" color="warning" %}}
Driver mounts expose large host directories (e.g. `/Users`, `/home`) to the
guest VM by default. This is a security risk — any process in the VM has read
and write access to your entire home directory. Consider disabling driver mounts
with `--disable-driver-mounts` and using `--mount-string` to share only the
specific directories you need.
{{% /alert %}}

| Driver         | OS      | Host         | Guest             |
|----------------|---------|--------------|-------------------|
| VirtualBox     | Linux   | /home        | /hosthome         |
| VirtualBox     | macOS   | /Users       | /Users            |
| VirtualBox     | Windows | C://Users    | /c/Users          |
| VMware Fusion  | macOS   | /Users       | /mnt/hgfs/Users   |

Built-in mounts can be disabled by passing the `--disable-driver-mounts` flag to
`minikube start`.

## File Sync

See [File Sync]({{<ref "filesync.md" >}})
