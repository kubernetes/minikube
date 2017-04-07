## minikube addons

Modify minikube's kubernetes addons

### Synopsis


addons modifies minikube addons files using subcommands like "minikube addons enable heapster"

```
minikube addons SUBCOMMAND [flags]
```

### Options

```
      --format string   Go template format string for the addon list output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
For the list of accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd/config#AddonListTemplate (default "- {{.AddonName}}: {{.AddonStatus}}
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
* [minikube](minikube.md)	 - Minikube is a tool for managing local Kubernetes clusters.
* [minikube addons configure](minikube_addons_configure.md)	 - Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list 
* [minikube addons disable](minikube_addons_disable.md)	 - Disables the addon w/ADDON_NAME within minikube (example: minikube addons disable dashboard). For a list of available addons use: minikube addons list 
* [minikube addons enable](minikube_addons_enable.md)	 - Enables the addon w/ADDON_NAME within minikube (example: minikube addons enable dashboard). For a list of available addons use: minikube addons list 
* [minikube addons list](minikube_addons_list.md)	 - Lists all available minikube addons as well as there current status (enabled/disabled)
* [minikube addons open](minikube_addons_open.md)	 - Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list 

