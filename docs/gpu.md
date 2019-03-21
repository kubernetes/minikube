# (Experimental) NVIDIA GPU support in minikube

minikube has experimental support for using NVIDIA GPUs on Linux.

## Using NVIDIA GPUs on minikube on Linux with `--vm-driver=kvm2`

When using NVIDIA GPUs with the kvm2 vm-driver. We passthrough spare GPUs on the
host to the minikube VM. Doing so has a few prerequisites:

- You must install the [kvm2 driver](drivers.md#kvm2-driver). If you already had
  this installed make sure that you fetch the latest
  `docker-machine-driver-kvm2` binary that has GPU support.

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
  GPUs with kvm2. Run the following command to start minikube:
  ```shell
  minikube start --vm-driver kvm2 --gpu
  ```

  This command will check if all the above conditions are satisfied and
  passthrough spare GPUs found on the host to the VM.

  If this succeeded, run the following commands:
  ```shell
  minikube addons enable nvidia-gpu-device-plugin
  minikube addons enable nvidia-driver-installer
  ```

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

### Why are so many manual steps required to use GPUs with kvm2 on minikube?

These steps require elevated privileges which minikube doesn't run with and they
are disruptive to the host, so we decided to not do them automatically.

## Using NVIDIA GPU on minikube on Linux with `--vm-driver=none`

NOTE: This approach used to expose GPUs here is different than the approach used
to expose GPUs with `--vm-driver=kvm2`. Please don't mix these instructions.

- Install minikube.

- Install the nvidia driver, nvidia-docker and configure docker with nvidia as
  the default runtime. See instructions at
  <https://github.com/NVIDIA/nvidia-docker>

- Start minikube:
  ```shell
  minikube start --vm-driver=none --apiserver-ips 127.0.0.1 --apiserver-name localhost
  ```

- Install NVIDIA's device plugin:
  ```shell
  kubectl create -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v1.10/nvidia-device-plugin.yml
  ```

## Why does minikube not support NVIDIA GPUs on macOS?

VM drivers supported by minikube for macOS doesn't support GPU passthrough:

- [mist64/xhyve#108](https://github.com/mist64/xhyve/issues/108)
- [moby/hyperkit#159](https://github.com/moby/hyperkit/issues/159)
- [VirtualBox docs](http://www.virtualbox.org/manual/ch09.html#pcipassthrough)

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
  host](http://www.virtualbox.org/manual/ch09.html#pcipassthrough).

- Hyper-V supports DDA (discrete device assignment) but [only for Windows Server
  2016](https://docs.microsoft.com/en-us/windows-server/virtualization/hyper-v/plan/plan-for-deploying-devices-using-discrete-device-assignment)

Since the only possibility of supporting GPUs on minikube on Windows is on a
server OS where users don't usually run minikube, we haven't invested time in
trying to support NVIDIA GPUs on minikube on Windows.

Also, nvidia-docker [doesn't support
Windows](https://github.com/NVIDIA/nvidia-docker/issues/197) either.
