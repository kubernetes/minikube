## minikube start

Starts a local kubernetes cluster.

### Synopsis


Starts a local kubernetes cluster using Virtualbox. This command
assumes you already have Virtualbox installed.

```
minikube start
```

### Options

```
      --cpus=1: Number of CPUs allocated to the minikube VM
      --iso-url="https://storage.googleapis.com/minikube/minikube-0.4.iso": Location of the minikube iso
      --memory=1024: Amount of RAM allocated to the minikube VM
      --vm-driver="virtualbox": VM driver is one of: [virtualbox vmwarefusion]
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

