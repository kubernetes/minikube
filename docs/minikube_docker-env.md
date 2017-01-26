## minikube docker-env

sets up docker env variables; similar to '$(docker-machine env)'

### Synopsis


sets up docker env variables; similar to '$(docker-machine env)'

```
minikube docker-env
```

### Options

```
      --no-proxy       Add machine IP to NO_PROXY environment variable
      --shell string   Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect
  -u, --unset          Unset variables instead of setting them
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

