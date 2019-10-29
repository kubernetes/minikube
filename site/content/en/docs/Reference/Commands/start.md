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
--addons                            Enable addons. see `minikube addons list` for a list of valid addon names.
--apiserver-ips ipSlice             A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine (default [])
--apiserver-name string             The apiserver name which is used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine (default "minikubeCA")
--apiserver-names stringArray       A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine
--apiserver-port int                The apiserver listening port (default 8443)
--cache-images                      If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --vm-driver=none. (default true)
--container-runtime string          The container runtime to be used (docker, crio, containerd). (default "docker")
--cpus int                          Number of CPUs allocated to the minikube VM. (default 2)
--cri-socket string                 The cri socket path to be used.
--disable-driver-mounts             Disables the filesystem mounts provided by the hypervisors
--disk-size string                  Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g). (default "20000mb")
--dns-domain string                 The cluster dns domain name used in the kubernetes cluster (default "cluster.local")
--dns-proxy                         Enable proxy for NAT DNS requests (virtualbox)
--docker-env stringArray            Environment variables to pass to the Docker daemon. (format: key=value)
--docker-opt stringArray            Specify arbitrary flags to pass to the Docker daemon. (format: key=value)
--download-only                     If true, only download and cache files for later use - don't install or start anything.
--embed-certs                       if true, will embed the certs in kubeconfig.
--enable-default-cni                Enable the default CNI plugin (/etc/cni/net.d/k8s.conf). Used in conjunction with "--network-plugin=cni".
--extra-config ExtraOption          A set of key=value pairs that describe configuration that may be passed to different components.
The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
Valid kubeadm parameters: ignore-preflight-errors, dry-run, kubeconfig, kubeconfig-dir, node-name, cri-socket, experimental-upload-certs, certificate-key, rootfs, pod-network-cidr
--feature-gates string              A set of key=value pairs that describe feature gates for alpha/experimental features.
--force                             Force minikube to perform possibly dangerous operations
-h, --help                              help for start
--host-dns-resolver                 Enable host resolver for NAT DNS requests (virtualbox) (default true)
--host-only-cidr string             The CIDR to be used for the minikube VM (only supported with Virtualbox driver) (default "192.168.99.1/24")
--hyperkit-vpnkit-sock string       Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock.
--hyperkit-vsock-ports strings      List of guest VSock ports that should be exposed as sockets on the host (Only supported on with hyperkit now).
--hyperv-virtual-switch string      The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)
--image-mirror-country string       Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn
--image-repository string           Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to "auto" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers
--insecure-registry strings         Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.
--iso-url string                    Location of the minikube iso. (default "https://storage.googleapis.com/minikube/iso/minikube-v1.3.0.iso")
--keep-context                      This will keep the existing kubectl context and will create a minikube context.
--kubernetes-version string         The kubernetes version that the minikube VM will use (ex: v1.2.3) (default "v1.15.2")
--kvm-gpu                           Enable experimental NVIDIA GPU support in minikube
--kvm-hidden                        Hide the hypervisor signature from the guest in minikube
--kvm-network string                The KVM network name. (only supported with KVM driver) (default "default")
--kvm-qemu-uri string               The KVM QEMU connection URI. (works only with kvm2 driver on linux) (default "qemu:///system")
--memory string                     Amount of RAM allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g). (default "2000mb")
--mount                             This will start the mount daemon and automatically mount files into minikube.
--mount-string string               The argument to pass the minikube mount command on start. (default "/Users:/minikube-host")
--network-plugin string             The name of the network plugin.
--nfs-share strings                 Local folders to share with Guest via NFS mounts (Only supported on with hyperkit now)
--nfs-shares-root string            Where to root the NFS Shares (defaults to /nfsshares, only supported with hyperkit now) (default "/nfsshares")
--no-vtx-check                      Disable checking for the availability of hardware virtualization before the vm is started (virtualbox)
--registry-mirror strings           Registry mirrors to pass to the Docker daemon
--service-cluster-ip-range string   The CIDR to be used for service cluster IPs. (default "10.96.0.0/12")
--uuid string                       Provide VM UUID to restore MAC address (only supported with Hyperkit driver).
--vm-driver string                  VM driver is one of: [virtualbox parallels vmwarefusion hyperkit vmware] (default "virtualbox")
--wait                              Wait until Kubernetes core services are healthy before exiting. (default true)
--wait-timeout duration             max time to wait per Kubernetes core services to be healthy. (default 3m0s)
```

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
