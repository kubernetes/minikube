## minikube config

Modify minikube config

### Synopsis


config modifies minikube config files using subcommands like "minikube config set vm-driver kvm" 
Configurable fields: 

 * vm-driver
 * v
 * cpus
 * disk-size
 * host-only-cidr
 * memory
 * show-libmachine-logs
 * log_dir
 * kubernetes-version
 * WantUpdateNotification
 * ReminderWaitPeriodInHours
 * WantReportError
 * WantReportErrorPrompt
 * dashboard
 * addon-manager
 * kube-dns

```
minikube config SUBCOMMAND [flags]
```

### Options inherited from parent commands

```
      --alsologtostderr value    log to standard error as well as files
      --log_backtrace_at value   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir value            If non-empty, write log files in this directory
      --logtostderr value        log to standard error instead of files
      --show-libmachine-logs     Whether or not to show logs from libmachine.
      --stderrthreshold value    logs at or above this threshold go to stderr (default 2)
  -v, --v value                  log level for V logs
      --vmodule value            comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minikube](minikube.md)	 - Minikube is a tool for managing local Kubernetes clusters.
* [minikube config get](minikube_config_get.md)	 - Gets the value of PROPERTY_NAME from the minikube config file
* [minikube config set](minikube_config_set.md)	 - Sets an individual value in a minikube config file
* [minikube config unset](minikube_config_unset.md)	 - unsets an individual value in a minikube config file
* [minikube config view](minikube_config_view.md)	 - Display values currently set in the minikube config file

