---
title: "addons"
description: >
  Enable or disable a minikube addon
---


## minikube addons

Enable or disable a minikube addon

### Synopsis

addons modifies minikube addons files using subcommands like "minikube addons enable dashboard"

```
minikube addons SUBCOMMAND [flags]
```

### Options

```
  -h, --help   help for addons
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

## minikube addons configure

Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list 

### Synopsis

Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list 

```
minikube addons configure ADDON_NAME [flags]
```

### Options

```
  -h, --help   help for configure
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

## minikube addons disable

Disables the addon w/ADDON_NAME within minikube (example: minikube addons disable dashboard). For a list of available addons use: minikube addons list 

### Synopsis

Disables the addon w/ADDON_NAME within minikube (example: minikube addons disable dashboard). For a list of available addons use: minikube addons list 

```
minikube addons disable ADDON_NAME [flags]
```

### Options

```
  -h, --help   help for disable
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

## minikube addons enable

Enables the addon w/ADDON_NAME within minikube (example: minikube addons enable dashboard). For a list of available addons use: minikube addons list 

### Synopsis

Enables the addon w/ADDON_NAME within minikube (example: minikube addons enable dashboard). For a list of available addons use: minikube addons list 

```
minikube addons enable ADDON_NAME [flags]
```

### Options

```
  -h, --help   help for enable
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

## minikube addons help

Help about any command

### Synopsis

Help provides help for any command in the application.
Simply type addons help [path to command] for full details.

```
minikube addons help [command] [flags]
```

### Options

```
  -h, --help   help for help
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

## minikube addons list

Lists all available minikube addons as well as their current statuses (enabled/disabled)

### Synopsis

Lists all available minikube addons as well as their current statuses (enabled/disabled)

```
minikube addons list [flags]
```

### Options

```
  -h, --help            help for list
  -o, --output string   minikube addons list --output OUTPUT. json, list (default "list")
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

## minikube addons open

Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list 

### Synopsis

Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list 

```
minikube addons open ADDON_NAME [flags]
```

### Options

```
      --format string   Format to output addons URL in.  This format will be applied to each url individually and they will be printed one at a time. (default "http://{{.IP}}:{{.Port}}")
  -h, --help            help for open
      --https           Open the addons URL with https instead of http
      --interval int    The time interval for each check that wait performs in seconds (default 1)
      --url             Display the Kubernetes addons URL in the CLI instead of opening it in the default browser
      --wait int        Amount of time to wait for service in seconds (default 2)
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

