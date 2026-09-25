---
title: "Mixed Linux/Windows Cluster (Experimental)"
linkTitle: "Mixed Linux/Windows Cluster"
weight: 6
date: 2026-06-11
---

## Overview

This tutorial describes the experimental workflow for creating one Linux control-plane node and one Windows
worker with the Hyper-V driver, using a repeatable `--node` flag.

**This is development-branch documentation, not a feature available in a standard minikube release.**
It is being developed in [kubernetes/minikube#22503](https://github.com/kubernetes/minikube/pull/22503) alongside
the Windows node image pipeline in [windows-node-image-builder](https://github.com/bobsira/windows-node-image-builder).

## Prerequisites

- Windows 10/11 or Windows Server with Hyper-V enabled
- Administrator rights for Hyper-V VM creation
- A minikube binary built from the `feature/windows-node-support` branch, as described below
- Git Bash, `make`, and the Go toolchain required by that branch
- `kubectl` on `PATH`; the experimental Windows networking setup invokes it
- Internet access to download the default hosted Windows VHDX, or a prepared Windows Server 2025 VHDX
  with compatible Kubernetes binaries and container runtime if you choose a custom image
- Enough disk space for the cached VHDX, a separate worker copy, the Linux VM, and container images;
  image size and disk growth vary
- At least **8 GB** of RAM free for the two VMs

## Node specifications and current scope

- Each `--node` occurrence describes one node. `role=control-plane` or `role=worker` is required; `os` defaults
  to `linux`. Node names are generated automatically, and the control plane is ordered before workers.
- `--node role=control-plane` alone creates one Linux control-plane node that can also run workloads. It does
  not create a separate worker. A worker-only specification fails because a Linux control plane is required.
- The full feature branch supports exactly one Linux control plane and one Windows worker for a mixed-OS
  cluster. Windows control planes, additional mixed-OS workers, and HA through `--node` are not supported.
- Do not combine `--node` with `--nodes`/`-n` or `--ha` on `start`. HA must also be disabled in configuration
  and the environment. Existing Linux `--nodes` and `--ha` workflows remain unchanged.
- In the full feature branch, a mixed-OS request selects Hyper-V, Flannel, and containerd if omitted, and
  rejects conflicting explicit settings. The smaller interface PR does not introduce these Windows defaults.
- Use a fresh profile. On an existing profile, `--node` describes the complete saved topology; it does not
  add a Windows worker or convert a Linux node. Matching specifications do not resolve the known Windows
  restart/reprovisioning limitations.
- YAML node definitions and global configuration of node specifications are deferred.

## Tutorial

### 1. Build the experimental binary

In Git Bash, clone and build the full feature branch:

```sh
git clone --branch feature/windows-node-support https://github.com/bobsira/minikube.git minikube-windows
cd minikube-windows
make
```

For the remaining steps, open an **Administrator PowerShell** session in that repository's root directory.
Use `.\out\minikube.exe` rather than a released binary elsewhere on `PATH`.

### 2. Select compatible versions and start the cluster

Use a fresh `mixed-os` profile to keep this experimental cluster separate from the default `minikube`
cluster. Choose one of the following examples.

#### Default: use the hosted Windows image

You do not need to build your own VHDX. The full feature branch uses its configured hosted Windows image,
downloading it automatically when it is not already cached:

```powershell
.\out\minikube.exe start --profile=mixed-os --node role=control-plane --node role=worker,os=windows
```

The node count comes from the two `--node` arguments, so no `--nodes=2` flag is needed. minikube selects
Hyper-V, Flannel, and containerd, acquires the Windows VHDX, and copies it into the worker's machine directory
before booting the VM. It initializes the Linux control plane, joins the Windows worker, and applies the
Windows networking configuration.

**Do not assume the hosted image and branch defaults are compatible.** In the September 24, 2026 development
run, the hosted Windows image reported kubelet `v1.37.0`, while the default Linux control plane was `v1.36.4`.
A kubelet newer than the API server is unsupported, even if the node reaches `Ready`.

The Windows image contains its own Kubernetes binaries. `--kubernetes-version` selects the cluster version;
it does **not** replace the kubelet already installed in the Windows VHDX. Check the image's versions before
use, select a compatible control-plane version, and check the Windows kube-proxy version against the
[Kubernetes version-skew policy](https://kubernetes.io/releases/version-skew-policy/).

#### Optional: use a custom Windows image

Use this alternative when you want to supply your own image and select the matching Kubernetes version.
The example below assumes you have prepared a Windows Server 2025 image with Kubernetes `v1.36.4` binaries.
Replace the version and image path together to match your chosen image; see
[Building a custom Windows node image](#building-a-custom-windows-node-image).

Use an isolated minikube home when testing a different image. An existing cached Windows VHDX is reused;
changing `--windows-vhd-url` or the profile name alone does not replace it. Choose an unused directory and
keep `MINIKUBE_HOME` set for all subsequent minikube commands in this session.

```powershell
$env:MINIKUBE_HOME = 'C:\minikube-homes\windows-image-test'
$kubernetesVersion = 'v1.36.4'
$windowsVhd = 'C:\vhd\hybrid-minikube-windows-server.vhdx'

.\out\minikube.exe start --profile=mixed-os `
  --node role=control-plane `
  --node role=worker,os=windows `
  --kubernetes-version=$kubernetesVersion `
  --windows-vhd-url=$windowsVhd
```

Download, copy, boot, and join times depend on the image, cache, storage, network, and available host resources.

### 3. Verify the cluster and networking

```powershell
kubectl --context=mixed-os get nodes -o wide -L kubernetes.io/os
kubectl --context=mixed-os get pods -A -o wide
kubectl --context=mixed-os -n kube-flannel rollout status daemonset/kube-flannel-ds-windows-amd64 --timeout=5m
kubectl --context=mixed-os -n kube-system rollout status daemonset/kube-proxy-windows --timeout=5m
```

Expect two nodes: `mixed-os` (Linux control plane) and `mixed-os-m02` (Windows Server 2025 worker), both
`Ready`. Check their reported Kubernetes versions as well as readiness.

**Node readiness is not enough:** Windows Flannel and kube-proxy must also be healthy before testing a
workload. Windows kube-proxy can restart during initial HNS/network setup; this remains an experimental
limitation, not a guarantee that repeated restarts will resolve themselves.

### 4. Deploy a Windows workload

Windows containers require a `nodeSelector` for `kubernetes.io/os: windows` and a toleration for the
`node.kubernetes.io/os=windows:NoSchedule` taint used by this experimental cluster. The container image must
also be compatible with the Windows host; this example uses Windows Server 2025.

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

```powershell
kubectl --context=mixed-os apply -f win-webserver.yaml
```

Wait for the pod to start (Windows container images are large; first pull may take a few minutes):

```powershell
kubectl --context=mixed-os rollout status deployment/win-webserver --timeout=10m
```

Confirm the pod landed on the Windows node:

```powershell
kubectl --context=mixed-os get pods -l app=win-webserver -o wide
```

The pod should be running on `mixed-os-m02`. Get the service URL:

```powershell
.\out\minikube.exe --profile=mixed-os service win-webserver --url
```

Open the returned URL to check the response. A running pod alone does not verify NodePort connectivity.

### 5. Clean up

To remove only the example workload while keeping the cluster, run:

```powershell
kubectl --context=mixed-os delete -f win-webserver.yaml
```

When you are finished with the cluster, delete the tutorial profile from Administrator PowerShell:

```powershell
.\out\minikube.exe delete --profile=mixed-os
```

This deletes the `mixed-os` cluster's VMs, profile, and kubeconfig context. **Workloads and data stored in
those VMs are also deleted.** You do not need to delete the example workload separately before deleting
the cluster. Other cluster profiles and the cached Windows VHDX are retained; do not use `--all` or
`--purge` for this tutorial's cleanup.

If you used the custom-image example, keep `MINIKUBE_HOME` set to the same directory used to create the
cluster until deletion is complete. You can then restore its previous setting.

Confirm that neither tutorial VM remains in Hyper-V:

```powershell
Get-VM | Where-Object { $_.Name -in @('mixed-os', 'mixed-os-m02') }
```

This should return no VMs. If either remains, review the deletion output for errors before continuing.

## Building a custom Windows node image

The VHDX must already contain Windows Server 2025, a container runtime, compatible Kubernetes node binaries,
and the prerequisites required by the feature branch. minikube boots the prepared disk; it does not install
Windows from an ISO.

Use [windows-node-image-builder](https://github.com/bobsira/windows-node-image-builder), the currently used
Windows node image builder, to build a custom image.
Record the image's Windows build and Kubernetes versions so the setup is reproducible.

`--windows-vhd-url` accepts a remote URL, a local absolute path, or a `file://` URI. For a remotely hosted
custom image, replace the local `$windowsVhd` assignment in the custom-image example with:

```powershell
$windowsVhd = 'https://your-storage/your-image.vhdx'
```

If `--windows-vhd-url` is omitted, the branch uses its configured hosted image. That experimental Azure Blob
image is maintained by a Microsoft contributor, is **not an official minikube release artifact**, and may
change or be removed. Check its versions before use. Use a new isolated minikube home when changing images;
the cache is reused based on the existing file, not the newly supplied URL.

## Troubleshooting

### Windows provisioning is not available in this build

The smaller `--node` interface PR deliberately rejects Windows provisioning. Build the full
`feature/windows-node-support` branch and use its `.\out\minikube.exe` binary for this tutorial.

### The Windows node or networking pods are not ready

Inspect the node, pod events, and networking logs:

```powershell
kubectl --context=mixed-os describe node mixed-os-m02
kubectl --context=mixed-os get pods -A -o wide
kubectl --context=mixed-os -n kube-flannel logs daemonset/kube-flannel-ds-windows-amd64 --all-containers=true --tail=100
kubectl --context=mixed-os -n kube-system logs daemonset/kube-proxy-windows --all-containers=true --tail=100
```

Use `kubectl describe pod` for a failing pod and `kubectl logs --previous` for a restarted container.
Check the Kubernetes version compatibility and Windows HNS/networking state rather than assuming that
`Ready` nodes or a successful join prove the network is working.

To inspect the Windows VM, open an SSH session:

```powershell
.\out\minikube.exe --profile=mixed-os ssh -n mixed-os-m02
```

The `-n` above selects a node for `ssh`; it is not the `start --nodes` flag. Run diagnostics such as
`Get-Service kubelet,containerd` in a PowerShell session inside the Windows VM. The SSH command without
`-n` targets the Linux control plane.

### First start is slow

A remote VHDX is downloaded and cached on first use. Creating a Windows worker also requires a separate
copy in its machine directory. Warm-cache creation avoids the download but not that copy, and the first
Windows workload may need to pull a large container image. Do not rely on fixed startup-time estimates.

### Hyper-V permission error

Run cluster creation from an **elevated (Administrator) PowerShell** prompt.

### kubeadm join failed or timed out

The current implementation runs a single Windows `kubeadm join` using kubeadm's phase timeouts and reports
its output on failure. It does not automatically replay a failed join. Inspect the startup output,
kubelet/container runtime state, and control-plane connectivity before retrying.

### Restarting an existing Windows worker fails

Windows restart/reprovisioning remains follow-up work. Repeating the same `--node` arguments validates the
saved topology, but does not fix that lifecycle limitation. Use a fresh test profile for new experiments;
do not assume Linux stop/start behavior is fully supported for the Windows worker.
