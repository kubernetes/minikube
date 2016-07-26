## minikube docker-env

sets up docker env variables; similar to '$(docker-machine env)'

### Synopsis


sets up docker env variables; similar to '$(docker-machine env)'

```
minikube docker-env
```

### Options

```
      --no-proxy[=false]: Add machine IP to NO_PROXY environment variable
      --shell="": Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect
  -u, --unset[=false]: Unset variables instead of setting them
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

