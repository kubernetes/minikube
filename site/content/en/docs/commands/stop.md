---
title: "stop"
description: >
  Stops a running local Kubernetes cluster
---


## minikube stop

Stops a running local Kubernetes cluster

### Synopsis

Stops a local Kubernetes cluster running in Virtualbox. This command stops the VM
itself, leaving all files intact. The cluster can be started again with the "start" command.

```
minikube stop [flags]
```

### Options

```
      --all                   Set flag to stop all profiles (clusters)
  -h, --help                  help for stop
      --keep-context-active   keep the kube-context active after cluster is stopped. Defaults to false.
      --schedule string       Schedule stop for this cluster (e.g. --schedule=5m) 
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

