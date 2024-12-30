---
title: "Using NVIDIA GPUs with minikube"
linkTitle: "Using NVIDIA GPUs with minikube"
weight: 1
date: 2018-01-02
---

## Prerequisites

- Linux
- Latest NVIDIA GPU drivers
- minikube v1.32.0-beta.0 or later (docker driver only)

## Instructions per driver

{{% tabs %}}
{{% tab docker %}}
## Using the docker driver

- Ensure you have an NVIDIA driver installed, you can check if one is installed by running `nvidia-smi`, if one is not installed follow the [NVIDIA Driver Installation Guide](https://docs.nvidia.com/datacenter/tesla/tesla-installation-notes/index.html)

- Check if `bpf_jit_harden` is set to `0`
  ```shell
  sudo sysctl net.core.bpf_jit_harden
  ```

  - If it's not `0` run:
  ```shell
  echo "net.core.bpf_jit_harden=0" | sudo tee -a /etc/sysctl.conf
  sudo sysctl -p
  ```

- Install the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) on your host machine

- Configure Docker:
  ```shell
  sudo nvidia-ctk runtime configure --runtime=docker && sudo systemctl restart docker
  ```

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

- Install NVIDIA's device plugin:
  ```shell
  minikube addons enable nvidia-device-plugin
  ```
{{% /tab %}}
{{% tab kvm %}}
## Using the kvm driver

When using NVIDIA GPUs with the kvm driver, we passthrough spare GPUs on the
host to the minikube VM. Doing so has a few prerequisites:

- You must install the [kvm driver]({{< ref "/docs/drivers/kvm2" >}}) If you already had
  this installed make sure that you fetch the latest
  `docker-machine-driver-kvm` binary that has GPU support.

- Your CPU must support IOMMU. Different vendors have different names for this
  technology. Intel calls it Intel VT-d. AMD calls it AMD-Vi. Your motherboard
  must also support IOMMU.

- You must enable IOMMU in the kernel: add `intel_iommu=on` or `amd_iommu=on`
  (depending to your CPU vendor) to the kernel command line. Also add `iommu=pt`
  to the kernel command line.

- You must have spare GPUs that are not used on the host and can be passthrough
  to the VM. These GPUs must not be controlled by the nvidia/nouveau driver. You
  can ensure this by either not loading the nvidia/nouveau driver on the host at
  all or assigning the spare GPU devices to stub kernel modules like `vfio-pci`
  or `pci-stub` at boot time. You can do that by adding the
  [vendorId:deviceId](https://pci-ids.ucw.cz/read/PC/10de) of your spare GPU to
  the kernel command line. For ex. for Quadro M4000 add `pci-stub.ids=10de:13f1`
  to the kernel command line. Note that you will have to do this for all GPUs
  you want to passthrough to the VM and all other devices that are in the IOMMU
  group of these GPUs.

- Once you reboot the system after doing the above, you should be ready to use
  GPUs with kvm. Run the following command to start minikube:
  ```shell
  minikube start --driver kvm --kvm-gpu
  ```

  This command will check if all the above conditions are satisfied and
  passthrough spare GPUs found on the host to the VM.

  If this succeeded, run the following commands:
  ```shell
  minikube addons enable nvidia-device-plugin
  minikube addons enable nvidia-driver-installer
  ```
  NOTE: `nvidia-gpu-device-plugin` addon has been deprecated and it's functionality is merged inside of `nvidia-device-plugin` addon.

  This will install the NVIDIA driver (that works for GeForce/Quadro cards)
  on the VM.

- If everything succeeded, you should be able to see `nvidia.com/gpu` in the
  capacity:
  ```shell
  kubectl get nodes -ojson | jq .items[].status.capacity
  ```

### Where can I learn more about GPU passthrough?

See the excellent documentation at
<https://wiki.archlinux.org/index.php/PCI_passthrough_via_OVMF>

### Why are so many manual steps required to use GPUs with kvm on minikube?

These steps require elevated privileges which minikube doesn't run with and they
are disruptive to the host, so we decided to not do them automatically.
{{% /tab %}}
{{% /tabs %}}

## Why does minikube not support NVIDIA GPUs on macOS?

drivers supported by minikube for macOS doesn't support GPU passthrough:

- [mist64/xhyve#108](https://github.com/mist64/xhyve/issues/108)
- [moby/hyperkit#159](https://github.com/moby/hyperkit/issues/159)
- [VirtualBox docs](https://www.virtualbox.org/manual/ch09.html#pcipassthrough)

Also: 

- For quite a while, all Mac hardware (both laptops and desktops) have come with
  Intel or AMD GPUs (and not with NVIDIA GPUs). Recently, Apple added [support
  for eGPUs](https://support.apple.com/en-us/HT208544), but even then all the
  supported GPUs listed are AMD’s.

- nvidia-docker [doesn't support
  macOS](https://github.com/NVIDIA/nvidia-docker/issues/101) either.

## Why does minikube not support NVIDIA GPUs on Windows?

minikube supports Windows host through Hyper-V or VirtualBox.

- VirtualBox doesn't support PCI passthrough for [Windows
  host](https://www.virtualbox.org/manual/ch09.html#pcipassthrough).

- Hyper-V supports DDA (discrete device assignment) but [only for Windows Server
  2016](https://docs.microsoft.com/en-us/windows-server/virtualization/hyper-v/plan/plan-for-deploying-devices-using-discrete-device-assignment)

Since the only possibility of supporting GPUs on minikube on Windows is on a
server OS where users don't usually run minikube, we haven't invested time in
trying to support NVIDIA GPUs on minikube on Windows.

Also, nvidia-docker [doesn't support
Windows](https://github.com/NVIDIA/nvidia-docker/issues/197) either.
