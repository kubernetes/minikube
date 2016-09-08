## minikube service

Gets the kubernetes URL for the specified service in your local cluster

### Synopsis


Gets the kubernetes URL for the specified service in your local cluster

```
minikube service [flags] SERVICE
```

### Options

```
  -n, --namespace string   The service namespace (default "default")
      --url                Display the kubernetes service URL in the CLI instead of opening it in the default browser
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

