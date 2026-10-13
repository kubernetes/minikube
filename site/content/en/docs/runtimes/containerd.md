---
title: "containerd"
aliases:
    - /docs/reference/runtimes/containerd
---

Home page: <https://containerd.io>

<https://github.com/containerd/containerd>  
<https://github.com/moby/buildkit>

## Requirements

The containerd runtime requires the host memory cgroup controller. If the
controller is disabled (for example with `cgroup_disable=memory` on the kernel
cmdline), `minikube start` fails while kubelet and runc report
`controller "memory" not available`.

Check whether your host has it:

```shell
cat /sys/fs/cgroup/cgroup.controllers
```

The output must list `memory`. If it does not, enable the memory controller in
your host kernel config and reboot. As a workaround, use the docker runtime:

```shell
minikube start --container-runtime=docker
```
