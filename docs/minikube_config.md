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
 * log_dir
 * kubernetes-version
 * iso-url
 * WantUpdateNotification
 * ReminderWaitPeriodInHours
 * WantReportError
 * WantReportErrorPrompt
 * WantKubectlDownloadMsg
 * dashboard
 * addon-manager
 * kube-dns
 * heapster
 * ingress
 * registry-creds
 * default-storageclass
 * hyperv-virtual-switch
 * use-vendored-driver

```
minikube config SUBCOMMAND [flags]
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory (default "")
      --logtostderr                      log to standard error instead of files
      --show-libmachine-logs             Deprecated: To enable libmachine logs, set --v=3 or higher
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
      --use-vendored-driver              Use the vendored in drivers instead of RPC
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minikube](minikube.md)	 - Minikube is a tool for managing local Kubernetes clusters.
* [minikube config get](minikube_config_get.md)	 - Gets the value of PROPERTY_NAME from the minikube config file
* [minikube config set](minikube_config_set.md)	 - Sets an individual value in a minikube config file
* [minikube config unset](minikube_config_unset.md)	 - unsets an individual value in a minikube config file
* [minikube config view](minikube_config_view.md)	 - Display values currently set in the minikube config file

