---
title: "docker-env"
description: >
  Configure environment to use minikube's Docker daemon
---


## minikube docker-env

Configure environment to use minikube's Docker daemon

### Synopsis

Sets up docker env variables; similar to '$(docker-machine env)'.

```
minikube docker-env [flags]
```

### Options

```
  -h, --help           help for docker-env
      --no-proxy       Add machine IP to NO_PROXY environment variable
      --shell string   Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect
  -u, --unset          Unset variables instead of setting them
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

