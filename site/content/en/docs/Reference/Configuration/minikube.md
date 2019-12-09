---
title: "minikube"
linkTitle: "minikube"
weight: 2
date: 2019-08-01
description: >
  minikube configuration reference
---

## Flags

Most minikube configuration is done via the flags interface. To see which flags are possible for the start command, run:

```shell
minikube start --help
```

For example:

```shell
minikube start --apiserver-port 9999
```

Many of these flags are also available to be set via persistent configuration or environment variables. 
While most flags are applicable to any command, some are globally scoped:

```
Flags:
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the kubernetes cluster. (default "kubeadm")
  -h, --help                             help for minikube
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used.  
                                         	This can be modified to allow for multiple minikube instances to be run independently (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

## Persistent Configuration

minikube allows users to persistently store new default values to be used across all profiles, using the `minikube config` command. This is done providing a property name, and a property value.

### Listing config properties

```shell
minikube config
```

Example:

```shell
Configurable fields: 

 * vm-driver
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
 * dashboard
 * addon-manager
 * default-storageclass
 * efk
 * ingress
 * registry
 * registry-creds
 * freshpod
 * default-storageclass
 * storage-provisioner
 * storage-provisioner-gluster
 * metrics-server
 * nvidia-driver-installer
 * nvidia-gpu-device-plugin
 * logviewer
 * gvisor
 * hyperv-virtual-switch
 * disable-driver-mounts
 * cache
 * embed-certs
```

### Listing your property overrides

```shell
minikube config view
```

Example output:

```shell
- memory: 4096
- registry: true
- vm-driver: vmware
- dashboard: true
- gvisor: true
```

### Setting a new property override


```shell
minikube config set <key> <value>
```

For example:

```shell
minikube config set vm-driver hyperkit
```

## Environment Configuration

### Config variables

minikube supports passing environment variables instead of flags for every value listed in `minikube config`.  This is done by passing an environment variable with the prefix `MINIKUBE_`.

For example the `minikube start --iso-url="$ISO_URL"` flag can also be set by setting the `MINIKUBE_ISO_URL="$ISO_URL"` environment variable.

### Other variables

Some features can only be accessed by environment variables, here is a list of these features:

* **MINIKUBE_HOME** - (string) sets the path for the .minikube directory that minikube uses for state/configuration

* **MINIKUBE_IN_STYLE** - (bool) manually sets whether or not emoji and colors should appear in minikube. Set to false or 0 to disable this feature, true or 1 to force it to be turned on.

* **MINIKUBE_WANTUPDATENOTIFICATION** - (bool) sets whether the user wants an update notification for new minikube versions

* **MINIKUBE_REMINDERWAITPERIODINHOURS** - (int) sets the number of hours to check for an update notification

* **CHANGE_MINIKUBE_NONE_USER** - (bool) automatically change ownership of ~/.minikube to the value of $SUDO_USER

* **MINIKUBE_ENABLE_PROFILING** - (int, `1` enables it) enables trace profiling to be generated for minikube

### Making environment variables persistent

To make the exported variables persistent:

* Linux and macOS: Add these declarations to `~/.bashrc` or wherever your shells environment variables are stored.
* Windows: Add these declarations via [system settings](https://support.microsoft.com/en-au/help/310519/how-to-manage-environment-variables-in-windows-xp) or using [setx](https://stackoverflow.com/questions/5898131/set-a-persistent-environment-variable-from-cmd-exe)

#### Example: Disabling emoji

```shell
export MINIKUBE_IN_STYLE=false
minikube start
```
