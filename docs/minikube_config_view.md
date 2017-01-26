## minikube config view

Display values currently set in the minikube config file

### Synopsis


Display values currently set in the minikube config file.

```
minikube config view
```

### Options

```
      --format string   Go template format string for the config view output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
For the list of accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd/config#ConfigViewTemplate (default "- {{.ConfigKey}}: {{.ConfigValue}}
")
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
* [minikube config](minikube_config.md)	 - Modify minikube config

