---
title: "kubectl"
description: >
  Run a kubectl binary matching the cluster version
---


## minikube kubectl

Run a kubectl binary matching the cluster version

### Synopsis

Run the Kubernetes client, download it if necessary. Remember -- after kubectl!

Examples:
minikube kubectl -- --help
minikube kubectl -- get pods --namespace kube-system

```
minikube kubectl [flags]
```

### Options

```
  -h, --help   help for kubectl
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

