---
title: "mount"
description: >
  Mounts the specified directory into minikube
---


## minikube mount

Mounts the specified directory into minikube

### Synopsis

Mounts the specified directory into minikube.

```
minikube mount [flags] <source directory>:<target directory>
```

### Options

```
      --9p-version string   Specify the 9p version that the mount should use (default "9p2000.L")
      --gid string          Default group id used for the mount (default "docker")
  -h, --help                help for mount
      --ip string           Specify the ip that the mount should be setup on
      --kill                Kill the mount process spawned by minikube start
      --mode uint           File permissions used for the mount (default 493)
      --msize int           The number of bytes to use for 9p packet payload (default 262144)
      --options strings     Additional mount options, such as cache=fscache
      --type string         Specify the mount filesystem type (supported types: 9p) (default "9p")
      --uid string          Default user id used for the mount (default "docker")
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

