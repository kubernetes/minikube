---
title: "pause"
linkTitle: "pause"
weight: 1
date: 2020-02-05
description: >
  pause the Kubernetes control plane or other namespaces
---

### Overview

The pause command allows you to freeze containers using the Linux [cgroup freezer](https://www.kernel.org/doc/Documentation/cgroup-v1/freezer-subsystem.txt). Once frozen, processes will no longer consume CPU cycles, but will remain in memory.

By default, the pause command will pause the Kubernetes control plane (kube-system namespace), leaving your applications running. This reduces the background CPU usage of a minikube cluster to a negligable 2-3% of a CPU.

### Usage

```
minikube pause [flags]
```

### Options

```
  -n, ----namespaces strings   namespaces to pause (default [kube-system,kubernetes-dashboard,storage-gluster,istio-operator])
  -A, --all-namespaces         If set, pause all namespaces
  -h, --help                   help for pause
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [unpause](unpause.md)

