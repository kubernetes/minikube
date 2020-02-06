---
title: "config"
linkTitle: "config"
weight: 1
date: 2019-08-01
description: >
  Modify minikube config
---

### Overview

config modifies minikube config files using subcommands like "minikube config set vm-driver kvm"

Configurable fields: 

 * vm-driver
 * container-runtime
 * feature-gates
 * v
 * cpus
 * disk-size
 * host-only-cidr
 * memory
 * log_dir
 * kubernetes-version
 * iso-url
 * WantUpdateNotification
 * ReminderWaitPeriodInHours
 * WantReportError
 * WantReportErrorPrompt
 * WantKubectlDownloadMsg
 * WantNoneDriverWarning
 * profile
 * bootstrapper
 * ShowDriverDeprecationNotification
 * ShowBootstrapperDeprecationNotification
 * insecure-registry
 * hyperv-virtual-switch
 * disable-driver-mounts
 * cache
 * embed-certs
 * native-ssh


### subcommands

- **get**: Gets the value of PROPERTY_NAME from the minikube config file

## minikube config get

Returns the value of PROPERTY_NAME from the minikube config file.  Can be overwritten at runtime by flags or environmental variables.

### Usage

```
minikube config get PROPERTY_NAME [flags]
```

## minikube config set

Sets the PROPERTY_NAME config value to PROPERTY_VALUE
	These values can be overwritten by flags or environment variables at runtime.

### Usage

```
minikube config set PROPERTY_NAME PROPERTY_VALUE [flags]
```

## minikube config unset

unsets PROPERTY_NAME from the minikube config file.  Can be overwritten by flags or environmental variables

### Usage

```
minikube config unset PROPERTY_NAME [flags]
```


## minikube config view

### Overview

Display values currently set in the minikube config file.

### Usage

```
minikube config view [flags]
```

### Options

```
      --format string   Go template format string for the config view output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
                        For the list of accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd/config#ConfigViewTemplate (default "- {{.ConfigKey}}: {{.ConfigValue}}\n")
  -h, --help            help for view
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
