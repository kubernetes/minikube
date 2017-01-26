## minikube addons open

Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list 

### Synopsis


Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list 

```
minikube addons open ADDON_NAME
```

### Options

```
      --format string   Format to output addons URL in.  This format will be applied to each url individually and they will be printed one at a time. (default "http://{{.IP}}:{{.Port}}")
      --https           Open the addons URL with https instead of http
      --url             Display the kubernetes addons URL in the CLI instead of opening it in the default browser
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
* [minikube addons](minikube_addons.md)	 - Modify minikube's kubernetes addons

