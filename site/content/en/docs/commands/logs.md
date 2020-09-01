---
title: "logs"
description: >
  Returns logs to debug a local Kubernetes cluster
---


## minikube logs

Returns logs to debug a local Kubernetes cluster

### Synopsis

Gets the logs of the running instance, used for debugging minikube, not user code.

```
minikube logs [flags]
```

### Options

```
  -f, --follow        Show only the most recent journal entries, and continuously print new entries as they are appended to the journal.
  -h, --help          help for logs
  -n, --length int    Number of lines back to go within the log (default 60)
      --node string   The node to get logs from. Defaults to the primary control plane.
      --problems      Show only log entries which point to known problems
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

