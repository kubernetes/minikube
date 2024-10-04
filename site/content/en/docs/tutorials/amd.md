---
title: "Using NVIDIA GPUs with minikube"
linkTitle: "Using NVIDIA GPUs with minikube"
weight: 1
date: 2018-01-02
---

## Prerequisites

- Linux
- Latest AMD GPU Drivers
- minikube v1.32.0-beta.0 or later (docker driver only)

## Instructions per driver

{{% tabs %}}
{{% tab docker %}}
## Using the docker driver

- Ensure you have an NVIDIA driver installed, you can check if one is installed by running `nvidia-smi`, if one is not installed follow the [NVIDIA Driver Installation Guide](https://docs.nvidia.com/datacenter/tesla/tesla-installation-notes/index.html)

- Delete existing minikube (optional)

  If you have an existing minikube instance, you may need to delete it if it was built before installing the nvidia runtime shim.
  ```shell
  minikube delete
  ```
  This will make sure minikube does any required setup or addon installs now that the nvidia runtime is available.
  
- Start minikube:
  ```shell
  minikube start --driver docker --container-runtime docker --gpus all
  ```

{{% /tab %}}
{{% tab none %}}
## Using the 'none' driver

NOTE: This approach used to expose GPUs here is different than the approach used
to expose GPUs with `--driver=kvm`. Please don't mix these instructions.

- Install minikube.

- Install the nvidia driver, nvidia-docker and configure docker with nvidia as
  the default runtime. See instructions at
  <https://github.com/NVIDIA/nvidia-docker>

- Start minikube:
  ```shell
  minikube start --driver=none --apiserver-ips 127.0.0.1 --apiserver-name localhost
  ```

- Install AMD's device plugin:
  ```shell
  minikube addons enable amd-gpu-device-plugin
  ```
{{% /tab %}}
{{% tab kvm %}}

### Where can I learn more about GPU passthrough?

See the excellent documentation at
<https://wiki.archlinux.org/index.php/PCI_passthrough_via_OVMF>

## Why does minikube not support NVIDIA GPUs on Windows?

minikube supports Windows host through Hyper-V or VirtualBox.

- VirtualBox doesn't support PCI passthrough for [Windows
  host](https://www.virtualbox.org/manual/ch09.html#pcipassthrough).

- Hyper-V supports DDA (discrete device assignment) but [only for Windows Server
  2016](https://docs.microsoft.com/en-us/windows-server/virtualization/hyper-v/plan/plan-for-deploying-devices-using-discrete-device-assignment)

Since the only possibility of supporting GPUs on minikube on Windows is on a
server OS where users don't usually run minikube, we haven't invested time in
trying to support GPUs on minikube on Windows.
