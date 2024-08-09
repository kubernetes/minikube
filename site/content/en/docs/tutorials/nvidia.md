---
title: "Using NVIDIA GPUs with minikube"
linkTitle: "Using NVIDIA GPUs with minikube"
weight: 1
date: 2018-01-02
---

## Prerequisites

- Linux or Windows with WSL2 installed
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
- Start minikube:
  ```shell
  minikube start --driver docker --container-runtime docker --gpus all
  ```
{{% /tab %}}
{{% tab Windows-WSL %}}
## Using the Windows-WSL2 driver

- Endure you have already enabled WSL2. You also need to install the Docker Desktop For Windows.

- Ensure you have an NVIDIA driver installed(via Windows only), if one is not installed follow the [NVIDIA Driver Installation Guide](https://docs.nvidia.com/datacenter/tesla/tesla-installation-notes/index.html)

**Note: Make sure only install the driver on windows, and DO NOT install any linux nvidia driver**

- After instalation of windows driver, you many also need to execute `cp /usr/lib/wsl/lib/nvidia-smi /usr/bin/nvidia-smi` and `chmod ogu+x /usr/bin/nvidia-smi` in WSL2, because otherwise the nvidia-smi may not be found in PATH. You can check if one is installed by running `nvidia-smi`,

-  Install the [Cuda Toolkit for WSL2](https://developer.nvidia.com/cuda-downloads?target_os=Linux&target_arch=x86_64&Distribution=WSL-Ubuntu&target_version=2.0&target_type=deb_local) inside WSL2. Note you need to select targetOS as linux and distribution as WSL-Ubuntu

- Check if `bpf_jit_harden` is set to `0` inside WSL2
  ```shell
  sudo sysctl net.core.bpf_jit_harden
  ```
  - If it's not `0` run:
  ```shell
  echo "net.core.bpf_jit_harden=0" | sudo tee -a /etc/sysctl.conf
  sudo sysctl -p
  ```

- Install the [NVIDIA Container Toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) inside WSL2

- Configure Docker inside WSL2:
  ```shell
  sudo nvidia-ctk runtime configure --runtime=docker && sudo systemctl restart docker
  ```
- Start minikube inside WSL2:
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


## Hand-on try: an example about training ML model in a Pod of minikube k8s cluster
Here is a simplest example program from [Pytorch website](https://pytorch.org/tutorials/beginner/basics/quickstart_tutorial.html), which trains a model on MNIST data set. Have a try on it to see that minikube gpu support actually works.

```python
import torch
from torch import nn
from torch.utils.data import DataLoader
from torchvision import datasets
from torchvision.transforms import ToTensor

# Download training data from open datasets.
training_data = datasets.FashionMNIST(
    root="data",
    train=True,
    download=True,
    transform=ToTensor(),
)

# Download test data from open datasets.
test_data = datasets.FashionMNIST(
    root="data",
    train=False,
    download=True,
    transform=ToTensor(),
)

batch_size = 64

# Create data loaders.
train_dataloader = DataLoader(training_data, batch_size=batch_size)
test_dataloader = DataLoader(test_data, batch_size=batch_size)

for X, y in test_dataloader:
    print(f"Shape of X [N, C, H, W]: {X.shape}")
    print(f"Shape of y: {y.shape} {y.dtype}")
    break
# Get cpu, gpu or mps device for training.
device = (
    "cuda"
    if torch.cuda.is_available()
    else "mps"
    if torch.backends.mps.is_available()
    else "cpu"
)
print(f"Using {device} device")

# Define model
class NeuralNetwork(nn.Module):
    def __init__(self):
        super().__init__()
        self.flatten = nn.Flatten()
        self.linear_relu_stack = nn.Sequential(
            nn.Linear(28*28, 512),
            nn.ReLU(),
            nn.Linear(512, 512),
            nn.ReLU(),
            nn.Linear(512, 10)
        )

    def forward(self, x):
        x = self.flatten(x)
        logits = self.linear_relu_stack(x)
        return logits

model = NeuralNetwork().to(device)
print(model)

loss_fn = nn.CrossEntropyLoss()
optimizer = torch.optim.SGD(model.parameters(), lr=1e-3)

def train(dataloader, model, loss_fn, optimizer):
    size = len(dataloader.dataset)
    model.train()
    for batch, (X, y) in enumerate(dataloader):
        X, y = X.to(device), y.to(device)

        # Compute prediction error
        pred = model(X)
        loss = loss_fn(pred, y)

        # Backpropagation
        loss.backward()
        optimizer.step()
        optimizer.zero_grad()

        if batch % 100 == 0:
            loss, current = loss.item(), (batch + 1) * len(X)
            print(f"loss: {loss:>7f}  [{current:>5d}/{size:>5d}]")

def test(dataloader, model, loss_fn):
    size = len(dataloader.dataset)
    num_batches = len(dataloader)
    model.eval()
    test_loss, correct = 0, 0
    with torch.no_grad():
        for X, y in dataloader:
            X, y = X.to(device), y.to(device)
            pred = model(X)
            test_loss += loss_fn(pred, y).item()
            correct += (pred.argmax(1) == y).type(torch.float).sum().item()
    test_loss /= num_batches
    correct /= size
    print(f"Test Error: \n Accuracy: {(100*correct):>0.1f}%, Avg loss: {test_loss:>8f} \n")

epochs = 5
for t in range(epochs):
    print(f"Epoch {t+1}\n-------------------------------")
    train(train_dataloader, model, loss_fn, optimizer)
    test(test_dataloader, model, loss_fn)
print("Done!")
torch.save(model.state_dict(), "model.pth")
print("Saved PyTorch Model State to model.pth")
```

Start minikube with gpu support:
```shell
minikube start --driver docker --container-runtime docker --gpus all
```

Create a pod using `pytorch/pytorch` image, which have all necessary libraries installed, and get a shell from this pod.
```
kubectl run torch --image=pytorch/pytorch -it -- /bin/bash
```

Now copy the file into the pod, and run it with python3. You will see the model is trained with Nvidia GPU and Cuda acceleration. 

```
... ...
Shape of X [N, C, H, W]: torch.Size([64, 1, 28, 28])
Shape of y: torch.Size([64]) torch.int64
Using cuda device
NeuralNetwork(
  (flatten): Flatten(start_dim=1, end_dim=-1)
  (linear_relu_stack): Sequential(
    (0): Linear(in_features=784, out_features=512, bias=True)
    (1): ReLU()
    (2): Linear(in_features=512, out_features=512, bias=True)
    (3): ReLU()
    (4): Linear(in_features=512, out_features=10, bias=True)
  )
)
Epoch 1
... ...
```


