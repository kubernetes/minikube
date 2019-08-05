---
title: "Filesystem mounts"
date: 2017-01-05
weight: 4
description: >
  How to mount a host directory into the VM
---

It is possible to mount directories from the host into the guest using the `mount` subcommand:

```
minikube mount <source directory>:<target directory>
```

For example, this would mount your home directory to appear as /host within the minikube VM:

```
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
            "args": [
              "bash"
            ],
            "stdin": true,
            "stdinOnce": true,
            "tty": true,
            "workingDir": "/host",
            "volumeMounts": [{
              "mountPath": "/host",
              "name": "host-mount"
            }]
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

