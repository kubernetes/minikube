---
title: "addons"
linkTitle: "addons"
weight: 1
date: 2019-08-01
description: >
  Modifies minikube addons files using subcommands like "minikube addons enable dashboard"
---

## Overview

* **configure**:   Configures the addon w/ADDON_NAME within minikube
* **disable**:     Disables the addon w/ADDON_NAME within minikube
* **enable**:      Enables the addon w/ADDON_NAME within minikube
* **list**:        Lists all available minikube addons as well as their current statuses (enabled/disabled)
* **open**:        Opens the addon w/ADDON_NAME within minikube

## minikube addons configure

Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list 

```
minikube addons configure ADDON_NAME [flags]
```

## minikube addons disable

Disables the addon w/ADDON_NAME within minikube (example: minikube addons disable dashboard). For a list of available addons use: minikube addons list 

```
minikube addons disable ADDON_NAME [flags]
```

## minikube addons enable

Enables the addon w/ADDON_NAME within minikube (example: minikube addons enable dashboard). For a list of available addons use: minikube addons list 

```
minikube addons enable ADDON_NAME [flags]
```

or

```
minikube start --addons ADDON_NAME [flags]
```

## minikube addons list

Lists all available minikube addons as well as their current statuses (enabled/disabled)

```
minikube addons list [flags]
```

### Options

```
  -h, --help            help for list
  -o, --output string   minikube addons list --output OUTPUT. json, list (default "list")
```

## minikube addons open

Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list 

```
minikube addons open ADDON_NAME [flags]
```

### Options

```
      --format string   Format to output addons URL in.  This format will be applied to each url individually and they will be printed one at a time. (default "http://{{.IP}}:{{.Port}}")
  -h, --help            help for open
      --https           Open the addons URL with https instead of http
      --interval int    The time interval for each check that wait performs in seconds (default 6)
      --url             Display the kubernetes addons URL in the CLI instead of opening it in the default browser
      --wait int        Amount of time to wait for service in seconds (default 20)
```


## Options inherited from parent commands

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
