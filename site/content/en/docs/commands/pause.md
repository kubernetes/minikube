---
title: "pause"
description: >
  pause Kubernetes
---


## minikube pause

pause Kubernetes

### Synopsis

pause Kubernetes

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
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the Kubernetes cluster. (default "kubeadm")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

