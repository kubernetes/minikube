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
      --auto-pause-interval duration      Duration of inactivity before the minikube VM is paused (default 1m0s) (default 1m0s)
      --auto-update-drivers               If set, automatically updates drivers to the latest version. Defaults to true. (default true)
      --base-image string                 The base image to use for docker/podman drivers. Intended for local development. (default "gcr.io/k8s-minikube/kicbase-builds:v0.0.45-1730282848-19883@sha256:e762c909ad2a507083ec25b1ad3091c71fc7d92824e4a659c9158bbfe5ae03d4")
      --binary-mirror string              Location to fetch kubectl, kubelet, & kubeadm binaries from.
      --cache-images                      If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --driver=none. (default true)
      --cert-expiration duration          Duration until minikube certificate expiration, defaults to three years (26280h). (default 26280h0m0s)
      --cni string                        CNI plug-in to use. Valid options: auto, bridge, calico, cilium, flannel, kindnet, or path to a CNI manifest (default: auto)
  -c, --container-runtime string          The container runtime to be used. Valid options: docker, cri-o, containerd (default: auto)
      --cpus string                       Number of CPUs allocated to Kubernetes. Use "max" to use the maximum number of CPUs. Use "no-limit" to not specify a limit (Docker/Podman only) (default "2")
      --cri-socket string                 The cri socket path to be used.
      --delete-on-failure                 If set, delete the current cluster if start fails and try again. Defaults to false.
      --disable-driver-mounts             Disables the filesystem mounts provided by the hypervisors
      --disable-metrics                   If set, disables metrics reporting (CPU and memory usage), this can improve CPU usage. Defaults to false.
      --disable-optimizations             If set, disables optimizations that are set for local Kubernetes. Including decreasing CoreDNS replicas from 2 to 1. Defaults to false.
      --disk-size string                  Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g). (default "20000mb")
      --dns-domain string                 The cluster dns domain name used in the Kubernetes cluster (default "cluster.local")
      --dns-proxy                         Enable proxy for NAT DNS requests (virtualbox driver only)
      --docker-env stringArray            Environment variables to pass to the Docker daemon. (format: key=value)
      --docker-opt stringArray            Specify arbitrary flags to pass to the Docker daemon. (format: key=value)
      --download-only                     If true, only download and cache files for later use - don't install or start anything.
  -d, --driver string                     Used to specify the driver to run Kubernetes in. The list of available drivers depends on operating system.
      --dry-run                           dry-run mode. Validates configuration, but does not mutate system state
      --embed-certs                       if true, will embed the certs in kubeconfig.
      --enable-default-cni                DEPRECATED: Replaced by --cni=bridge
      --extra-config ExtraOption          A set of key=value pairs that describe configuration that may be passed to different components.
                                          		The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
                                          		Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
                                          		Valid kubeadm parameters: ignore-preflight-errors, dry-run, kubeconfig, kubeconfig-dir, node-name, cri-socket, experimental-upload-certs, certificate-key, rootfs, skip-phases, pod-network-cidr
      --extra-disks int                   Number of extra disks created and attached to the minikube VM (currently only implemented for hyperkit, kvm2, and qemu2 drivers)
      --feature-gates string              A set of key=value pairs that describe feature gates for alpha/experimental features.
      --force                             Force minikube to perform possibly dangerous operations
      --force-systemd                     If set, force the container runtime to use systemd as cgroup manager. Defaults to false.
  -g, --gpus string                       Allow pods to use your GPUs. Options include: [all,nvidia,amd] (Docker driver with Docker container-runtime only)
      --ha                                Create Highly Available Multi-Control Plane Cluster with a minimum of three control-plane nodes that will also be marked for work.
      --host-dns-resolver                 Enable host resolver for NAT DNS requests (virtualbox driver only) (default true)
      --host-only-cidr string             The CIDR to be used for the minikube VM (virtualbox driver only) (default "192.168.59.1/24")
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
      --iso-url strings                   Locations to fetch the minikube ISO from. The list depends on the machine architecture.
      --keep-context                      This will keep the existing kubectl context and will create a minikube context.
      --kubernetes-version string         The Kubernetes version that the minikube VM will use (ex: v1.2.3, 'stable' for v1.31.2, 'latest' for v1.31.2). Defaults to 'stable'.
      --kvm-gpu                           Enable experimental NVIDIA GPU support in minikube
      --kvm-hidden                        Hide the hypervisor signature from the guest in minikube (kvm2 driver only)
      --kvm-network string                The KVM default network name. (kvm2 driver only) (default "default")
      --kvm-numa-count int                Simulate numa node count in minikube, supported numa node count range is 1-8 (kvm2 driver only) (default 1)
      --kvm-qemu-uri string               The KVM QEMU connection URI. (kvm2 driver only) (default "qemu:///system")
      --listen-address string             IP Address to use to expose ports (docker and podman driver only)
      --memory string                     Amount of RAM to allocate to Kubernetes (format: <number>[<unit>], where unit = b, k, m or g). Use "max" to use the maximum amount of memory. Use "no-limit" to not specify a limit (Docker/Podman only)
      --mount                             This will start the mount daemon and automatically mount files into minikube.
      --mount-9p-version string           Specify the 9p version that the mount should use (default "9p2000.L")
      --mount-gid string                  Default group id used for the mount (default "docker")
      --mount-ip string                   Specify the ip that the mount should be setup on
      --mount-msize int                   The number of bytes to use for 9p packet payload (default 262144)
      --mount-options strings             Additional mount options, such as cache=fscache
      --mount-port uint16                 Specify the port that the mount should be setup on, where 0 means any free port.
      --mount-string string               The argument to pass the minikube mount command on start.
      --mount-type string                 Specify the mount filesystem type (supported types: 9p) (default "9p")
      --mount-uid string                  Default user id used for the mount (default "docker")
      --namespace string                  The named space to activate after start (default "default")
      --nat-nic-type string               NIC Type used for nat network. One of Am79C970A, Am79C973, 82540EM, 82543GC, 82545EM, or virtio (virtualbox driver only) (default "virtio")
      --native-ssh                        Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'. (default true)
      --network string                    network to run minikube with. Now it is used by docker/podman and KVM drivers. If left empty, minikube will create a new network.
      --network-plugin string             DEPRECATED: Replaced by --cni
      --nfs-share strings                 Local folders to share with Guest via NFS mounts (hyperkit driver only)
      --nfs-shares-root string            Where to root the NFS Shares, defaults to /nfsshares (hyperkit driver only) (default "/nfsshares")
      --no-kubernetes                     If set, minikube VM/container will start without starting or configuring Kubernetes. (only works on new clusters)
      --no-vtx-check                      Disable checking for the availability of hardware virtualization before the vm is started (virtualbox driver only)
  -n, --nodes int                         The total number of nodes to spin up. Defaults to 1. (default 1)
  -o, --output string                     Format to print stdout in. Options include: [text,json] (default "text")
      --ports strings                     List of ports that should be exposed (docker and podman driver only)
      --preload                           If set, download tarball of preloaded images if available to improve start time. Defaults to true. (default true)
      --qemu-firmware-path string         Path to the qemu firmware file. Defaults: For Linux, the default firmware location. For macOS, the brew installation location. For Windows, C:\Program Files\qemu\share
      --registry-mirror strings           Registry mirrors to pass to the Docker daemon
      --service-cluster-ip-range string   The CIDR to be used for service cluster IPs. (default "10.96.0.0/12")
      --socket-vmnet-client-path string   Path to the socket vmnet client binary (QEMU driver only)
      --socket-vmnet-path string          Path to socket vmnet binary (QEMU driver only)
      --ssh-ip-address string             IP address (ssh driver only)
      --ssh-key string                    SSH key (ssh driver only)
      --ssh-port int                      SSH port (ssh driver only) (default 22)
      --ssh-user string                   SSH user (ssh driver only) (default "root")
      --static-ip string                  Set a static IP for the minikube cluster, the IP must be: private, IPv4, and the last octet must be between 2 and 254, for example 192.168.200.200 (Docker and Podman drivers only)
      --subnet string                     Subnet to be used on kic cluster. If left empty, minikube will choose subnet address, beginning from 192.168.49.0. (docker and podman driver only)
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
      --alsologtostderr                  log to standard error as well as files (no effect when -logtostderr=true)
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the Kubernetes cluster. (default "kubeadm")
  -h, --help                             
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory (no effect when -logtostderr=true)
      --log_file string                  If non-empty, use this log file (no effect when -logtostderr=true)
      --log_file_max_size uint           Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                      log to standard error instead of files
      --one_output                       If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true)
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --rootless                         Force to use rootless driver (docker and podman driver only)
      --skip-audit                       Skip recording the current command in the audit logs.
      --skip_headers                     If true, avoid header prefixes in the log messages
      --skip_log_headers                 If true, avoid headers when opening log files (no effect when -logtostderr=true)
      --stderrthreshold severity         logs at or above this threshold go to stderr when writing to files and stderr (no effect when -logtostderr=true or -alsologtostderr=true) (default 2)
      --user string                      Specifies the user executing the operation. Useful for auditing operations executed by 3rd party tools. Defaults to the operating system username.
  -v, --v Level                          number for the log level verbosity
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

