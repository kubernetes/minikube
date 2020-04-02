---
title: "generate-docs"
linkTitle: "generate-docs"
weight: 1
date: 2020-04-02
description: >
  Populates the specified folder with documentation in markdown about minikube
---



## minikube generate-docs

Populates the specified folder with documentation in markdown about minikube

### Synopsis

Populates the specified folder with documentation in markdown about minikube

```
minikube generate-docs [flags]
```

### Examples

```
minikube generate-docs --path <FOLDER_PATH>
```

### Options

```
  -h, --help          help for generate-docs
      --path string   The path on the file system where the docs in markdown need to be saved
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

