---
title: "start"
description: >
  Starts a local Kubernetes cluster
---


## minikube start

Starts a local Kubernetes cluster

### Synopsis

Starts a local Kubernetes cluster

```shell
minikube start [flags]
```

### Options

```
      --addons minikube addons list       Enable addons. see minikube addons list for a list of valid addon names.
      --apiserver-ips ipSlice             A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine (default [])
      --apiserver-name string             The authoritative apiserver hostname for apiserver certificates and connectivity. This can be used if you want to make the apiserver available from outside the machine (default "minikubeCA")
      --apiserver-names strings           A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine
      --apiserver-port int                The apiserver listening port (default 8443)
      --auto-update-drivers               If set, automatically updates drivers to the latest version. Defaults to true. (default true)
      --base-image string                 The base image to use for docker/podman drivers. Intended for local development. (default "gcr.io/k8s-minikube/kicbase-multiarch:v0.0.15-snapshot4@sha256:todo")
      --cache-images                      If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --driver=none. (default true)
      --cni string                        CNI plug-in to use. Valid options: auto, bridge, calico, cilium, flannel, kindnet, or path to a CNI manifest (default: auto)
      --container-runtime string          The container runtime to be used (docker, cri-o, containerd). (default "docker")
      --cpus int                          Number of CPUs allocated to Kubernetes. (default 2)
      --cri-socket string                 The cri socket path to be used.
      --delete-on-failure                 If set, delete the current cluster if start fails and try again. Defaults to false.
      --disable-driver-mounts             Disables the filesystem mounts provided by the hypervisors
      --disk-size string                  Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g). (default "20000mb")
      --dns-domain string                 The cluster dns domain name used in the Kubernetes cluster (default "cluster.local")
      --dns-proxy                         Enable proxy for NAT DNS requests (virtualbox driver only)
      --docker-env stringArray            Environment variables to pass to the Docker daemon. (format: key=value)
      --docker-opt stringArray            Specify arbitrary flags to pass to the Docker daemon. (format: key=value)
      --download-only                     If true, only download and cache files for later use - don't install or start anything.
      --driver string                     Used to specify the driver to run Kubernetes in. The list of available drivers depends on operating system.
      --dry-run                           dry-run mode. Validates configuration, but does not mutate system state
      --embed-certs                       if true, will embed the certs in kubeconfig.
      --enable-default-cni                DEPRECATED: Replaced by --cni=bridge
      --extra-config ExtraOption          A set of key=value pairs that describe configuration that may be passed to different components.
                                          		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
                                          		Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
                                          		Valid kubeadm parameters: ignore-preflight-errors, dry-run, kubeconfig, kubeconfig-dir, node-name, cri-socket, experimental-upload-certs, certificate-key, rootfs, skip-phases, pod-network-cidr
      --feature-gates string              A set of key=value pairs that describe feature gates for alpha/experimental features.
      --force                             Force minikube to perform possibly dangerous operations
      --force-systemd                     If set, force the container runtime to use sytemd as cgroup manager. Currently available for docker and crio. Defaults to false.
      --host-dns-resolver                 Enable host resolver for NAT DNS requests (virtualbox driver only) (default true)
      --host-only-cidr string             The CIDR to be used for the minikube VM (virtualbox driver only) (default "192.168.99.1/24")
      --host-only-nic-type string         NIC Type used for host only network. One of Am79C970A, Am79C973, 82540EM, 82543GC, 82545EM, or virtio (virtualbox driver only) (default "virtio")
      --hyperkit-vpnkit-sock string       Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock (hyperkit driver only)
      --hyperkit-vsock-ports strings      List of guest VSock ports that should be exposed as sockets on the host (hyperkit driver only)
      --hyperv-external-adapter string    External Adapter on which external switch will be created if no external switch is found. (hyperv driver only)
      --hyperv-use-external-switch        Whether to use external switch over Default Switch if virtual switch not explicitly specified. (hyperv driver only)
      --hyperv-virtual-switch string      The hyperv virtual switch name. Defaults to first found. (hyperv driver only)
      --image-mirror-country string       Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn.
      --image-repository string           Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to "auto" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers
      --insecure-registry strings         Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.
      --install-addons                    If set, install addons. Defaults to true. (default true)
      --interactive                       Allow user prompts for more information (default true)
      --iso-url strings                   Locations to fetch the minikube ISO from. (default [https://storage.googleapis.com/minikube/iso/minikube-v1.16.0.iso,https://github.com/kubernetes/minikube/releases/download/v1.16.0/minikube-v1.16.0.iso,https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/iso/minikube-v1.16.0.iso])
      --keep-context                      This will keep the existing kubectl context and will create a minikube context.
      --kubernetes-version string         The Kubernetes version that the minikube VM will use (ex: v1.2.3, 'stable' for v1.20.0, 'latest' for v1.20.0). Defaults to 'stable'.
      --kvm-gpu                           Enable experimental NVIDIA GPU support in minikube
      --kvm-hidden                        Hide the hypervisor signature from the guest in minikube (kvm2 driver only)
      --kvm-network string                The KVM network name. (kvm2 driver only) (default "default")
      --kvm-qemu-uri string               The KVM QEMU connection URI. (kvm2 driver only) (default "qemu:///system")
      --memory string                     Amount of RAM to allocate to Kubernetes (format: <number>[<unit>], where unit = b, k, m or g).
      --mount                             This will start the mount daemon and automatically mount files into minikube.
      --mount-string string               The argument to pass the minikube mount command on start.
      --namespace string                  The named space to activate after start (default "default")
      --nat-nic-type string               NIC Type used for nat network. One of Am79C970A, Am79C973, 82540EM, 82543GC, 82545EM, or virtio (virtualbox driver only) (default "virtio")
      --native-ssh                        Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'. (default true)
      --network string                    network to run minikube with. Only available with the docker/podman drivers. If left empty, minikube will create a new network.
      --network-plugin string             Kubelet network plug-in to use (default: auto)
      --nfs-share strings                 Local folders to share with Guest via NFS mounts (hyperkit driver only)
      --nfs-shares-root string            Where to root the NFS Shares, defaults to /nfsshares (hyperkit driver only) (default "/nfsshares")
      --no-vtx-check                      Disable checking for the availability of hardware virtualization before the vm is started (virtualbox driver only)
  -n, --nodes int                         The number of nodes to spin up. Defaults to 1. (default 1)
  -o, --output string                     Format to print stdout in. Options include: [text,json] (default "text")
      --ports strings                     List of ports that should be exposed (docker and podman driver only)
      --preload                           If set, download tarball of preloaded images if available to improve start time. Defaults to true. (default true)
      --registry-mirror strings           Registry mirrors to pass to the Docker daemon
      --service-cluster-ip-range string   The CIDR to be used for service cluster IPs. (default "10.96.0.0/12")
      --trace string                      Send trace events. Options include: [gcp]
      --uuid string                       Provide VM UUID to restore MAC address (hyperkit driver only)
      --vm                                Filter to use only VM Drivers
      --vm-driver driver                  DEPRECATED, use driver instead.
      --wait strings                      comma separated list of Kubernetes components to verify and wait for after starting a cluster. defaults to "apiserver,system_pods", available options: "apiserver,system_pods,default_sa,apps_running,node_ready,kubelet" . other acceptable values are 'all' or 'none', 'true' and 'false' (default [apiserver,system_pods])
      --wait-timeout duration             max time to wait per Kubernetes or host to be healthy. (default 6m0s)
```

### Options inherited from parent commands

```
      --add_dir_header                   If true, adds the file directory to the header of the log messages
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the Kubernetes cluster. (default "kubeadm")
  -h, --help                             
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --log_file string                  If non-empty, use this log file
      --log_file_max_size uint           Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                      log to standard error instead of files
      --one_output                       If true, only write logs to their native severity level (vs also writing to each lower severity level
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --skip_headers                     If true, avoid header prefixes in the log messages
      --skip_log_headers                 If true, avoid headers when opening log files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          number for the log level verbosity
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

