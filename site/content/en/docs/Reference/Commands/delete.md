---
title: "delete"
linkTitle: "delete"
weight: 1
date: 2019-08-01
description: >
  Deletes a local Kubernetes cluster
---

### Overview

Deletes a local Kubernetes cluster. This command deletes the VM, and removes all
associated files.

## Usage

```
minikube delete [flags]
```

##### Delete all profiles
```
minikube delete --all
```

##### Delete profile & `.minikube` directory
Do note that the following command only works if you have only 1 profile. If there are multiple profiles, the command will error out.
```
minikube delete --purge
```

##### Delete all profiles & `.minikube` directory
This will delete all the profiles and `.minikube` directory.
```
minikube delete --purge --all
```

### Flags

```
      --all: Set flag to delete all profiles
      --purge: Set this flag to delete the '.minikube' folder from your user directory.
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the kubernetes cluster. (default "kubeadm")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```
