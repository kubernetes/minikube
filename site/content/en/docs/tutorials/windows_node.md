---
title: "Mixed Linux/Windows Cluster (Experimental)"
linkTitle: "Mixed Linux/Windows Cluster"
weight: 6
date: 2026-06-11
---

## Overview

This tutorial shows how to start a minikube cluster with a Linux control-plane node and a Windows worker node
using the Hyper-V driver. After completing it you will be able to schedule Windows container workloads alongside
Linux workloads on a local Kubernetes cluster.

This feature is **experimental** and currently only supported on the Hyper-V driver (Windows host). It is
being developed in [kubernetes/minikube#22503](https://github.com/kubernetes/minikube/pull/22503) alongside the
Windows node image pipeline in [kubernetes-sigs/minikube-os#2](https://github.com/kubernetes-sigs/minikube-os/pull/2).

The CI integration test infrastructure for Windows nodes gives us confidence to ship this as experimental even
for contributors who do not have a Windows machine available to test on their own.

## Prerequisites

- Windows 10/11 or Windows Server with Hyper-V enabled
- Administrator rights (the Hyper-V Default Switch requires elevation)
- minikube v1.39.0 or higher built from [#22503](https://github.com/kubernetes/minikube/pull/22503)
- kubectl
- At least **30 GB** of free disk space (the Windows VHD is ~22 GB and is cached in `~/.minikube/cache/`)
- At least **8 GB** of RAM free for the two VMs

## Caveats

- **Experimental**: This feature may change between releases. It is not yet enabled on a default cluster profile.
- Only the **Hyper-V** driver is supported. Running `minikube start --node-os='[linux,windows]'` automatically
  sets `--driver=hyperv`, `--cni=flannel`, and `--container-runtime=containerd`.
- The `--nodes` flag **must** be set to `2`. Clusters with more than one Windows worker are not yet supported.
- The Windows VHD (`hybrid-minikube-windows-server.vhdx`, ~22 GB) is downloaded automatically on first use and
  cached. Subsequent starts reuse the cache but still copy the file to the VM's machine directory, which takes
  approximately **4–5 minutes**. The default VHD is currently hosted on Azure Blob Storage maintained by a
  Microsoft contributor. You can build and host your own image using the
  [minikube-os](https://github.com/kubernetes-sigs/minikube-os/pull/2) pipeline and override the download
  location with `--windows-vhd-url=<your-url>`.
- The Windows node pre-installs its own kubelet version (v1.35.0 in Windows Server 2025 images). The node will
  join the cluster regardless of the control-plane Kubernetes version.
- `minikube ssh` targets the control-plane (Linux) node by default. To SSH into the Windows node, use
  `minikube ssh -n minikube-m02`.

## Tutorial

### 1. Start the cluster

```shell
minikube start -n 2 --node-os='[linux,windows]' --kubernetes-version=v1.34.0
```

minikube automatically selects the Hyper-V driver and flannel CNI. The Windows VHD is downloaded and copied
to the VM directory on the first run (allow **10–15 minutes** for first start):

```
* minikube v1.39.0 on Microsoft Windows 11 Enterprise N 10.0.26200
* Automatically selected the hyperv driver. Other choices: ssh, virtualbox
* Starting "minikube" primary control-plane node in "minikube" cluster
* Creating hyperv VM (CPUs=2, Memory=4000MB, Disk=20000MB) ...
* Preparing Kubernetes v1.34.0 on containerd 2.3.1 ...
* Configuring Flannel (Container Networking Interface) ...
* Verifying Kubernetes components...
* Enabled addons: storage-provisioner, default-storageclass
*
* Starting worker node minikube-m02 in cluster minikube
* Downloading Windows VHD (hybrid-minikube-windows-server.vhdx, ~22 GB) ...
* Copying VHD to machine directory (~5 min) ...
* Creating hyperv VM (CPUs=2, Memory=4000MB) ...
* Waiting for Windows node to boot and join the cluster ...
* Applying Windows flannel and kube-proxy manifests ...
* Done! kubectl is now configured to use "minikube" cluster and "default" namespace by default
```

### 2. Verify the cluster

```shell
kubectl get nodes -o wide
```

```
NAME           STATUS   ROLES           AGE   VERSION   INTERNAL-IP     EXTERNAL-IP   OS-IMAGE                                    KERNEL-VERSION     CONTAINER-RUNTIME
minikube       Ready    control-plane   5m    v1.34.0   172.26.217.88   <none>        Buildroot 2025.02.14                        6.6.95             containerd://2.3.1
minikube-m02   Ready    <none>          4m    v1.35.0   172.26.212.15   <none>        Windows Server 2025 Datacenter Evaluation   10.0.26100.32690   containerd://2.2.3
```

Both nodes show `Ready`. The Windows node reports its own pre-installed kubelet version (`v1.35.0`) and
`Windows Server 2025 Datacenter Evaluation` as the OS image.



### 3. Deploy a Windows workload

Windows containers require a `nodeSelector` for `kubernetes.io/os: windows` and a toleration for the
`node.kubernetes.io/os=windows:NoSchedule` taint.

Save the following as `win-webserver.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: win-webserver
  labels:
    app: win-webserver
spec:
  ports:
    - port: 80
      targetPort: 80
  selector:
    app: win-webserver
  type: NodePort
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: win-webserver
  name: win-webserver
spec:
  replicas: 1
  selector:
    matchLabels:
      app: win-webserver
  template:
    metadata:
      labels:
        app: win-webserver
      name: win-webserver
    spec:
      nodeSelector:
        kubernetes.io/os: windows
      tolerations:
        - key: "node.kubernetes.io/os"
          operator: "Equal"
          value: "windows"
          effect: "NoSchedule"
      containers:
        - name: windowswebserver
          image: mcr.microsoft.com/windows/servercore:ltsc2025
          command:
            - powershell.exe
            - -command
            - |
              $listener = New-Object System.Net.HttpListener
              $listener.Prefixes.Add('http://*:80/')
              $listener.Start()
              Write-Host 'Listening at http://*:80/'
              while ($listener.IsListening) {
                $ctx = $listener.GetContext()
                $resp = $ctx.Response
                $content = [System.Text.Encoding]::UTF8.GetBytes('Hello from Windows!')
                $resp.ContentLength64 = $content.Length
                $resp.OutputStream.Write($content, 0, $content.Length)
                $resp.Close()
              }
```

Apply it:

```shell
kubectl apply -f win-webserver.yaml
```

```
service/win-webserver created
deployment.apps/win-webserver created
```

Wait for the pod to start (Windows container images are large; first pull may take a few minutes):

```shell
kubectl rollout status deployment/win-webserver
```

```
deployment "win-webserver" successfully rolled out
```

Confirm the pod landed on the Windows node:

```shell
kubectl get pods -o wide
```

```
NAME                            READY   STATUS    RESTARTS   AGE   IP           NODE           NOMINATED NODE   READINESS GATES
win-webserver-xxxxxxxxxx-xxxxx  1/1     Running   0          2m    10.244.1.5   minikube-m02   <none>           <none>
```

Access the service:

```shell
minikube service win-webserver
```

## Building a custom Windows node image

The default VHD is built by the [minikube-os](https://github.com/kubernetes-sigs/minikube-os/pull/2) pipeline
and is currently hosted on Azure Blob Storage maintained by a Microsoft contributor. If you want to build your
own image — for example to use a different Windows version, include custom software, or self-host the download
— the minikube-os repository documents the full build process.

Once you have your own VHD hosted somewhere accessible, pass its URL at cluster creation time:

```shell
minikube start -n 2 --node-os='[linux,windows]' --windows-vhd-url=https://your-storage/your-image.vhdx
```

## Troubleshooting

**The Windows node is stuck in `NotReady`**

The Windows kubelet can take up to 2 minutes to become healthy after the VM boots. Wait and re-check with
`kubectl get nodes`. If the node remains `NotReady` after 5 minutes, check the kubelet logs via SSH:

```shell
minikube ssh -n minikube-m02 -- Get-EventLog -LogName Application -Source kubelet -Newest 20
```

**First start is very slow**

The ~22 GB VHD is downloaded once and cached in `~/.minikube/cache/`. After the first download, subsequent
starts only copy the file to the VM directory (~5 minutes). There is no way to skip this copy today.

**`minikube start` fails with Hyper-V permission error**

Run the command from an **elevated (Administrator) PowerShell** prompt. Hyper-V VM creation requires
administrator rights.

**kubeadm join timed out**

The first join attempt sometimes times out after 1 minute if the Windows VM is still initialising its
services. minikube automatically retries the join — this is expected behaviour and does not require
intervention.
