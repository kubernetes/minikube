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
      --disk-size="20g": Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g)
      --docker-env=[]: Environment variables to pass to the Docker daemon. (format: key=value)
      --host-only-cidr="192.168.99.1/24": The CIDR to be used for the minikube VM (only supported with Virtualbox driver)
      --insecure-registry=[]: Insecure Docker registries to pass to the Docker daemon
      --iso-url="https://storage.googleapis.com/minikube/minikube-0.6.iso": Location of the minikube iso
      --kubernetes-version="v1.3.5": The kubernetes version that the minikube VM will (ex: v1.2.3) 
 OR a URI which contains a localkube binary (ex: https://storage.googleapis.com/minikube/k8sReleases/v1.3.0/localkube-linux-amd64)
      --memory=1024: Amount of RAM allocated to the minikube VM
      --registry-mirror=[]: Registry mirrors to pass to the Docker daemon
      --vm-driver="virtualbox": VM driver is one of: [virtualbox vmwarefusion kvm xhyve hyperv]
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

