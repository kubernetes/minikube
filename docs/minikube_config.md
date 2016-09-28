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

```
minikube config SUBCOMMAND [flags]
```

### Options inherited from parent commands

```
      --alsologtostderr[=false]: log to standard error as well as files
      --log-flush-frequency=5s: Maximum number of seconds between log flushes
      --log_backtrace_at=:0: when logging hits line file:N, emit a stack trace
      --log_dir="": If non-empty, write log files in this directory
      --logtostderr[=false]: log to standard error instead of files
      --show-libmachine-logs[=false]: Whether or not to show logs from libmachine.
      --stderrthreshold=2: logs at or above this threshold go to stderr
      --v=0: log level for V logs
      --vmodule=: comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minikube](minikube.md)	 - Minikube is a tool for managing local Kubernetes clusters.
* [minikube config get](minikube_config_get.md)	 - Gets the value of PROPERTY_NAME from the minikube config file
* [minikube config set](minikube_config_set.md)	 - Sets an individual value in a minikube config file
* [minikube config unset](minikube_config_unset.md)	 - unsets an individual value in a minikube config file

