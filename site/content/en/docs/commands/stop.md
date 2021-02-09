---
title: "stop"
description: >
  Stops a running local Kubernetes cluster
---


## minikube stop

Stops a running local Kubernetes cluster

### Synopsis

Stops a local Kubernetes cluster. This command stops the underlying VM or container, but keeps user data intact. The cluster can be started again with the "start" command.

```shell
minikube stop [flags]
```

### Options

```
      --all                   Set flag to stop all profiles (clusters)
      --cancel-scheduled      cancel any existing scheduled stop requests
      --keep-context-active   keep the kube-context active after cluster is stopped. Defaults to false.
  -o, --output string         Format to print stdout in. Options include: [text,json] (default "text")
      --schedule duration     Set flag to stop cluster after a set amount of time (e.g. --schedule=5m)
```

### Options inherited from parent commands

```
      --add_dir_header                   If true, adds the file directory to the header of the log messages
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the Kubernetes cluster. (default "kubeadm")
  -h, --help                             
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --log_file string                  If non-empty, use this log file
      --log_file_max_size uint           Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                      log to standard error instead of files
      --one_output                       If true, only write logs to their native severity level (vs also writing to each lower severity level
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --skip_headers                     If true, avoid header prefixes in the log messages
      --skip_log_headers                 If true, avoid headers when opening log files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
      --user string                      Specifies the user executing the operation. Useful for auditing operations executed by 3rd party tools. Defaults to the operating system username.
  -v, --v Level                          number for the log level verbosity
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

