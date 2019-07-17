# minikube CLI Commands
This document serves as a reference to all the commands, flags and their accepted arguments

## Global Flags
These flags can be used globally with any command on the CLI. Following are the global flags -
```
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the kubernetes cluster. (default "kubeadm")
  -h, --help                             help for minikube
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used.
                                                This can be modified to allow for multiple minikube instances to be run independently (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

## Commands
In this section, all commands which are accepted by the `minikube` CLI are described. To get help about any command, you can also type in `minikube help <command>`

---
 ### addons
**Description -** Modifies minikube addons files using subcommands like `minikube addons enable heapster`
**Usage -** 
```
minikube addons SUBCOMMAND [flags]
minikube addons [command]
```
**Available Subcommands -**
```
configure   Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list
disable     Disables the addon w/ADDON_NAME within minikube (example: minikube addons disable dashboard). For a list of available addons use: minikube addons list
enable      Enables the addon w/ADDON_NAME within minikube (example: minikube addons enable dashboard). For a list of available addons use: minikube addons list
list        Lists all available minikube addons as well as their current statuses (enabled/disabled)
open        Opens the addon w/ADDON_NAME within minikube (example: minikube addons open dashboard). For a list of available addons use: minikube addons list
```

---
### cache
**Description -** Add or delete an image from the local cache.
**Usage -** `minikube cache [command]`
**Available Subcommands-**
```
add         Add an image to local cache.
delete      Delete an image from the local cache.
list        List all available images from the local cache.
```

---
### completion
**Description -** 

> Outputs minikube shell completion for the given shell (bash or zsh)
> 
>         This depends on the bash-completion binary.  Example installation instructions:
>         OS X:
>                 $ brew install bash-completion
>                 $ source $(brew --prefix)/etc/bash_completion
>                 $ minikube completion bash > ~/.minikube-completion  # for bash users
>                 $ minikube completion zsh > ~/.minikube-completion  # for zsh users
>                 $ source ~/.minikube-completion
>         Ubuntu:
>                 $ apt-get install bash-completion
>                 $ source /etc/bash-completion
>                 $ source <(minikube completion bash) # for bash users
>                 $ source <(minikube completion zsh) # for zsh users
> 
>         Additionally, you may want to output the completion to a file and source in your .bashrc
> 
>         Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2
**Usage -** `minikube completion SHELL`

---
### config
**Description -** config modifies minikube config files using subcommands like `minikube config set vm-driver kvm`
Configurable fields:
 * vm-driver
 * feature-gates
 * v
 * cpus
 * disk-size
 * host-only-cidr
 * memory
 * log_dir
 * kubernetes-version
 * iso-url
 * WantUpdateNotification
 * ReminderWaitPeriodInHours
 * WantReportError
 * WantReportErrorPrompt
 * WantKubectlDownloadMsg
 * WantNoneDriverWarning
 * profile
 * bootstrapper
 * ShowDriverDeprecationNotification
 * ShowBootstrapperDeprecationNotification
 * dashboard
 * addon-manager
 * default-storageclass
 * heapster
 * efk
 * ingress
 * registry
 * registry-creds
 * freshpod
 * default-storageclass
 * storage-provisioner
 * storage-provisioner-gluster
 * metrics-server
 * nvidia-driver-installer
 * nvidia-gpu-device-plugin
 * logviewer
 * gvisor
 * hyperv-virtual-switch
 * disable-driver-mounts
 * cache
 * embed-certs

**Usage -**
```
minikube config SUBCOMMAND [flags]
minikube config [command]
```
**Available Subcommands-**
```
get         Gets the value of PROPERTY_NAME from the minikube config file
set         Sets an individual value in a minikube config file
unset       unsets an individual value in a minikube config file
view        Display values currently set in the minikube config file
```

---
### dashboard
**Description -** Access the kubernetes dashboard running within the minikube cluster
**Usage -** `minikube dashboard [flags]`
**Available Flags -**
```
-h, --help   help for dashboard
    --url    Display dashboard URL instead of opening a browser
```

---
### delete
**Description -** Deletes a local kubernetes cluster. This command deletes the VM, and removes all
associated files.
**Usage -** `minikube delete`

---
### docker-env
**Description -** Sets up docker env variables; similar to '$(docker-machine env)'.
**Usage -** `minikube docker-env [flags]`
**Available Flags -**
```
  -h, --help           help for docker-env
      --no-proxy       Add machine IP to NO_PROXY environment variable
      --shell string   Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect
  -u, --unset          Unset variables instead of setting them
```

---
### help
**Description -** Help provides help for any command in the application. Simply type minikube help [path to command] for full details.
**Usage -** `minikube help [command] [flags]`

---
### ip
**Description -** Retrieves the IP address of the running cluster, and writes it to STDOUT.
**Usage -** `minikube ip`

---
### kubectl
**Description -** Run the kubernetes client, download it if necessary.
**Usage -** `minikube kubectl`

---
### logs
**Description -** Gets the logs of the running instance, used for debugging minikube, not user code.
**Usage -** `minikube logs [flags]`
**Available Flags -**
```
  -f, --follow       Show only the most recent journal entries, and continuously print new entries as they are appended to the journal.
  -h, --help         help for logs
  -n, --length int   Number of lines back to go within the log (default 50)
      --problems     Show only log entries which point to known problems
```

---
### mount
**Description -** Mounts the specified directory into minikube.
**Usage -** `minikube mount [flags] <source directory>:<target directory>`
**Available Flags -**
```
--9p-version string   Specify the 9p version that the mount should use (default "9p2000.L")
      --gid string          Default group id used for the mount (default "docker")
  -h, --help                help for mount
      --ip string           Specify the ip that the mount should be setup on
      --kill                Kill the mount process spawned by minikube start
      --mode uint           File permissions used for the mount (default 493)
      --msize int           The number of bytes to use for 9p packet payload (default 262144)
      --options strings     Additional mount options, such as cache=fscache
      --type string         Specify the mount filesystem type (supported types: 9p) (default "9p")
      --uid string          Default user id used for the mount (default "docker")
```

---
### profile
**Description -** Sets the current minikube profile, or gets the current profile if no arguments are provided.  This is used to run and manage multiple minikube instance.  You can return to the default minikube profile by running `minikube profile default`
**Usage -** 
```
minikube profile [MINIKUBE_PROFILE_NAME].  You can return to the default minikube profile by running `minikube profile default` [flags]
```

---
### service
**Description -** Gets the kubernetes URL(s) for the specified service in your local cluster. In the case of multiple URLs they will be printed one at a time.
**Usage -** 
```
minikube service [flags] SERVICE
minikube service [command]
```
**Available Commands -**
```
  list        Lists the URLs for the services in your local cluster
```
**Available Flags -**
```
      --format string      Format to output service URL in. This format will be applied to each url individually and they will be printed one at a time. (default "http://{{.IP}}:{{.Port}}")
  -h, --help               help for service
      --https              Open the service URL with https instead of http
      --interval int       The time interval for each check that wait performs in seconds (default 20)
  -n, --namespace string   The service namespace (default "default")
      --url                Display the kubernetes service URL in the CLI instead of opening it in the default browser
      --wait int           Amount of time to wait for a service in seconds (default 20)
```

---
### ssh
**Description -** Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'.
**Usage -** `minikube ssh`

---
### ssh-key
**Description -** Retrieve the ssh identity key path of the specified cluster.
**Usage -** `minikube ssh-key`

---
### start
**Description -** Starts a local kubernetes cluster.
**Usage -** `minikube start [flags]`
**Available Flags -**
```
      --apiserver-ips ipSlice             A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine (default [])
      --apiserver-name string             The apiserver name which is used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine (default "minikubeCA")
      --apiserver-names stringArray       A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine
      --apiserver-port int                The apiserver listening port (default 8443)
      --cache-images                      If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --vm-driver=none. (default true)
      --container-runtime string          The container runtime to be used (docker, crio, containerd) (default "docker")
      --cpus int                          Number of CPUs allocated to the minikube VM (default 2)
      --cri-socket string                 The cri socket path to be used
      --disable-driver-mounts             Disables the filesystem mounts provided by the hypervisors (vboxfs)
      --disk-size string                  Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g) (default "20000mb")
      --dns-domain string                 The cluster dns domain name used in the kubernetes cluster (default "cluster.local")
      --docker-env stringArray            Environment variables to pass to the Docker daemon. (format: key=value)
      --docker-opt stringArray            Specify arbitrary flags to pass to the Docker daemon. (format: key=value)
      --download-only                     If true, only download and cache files for later use - don't install or start anything.
      --enable-default-cni                Enable the default CNI plugin (/etc/cni/net.d/k8s.conf). Used in conjunction with "--network-plugin=cni"
      --extra-config ExtraOption          A set of key=value pairs that describe configuration that may be passed to different components.
                                                        The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
                                                        Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
                                                        Valid kubeadm parameters: ignore-preflight-errors, dry-run, kubeconfig, kubeconfig-dir, node-name, cri-socket, experimental-upload-certs, certificate-key, rootfs, pod-network-cidr
      --feature-gates string              A set of key=value pairs that describe feature gates for alpha/experimental features.
      --gpu                               Enable experimental NVIDIA GPU support in minikube (works only with kvm2 driver on Linux)
  -h, --help                              help for start
      --hidden                            Hide the hypervisor signature from the guest in minikube (works only with kvm2 driver on Linux)
      --host-only-cidr string             The CIDR to be used for the minikube VM (only supported with Virtualbox driver) (default "192.168.99.1/24")
      --hyperkit-vpnkit-sock string       Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock.
      --hyperkit-vsock-ports strings      List of guest VSock ports that should be exposed as sockets on the host (Only supported on with hyperkit now).
      --hyperv-virtual-switch string      The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)
      --image-mirror-country string       Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn
      --image-repository string           Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to "auto" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers
      --insecure-registry strings         Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.
      --iso-url string                    Location of the minikube iso (default "https://storage.googleapis.com/minikube/iso/minikube-v1.2.0.iso")
      --keep-context                      This will keep the existing kubectl context and will create a minikube context.
      --kubernetes-version string         The kubernetes version that the minikube VM will use (ex: v1.2.3) (default "v1.15.0")
      --kvm-network string                The KVM network name. (only supported with KVM driver) (default "default")
      --memory string                     Amount of RAM allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g) (default "2000mb")
      --mount                             This will start the mount daemon and automatically mount files into minikube
      --mount-string string               The argument to pass the minikube mount command on start (default "C:\\Users\\Pranav.Jituri:/minikube-host")
      --network-plugin string             The name of the network plugin
      --nfs-share strings                 Local folders to share with Guest via NFS mounts (Only supported on with hyperkit now)
      --nfs-shares-root string            Where to root the NFS Shares (defaults to /nfsshares, only supported with hyperkit now) (default "/nfsshares")
      --no-vtx-check                      Disable checking for the availability of hardware virtualization before the vm is started (virtualbox)
      --registry-mirror strings           Registry mirrors to pass to the Docker daemon
      --service-cluster-ip-range string   The CIDR to be used for service cluster IPs. (default "10.96.0.0/12")
      --uuid string                       Provide VM UUID to restore MAC address (only supported with Hyperkit driver).
      --vm-driver string                  VM driver is one of: [virtualbox parallels vmwarefusion kvm hyperv hyperkit kvm2 vmware none] (default "virtualbox")
```

---
### status
**Description -** Gets the status of a local kubernetes cluster. Exit status contains the status of minikube's VM, cluster and kubernetes encoded on it's bits in this order from right to left.
	        Eg: 7 meaning: 1 (for minikube NOK) + 2 (for cluster NOK) + 4 (for kubernetes NOK)
**Usage -** `minikube status [flags]`
**Available Flags -**
```
      --format string   Go template format string for the status output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
                        For the list accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#Status (default "host: {{.Host}}\nkubelet: {{.Kubelet}}\napiserver: {{.APIServer}}\nkubectl: {{.Kubeconfig}}\n")
```

---
### stop
**Description -** Stops a local kubernetes cluster running in Virtualbox. This command stops the VM
itself, leaving all files intact. The cluster can be started again with the `start` command.
**Usage -** `minikube stop`

---
### tunnel
**Description -** Creates a route to services deployed with type LoadBalancer and sets their Ingress to their ClusterIP
**Usage -** `minikube tunnel [flags]`
**Available Flags -**
```
  -c, --cleanup   call with cleanup=true to remove old tunnels
```

---
### update-check
**Description -** Print current and latest version number.
**Usage -** `minikube update-check`

---
### update-context
**Description -** Retrieves the IP address of the running cluster, checks it with IP in kubeconfig, and corrects kubeconfig if incorrect.
**Usage -** `minikube update-context`

---
### version
**Description -** Print the version of minikube.
**Usage -** `minikube version`
