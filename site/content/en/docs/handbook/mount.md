---
title: "Mounting filesystems"
date: 2017-01-05
weight: 12
description: >
  How to mount a host directory into the VM
aliases:
  - /docs/tasks/mount
---

## 9P Mounts

9P mounts are flexible and work across all hypervisors, but suffers from performance and reliability issues when used with large folders (>600 files). See **Driver Mounts** as an alternative.

To mount a directory from the host into the guest using the `mount` subcommand:

```shell
minikube mount <source directory>:<target directory>
```

For example, this would mount your home directory to appear as /host within the minikube VM:

```shell
minikube mount $HOME:/host
```

This directory may then be referenced from a Kubernetes manifest, for example:

```shell
{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "ubuntu"
  },
  "spec": {
    "containers": [
      {
        "name": "ubuntu",
        "image": "ubuntu:18.04",
        "args": ["bash"],
        "stdin": true,
        "stdinOnce": true,
        "tty": true,
        "workingDir": "/host",
        "volumeMounts": [
          {
            "mountPath": "/host",
            "name": "host-mount"
          }
        ]
      }
    ],
    "volumes": [
      {
        "name": "host-mount",
        "hostPath": {
          "path": "/host"
        }
      }
    ]
  }
}
```

## Driver mounts

Some hypervisors, have built-in host folder sharing. Driver mounts are reliable with good performance, but the paths are not predictable across operating systems or hypervisors:

| Driver | OS | HostFolder | VM |
| --- | --- | --- | --- |
| VirtualBox | Linux | /home | /hosthome |
| VirtualBox | macOS | /Users | /Users |
| VirtualBox | Windows | C://Users | /c/Users |
| KVM | Linux | Unsupported | |
| HyperKit | macOS | Supported |  |

These mounts can be disabled by passing `--disable-driver-mounts` to `minikube start`.

HyperKit mounts can use the following flags:
`--nfs-share=[]`: Local folders to share with Guest via NFS mounts
`--nfs-shares-root='/nfsshares'`: Where to root the NFS Shares, defaults to /nfsshares

## File Sync

See [File Sync]({{<ref "filesync.md" >}})
