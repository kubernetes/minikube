## minikube service list

Lists the URLs for the services in your local cluster

### Synopsis


Lists the URLs for the services in your local cluster

```
minikube service list [flags]
```

### Options

```
  -n, --namespace string   The services namespace
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --format string                    Format to output service URL in.  This format will be applied to each url individually and they will be printed one at a time. (default "http://{{.IP}}:{{.Port}}")
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
* [minikube service](minikube_service.md)	 - Gets the kubernetes URL(s) for the specified service in your local cluster

