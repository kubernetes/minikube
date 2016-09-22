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
      --container-runtime string    The container runtime to be used
      --cpus int                    Number of CPUs allocated to the minikube VM (default 1)
      --disk-size string            Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g) (default "20g")
      --docker-env value            Environment variables to pass to the Docker daemon. (format: key=value) (default [])
      --host-only-cidr string       The CIDR to be used for the minikube VM (only supported with Virtualbox driver) (default "192.168.99.1/24")
      --insecure-registry value     Insecure Docker registries to pass to the Docker daemon (default [])
      --iso-url string              Location of the minikube iso (default "https://storage.googleapis.com/minikube/minikube-0.7.iso")
      --kubernetes-version string   The kubernetes version that the minikube VM will (ex: v1.2.3) 
 OR a URI which contains a localkube binary (ex: https://storage.googleapis.com/minikube/k8sReleases/v1.3.0/localkube-linux-amd64) (default "v1.4.0-beta.10")
      --memory int                  Amount of RAM allocated to the minikube VM (default 1024)
      --network-plugin string       The name of the network plugin
      --registry-mirror value       Registry mirrors to pass to the Docker daemon (default [])
      --vm-driver string            VM driver is one of: [virtualbox vmwarefusion kvm xhyve hyperv] (default "virtualbox")
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

