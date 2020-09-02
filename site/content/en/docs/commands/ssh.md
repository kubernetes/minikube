---
title: "ssh"
description: >
  Log into the minikube environment (for debugging)
---


## minikube ssh

Log into the minikube environment (for debugging)

### Synopsis

Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'.

```
minikube ssh [flags]
```

### Options

```
  -h, --help          help for ssh
      --native-ssh    Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'. (default true)
  -n, --node string   The node to ssh into. Defaults to the primary control plane.
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

