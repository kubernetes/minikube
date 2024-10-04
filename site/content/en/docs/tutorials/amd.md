---
title: "Using AMD GPUs with minikube"
linkTitle: "Using AMD GPUs with minikube"
weight: 1
date: 2024-10-04
---

## Prerequisites

- Linux
- Latest AMD GPU Drivers
- minikube v1.32.0-beta.0 or later (docker driver only)

## Instructions per driver

{{% tabs %}}
{{% tab docker %}}
## Using the docker driver

- Ensure you have an AMD driver installed, you can check if one is installed by running `rocminfo`, if one is not installed follow the [Radeon™ Driver Installation Guide](https://amdgpu-install.readthedocs.io/en/latest/)

- Delete existing minikube (optional)

  If you have an existing minikube instance, you may need to delete it if it was built before installing the AMD drivers.
  ```shell
  minikube delete
  ```
  This will make sure minikube does any required setup or addon installs now that the nvidia runtime is available.
  
- Start minikube:
  ```shell
  minikube start --driver docker --container-runtime docker --gpus amd
  ```

{{% /tab %}}
{{% /tabs %}}

### Where can I learn more about GPU passthrough?

See the excellent documentation at
<https://wiki.archlinux.org/index.php/PCI_passthrough_via_OVMF>

## Why does minikube not support AMD GPUs on Windows?

minikube supports Windows host through Hyper-V or VirtualBox.

- VirtualBox doesn't support PCI passthrough for [Windows
  host](https://www.virtualbox.org/manual/ch09.html#pcipassthrough).

- Hyper-V supports DDA (discrete device assignment) but [only for Windows Server
  2016](https://docs.microsoft.com/en-us/windows-server/virtualization/hyper-v/plan/plan-for-deploying-devices-using-discrete-device-assignment)

Since the only possibility of supporting GPUs on minikube on Windows is on a
server OS where users don't usually run minikube, we haven't invested time in
trying to support GPUs on minikube on Windows.
