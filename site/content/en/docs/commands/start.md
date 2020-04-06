---
title: "start"
linkTitle: "start"
weight: 1
date: 2019-08-01
description: >
  Starts a local Kubernetes cluster
---

### Usage

```
minikube start [flags]
```

### Options

```
    --addons minikube addons list       Enable addons. see minikube addons list for a list of valid addon names.
      --apiserver-ips ipSlice             A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine (default [])
      --apiserver-name string             The apiserver name which is used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine (default "minikubeCA")
      --apiserver-names stringArray       A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine
      --apiserver-port int                The apiserver listening port (default 8443)
      --auto-update-drivers               If set, automatically updates drivers to the latest version. Defaults to true. (default true)
      --cache-images                      If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --vm-driver=none. (default true)
      --container-runtime string          The container runtime to be used (docker, crio, containerd). (default "docker")
      --cpus int                          Number of CPUs allocated to the minikube VM. (default 2)
      --cri-socket string                 The cri socket path to be used.
      --disable-driver-mounts             Disables the filesystem mounts provided by the hypervisors
      --disk-size string                  Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g). (default "20000mb")
      --dns-domain string                 The cluster dns domain name used in the kubernetes cluster (default "cluster.local")
      --dns-proxy                         Enable proxy for NAT DNS requests (virtualbox driver only)
      --docker-env stringArray            Environment variables to pass to the Docker daemon. (format: key=value)
      --docker-opt stringArray            Specify arbitrary flags to pass to the Docker daemon. (format: key=value)
      --download-only                     If true, only download and cache files for later use - don't install or start anything.
      --dry-run                           dry-run mode. Validates configuration, but does not mutate system state
      --embed-certs                       if true, will embed the certs in kubeconfig.
      --enable-default-cni                Enable the default CNI plugin (/etc/cni/net.d/k8s.conf). Used in conjunction with "--network-plugin=cni".
      --extra-config ExtraOption          A set of key=value pairs that describe configuration that may be passed to different components.

     The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
     
     Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
     
     Valid kubeadm parameters: ignore-preflight-errors, dry-run, kubeconfig, kubeconfig-dir, node-name, cri-socket, experimental-upload-certs, certificate-key, rootfs, skip-phases, pod-network-cidr
     
      --feature-gates string              A set of key=value pairs that describe feature gates for alpha/experimental features.
      --force                             Force minikube to perform possibly dangerous operations
  -h, --help                              help for start
      --host-dns-resolver                 Enable host resolver for NAT DNS requests (virtualbox driver only) (default true)
      --host-only-cidr string             The CIDR to be used for the minikube VM (virtualbox driver only) (default "192.168.99.1/24")
      --host-only-nic-type string         NIC Type used for host only network. One of Am79C970A, Am79C973, 82540EM, 82543GC, 82545EM, or virtio (virtualbox driver only) (default "virtio")
      --hyperkit-vpnkit-sock string       Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock (hyperkit driver only)
      --hyperkit-vsock-ports strings      List of guest VSock ports that should be exposed as sockets on the host (hyperkit driver only)
      --hyperv-virtual-switch string      The hyperv virtual switch name. Defaults to first found. (hyperv driver only)
      --image-mirror-country string       Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn.
      --image-repository string           Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to "auto" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers
      --insecure-registry strings         Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.
      --interactive                       Allow user prompts for more information (default true)
      --iso-url string                    Location of the minikube iso. (default "https://storage.googleapis.com/minikube/iso/minikube-v1.7.0.iso")
      --keep-context                      This will keep the existing kubectl context and will create a minikube context.
      --kubernetes-version string         The kubernetes version that the minikube VM will use (ex: v1.2.3)
      --kvm-gpu                           Enable experimental NVIDIA GPU support in minikube
      --kvm-hidden                        Hide the hypervisor signature from the guest in minikube (kvm2 driver only)
      --kvm-network string                The KVM network name. (kvm2 driver only) (default "default")
      --kvm-qemu-uri string               The KVM QEMU connection URI. (kvm2 driver only) (default "qemu:///system")
      --memory string                     Amount of RAM allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g). (default "2000mb")
      --mount                             This will start the mount daemon and automatically mount files into minikube.
      --mount-string string               The argument to pass the minikube mount command on start. (default "/Users:/minikube-host")
      --nat-nic-type string               NIC Type used for host only network. One of Am79C970A, Am79C973, 82540EM, 82543GC, 82545EM, or virtio (virtualbox driver only) (default "virtio")
      --native-ssh                        Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'. (default true)
      --network-plugin string             The name of the network plugin.
      --nfs-share strings                 Local folders to share with Guest via NFS mounts (hyperkit driver only)
      --nfs-shares-root string            Where to root the NFS Shares, defaults to /nfsshares (hyperkit driver only) (default "/nfsshares")
      --no-vtx-check                      Disable checking for the availability of hardware virtualization before the vm is started (virtualbox driver only)
      --registry-mirror strings           Registry mirrors to pass to the Docker daemon
      --service-cluster-ip-range string   The CIDR to be used for service cluster IPs. (default "10.96.0.0/12")
      --uuid string                       Provide VM UUID to restore MAC address (hyperkit driver only)
      --vm-driver string                  Driver is one of: virtualbox, parallels, vmwarefusion, hyperkit, vmware, docker (experimental) (defaults to auto-detect)
      --wait                              Block until the apiserver is servicing API requests (default true)
      --wait-timeout duration             max time to wait per Kubernetes core services to be healthy. (default 6m0s)```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the kubernetes cluster. (default "kubeadm")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```
