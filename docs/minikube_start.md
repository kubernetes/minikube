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
      --apiserver-name string           The apiserver name which is used in the generated certificate for localkube/kubernetes.  This can be used if you want to make the apiserver available from outside the machine (default "minikubeCA")
      --container-runtime string        The container runtime to be used
      --cpus int                        Number of CPUs allocated to the minikube VM (default 2)
      --disk-size string                Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g) (default "20g")
      --docker-env stringArray          Environment variables to pass to the Docker daemon. (format: key=value)
      --extra-config ExtraOption        A set of key=value pairs that describe configuration that may be passed to different components.
		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
		Valid components are: kubelet, apiserver, controller-manager, etcd, proxy, scheduler.
      --feature-gates string            A set of key=value pairs that describe feature gates for alpha/experimental features.
      --host-only-cidr string           The CIDR to be used for the minikube VM (only supported with Virtualbox driver) (default "192.168.99.1/24")
      --hyperv-virtual-switch string    The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)
      --insecure-registry stringSlice   Insecure Docker registries to pass to the Docker daemon
      --iso-url string                  Location of the minikube iso (default "https://storage.googleapis.com/minikube/iso/minikube-v1.0.7.iso")
      --keep-context                    This will keep the existing kubectl context and will create a minikube context.
      --kubernetes-version string       The kubernetes version that the minikube VM will use (ex: v1.2.3) 
 OR a URI which contains a localkube binary (ex: https://storage.googleapis.com/minikube/k8sReleases/v1.3.0/localkube-linux-amd64) (default "v1.6.0-beta.2")
      --kvm-network string              The KVM network name. (only supported with KVM driver) (default "default")
      --memory int                      Amount of RAM allocated to the minikube VM (default 2048)
      --network-plugin string           The name of the network plugin
      --registry-mirror stringSlice     Registry mirrors to pass to the Docker daemon
      --vm-driver string                VM driver is one of: [virtualbox vmwarefusion kvm xhyve hyperv] (default "virtualbox")
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

