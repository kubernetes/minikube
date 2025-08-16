---
title: "Mounting filesystems"
date: 2017-01-05
weight: 12
description: >
  How to mount a host directory into the VM
aliases:
  - /docs/tasks/mount
---

## Mount During Start

To mount a directory from the host into the guest using the `start` subcommand
use the `--mount-string` flag:

```shell
minikube start --mount-string <host directory>:<guest directory>
```

For example, this would mount your `~/models` directory to appear as
`/mnt/models` within the minikube VM.

```shell
minikube start ~/models:/mnt/models
```

The directory remains mounted while the cluster is running.

The mount will use the best method supported by the driver:

| Driver         | OS      | Method           | Notes                  |
|----------------|---------|------------------|------------------------|
| docker         | Linux   | container volume |                        |
| podman         | Linux   | container volume |                        |
| kvm            | Linux   | 9p               |                        |
| Vfkit          | macOS   | virtiofs         | Since minikube 1.37.0  |
| Krunkit        | macOS   | virtiofs         | Since minikube 1.37.0  |
| qemu           | macOS   | 9p               |                        |

The `--mount-string` flag supports only a single mount. If you want to mount
additional directories, or mount and unmount directories after the cluster was
started see [Mount Command](#Mount-command).

## Mount Command

The mount command allows mounting and unmounting directories after the minikube
cluster was started using the *9p* filesystem. The command works with all VM
drivers.

To mount a directory from the host into the guest using the `mount` subcommand:

```shell
minikube mount <host directory>:<guest> directory>
```

For example, this would mount your `~/models` directory to appear as
`/mnt/models` within the minikube VM:

```shell
minikube mount ~/models:/mnt/models
```

The directory remains mounted while the mount command is running. To unmount the
directory terminate the mount command with `Control+C`.

*9p* mounts are flexible and work across all hypervisors, but suffers from
performance and reliability issues when used with large folders (>600 files).
See [Driver Automatic Mounts](#driver-automatic-mounts) and
[Driver Specific Mounts](#driver-specific-mounts) as alternatives.

## Driver automatic mounts

Some hypervisors, have built-in host folder sharing. Driver mounts are reliable
with good performance, but the paths may differ across operating systems or
hypervisors:

| Driver         | OS      | Host         | Guest             |
|----------------|---------|--------------|-------------------|
| VirtualBox     | Linux   | /home        | /hosthome         |
| VirtualBox     | macOS   | /Users       | /Users            |
| VirtualBox     | Windows | C://Users    | /c/Users          |
| VMware Fusion  | macOS   | /Users       | /mnt/hgfs/Users   |

Built-in mounts can be disabled by passing the `--disable-driver-mounts` flag to
`minikube start`.

## Driver Specific Mounts

HyperKit driver can also use the following start flags:
- `--nfs-share=[]`: Local folders to share with Guest via NFS mounts
- `--nfs-shares-root='/nfsshares'`: Where to mount the NFS Shares, defaults to
  `/nfsshares`

## File Sync

See [File Sync]({{<ref "filesync.md" >}})
