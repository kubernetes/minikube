---
title: "start"
linkTitle: "start"
weight: 1
date: 2019-08-01
description: >
  Starts a local kubernetes cluster
---

### Usage

```
minikube start [flags]
```

### Options

```
--apiserver-ips=[]: A set of apiserver IP Addresses which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine
--apiserver-name='minikubeCA': The apiserver name which is used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine
--apiserver-names=[]: A set of apiserver names which are used in the generated certificate for kubernetes.  This can be used if you want to make the apiserver available from outside the machine
--apiserver-port=8443: The apiserver listening port
--cache-images=true: If true, cache docker images for the current bootstrapper and load them into the machine. Always false with --vm-driver=none.
--container-runtime='docker': The container runtime to be used (docker, crio, containerd).
--cpus=2: Number of CPUs allocated to the minikube VM.
--cri-socket='': The cri socket path to be used.
--disable-driver-mounts=false: Disables the filesystem mounts provided by the hypervisors
--disk-size='20000mb': Disk size allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g).
--dns-domain='cluster.local': The cluster dns domain name used in the kubernetes cluster
--dns-proxy=false: Enable proxy for NAT DNS requests (virtualbox)
--docker-env=[]: Environment variables to pass to the Docker daemon. (format: key=value)
--docker-opt=[]: Specify arbitrary flags to pass to the Docker daemon. (format: key=value)
--download-only=false: If true, only download and cache files for later use - don't install or start anything.
--embed-certs=false: if true, will embed the certs in kubeconfig.
--enable-default-cni=false: Enable the default CNI plugin (/etc/cni/net.d/k8s.conf). Used in conjunction with "--network-plugin=cni".
--extra-config=: A set of key=value pairs that describe configuration that may be passed to different components.
The key should be '.' separated, and the first part before the dot is the component to apply the configuration to.
Valid components are: kubelet, kubeadm, apiserver, controller-manager, etcd, proxy, scheduler
Valid kubeadm parameters: ignore-preflight-errors, dry-run, kubeconfig, kubeconfig-dir, node-name, cri-socket, experimental-upload-certs, certificate-key, rootfs, pod-network-cidr
--feature-gates='': A set of key=value pairs that describe feature gates for alpha/experimental features.
--force=false: Force minikube to perform possibly dangerous operations
--host-dns-resolver=true: Enable host resolver for NAT DNS requests (virtualbox)
--host-only-cidr='192.168.99.1/24': The CIDR to be used for the minikube VM (only supported with Virtualbox driver)
--hyperkit-vpnkit-sock='': Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock.
--hyperkit-vsock-ports=[]: List of guest VSock ports that should be exposed as sockets on the host (Only supported on with hyperkit now).
--hyperv-virtual-switch='': The hyperv virtual switch name. Defaults to first found. (only supported with HyperV driver)
--image-mirror-country='': Country code of the image mirror to be used. Leave empty to use the global one. For Chinese mainland users, set it to cn
--image-repository='': Alternative image repository to pull docker images from. This can be used when you have limited access to gcr.io. Set it to "auto" to let minikube decide one for you. For Chinese mainland users, you may use local gcr.io mirrors such as registry.cn-hangzhou.aliyuncs.com/google_containers
--insecure-registry=[]: Insecure Docker registries to pass to the Docker daemon.  The default service CIDR range will automatically be added.
--iso-url='https://storage.googleapis.com/minikube/iso/minikube-v1.3.0.iso': Location of the minikube iso.
--keep-context=false: This will keep the existing kubectl context and will create a minikube context.
--kubernetes-version='v1.15.2': The kubernetes version that the minikube VM will use (ex: v1.2.3)
--kvm-gpu=false: Enable experimental NVIDIA GPU support in minikube
--kvm-hidden=false: Hide the hypervisor signature from the guest in minikube
--kvm-network='default': The KVM network name. (only supported with KVM driver)
--kvm-qemu-uri='qemu:///system': The KVM QEMU connection URI. (works only with kvm2 driver on linux)
--memory='2000mb': Amount of RAM allocated to the minikube VM (format: <number>[<unit>], where unit = b, k, m or g).
--mount=false: This will start the mount daemon and automatically mount files into minikube.
--mount-string='/Users:/minikube-host': The argument to pass the minikube mount command on start.
--network-plugin='': The name of the network plugin.
--nfs-share=[]: Local folders to share with Guest via NFS mounts (Only supported on with hyperkit now)
--nfs-shares-root='/nfsshares': Where to root the NFS Shares (defaults to /nfsshares, only supported with hyperkit now)
--no-vtx-check=false: Disable checking for the availability of hardware virtualization before the vm is started (virtualbox)
--registry-mirror=[]: Registry mirrors to pass to the Docker daemon
--service-cluster-ip-range='10.96.0.0/12': The CIDR to be used for service cluster IPs.
--uuid='': Provide VM UUID to restore MAC address (only supported with Hyperkit driver).
--vm-driver='virtualbox': VM driver is one of: [virtualbox parallels vmwarefusion hyperkit vmware]
--wait=true: Wait until Kubernetes core services are healthy before exiting.
--wait-timeout=3m0s: max time to wait per Kubernetes core services to be healthy.
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
