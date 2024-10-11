---
title: "Using AMD GPUs with minikube"
linkTitle: "Using AMD GPUs with minikube"
weight: 1
date: 2024-10-04
---

This tutorial shows how to start minikube with support for AMD GPUs.

Support is provided by the [AMD GPU device plugin for Kubernetes](https://github.com/ROCm/k8s-device-plugin).


## Prerequisites

- Linux
- Latest AMD GPU Drivers 6.2.1 or greater
- minikube v1.35.0 or later (docker driver only)

## Using the docker driver

- Ensure you have an AMD driver installed, you can check if one is installed by running `rocminfo`, if one is not installed follow the [Radeon™ Driver Installation Guide](https://amdgpu-install.readthedocs.io/en/latest/)

- Delete existing minikube (optional)

  If you have an existing minikube instance, you may need to delete it if it was built before installing the AMD drivers.
  ```shell
  minikube delete
  ```
  
- Start minikube:
  ```shell
  minikube start --driver docker --container-runtime docker --gpus amd
  ```

## Verifying the GPU is available

Test the AMD GPUs are available to the cluster.

1. Create the following Job:

    ```shell
    cat <<'EOF' | kubectl apply -f -
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: amd-gpu-check
      labels:
        purpose: amd-gpu-check
    spec:
      ttlSecondsAfterFinished: 100
      template:
        spec:
          restartPolicy: Never
          securityContext:
            supplementalGroups: 
            - 44
            - 110
          containers:
            - name: amd-gpu-checker
              image: rocm/rocm-terminal
              workingDir: /root
              command: ["rocminfo"]
              args: []
              resources:
                limits:
                  amd.com/gpu: 1 # requesting a GPU
    EOF
    ```

2. Check the Job output `kubectl logs jobs/amd-gpu-check` looks something like the following:

    ```plain
    ROCk module version 6.8.5 is loaded
    =====================    
    HSA System Attributes    
    =====================    
    Runtime Version:         1.14
    Runtime Ext Version:     1.6
    System Timestamp Freq.:  1000.000000MHz
    Sig. Max Wait Duration:  18446744073709551615 (0xFFFFFFFFFFFFFFFF) (timestamp count)
    Machine Model:           LARGE                              
    System Endianness:       LITTLE                             
    Mwaitx:                  DISABLED
    DMAbuf Support:          YES

    ==========               
    HSA Agents               
    ==========               
    *******                  
    Agent 1                  
    *******                  
      Name:                    AMD Ryzen 7 7840U w/ Radeon  780M Graphics
      Uuid:                    CPU-XX                             
    ...
    ```

## Where can I learn more about GPU passthrough?

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
