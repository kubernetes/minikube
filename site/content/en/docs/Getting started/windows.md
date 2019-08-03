---
title: "Windows"
linkTitle: "Windows"
weight: 3
---

### Prerequisites

  * Windows 8 or above
  * A hypervisor, such as Hyper-V or VirtualBox
  * Hardware virtualization support must be enabled in BIOS
  * 4GB of RAM

### Installation

{{% tabs %}}
{{% tab "Direct" %}}
Download and run the [installer](https://storage.googleapis.com/minikube/releases/latest/minikube-installer.exe)
{{% /tab %}}

{{% tab "Chocolatey" %}}

If you have the [Chocolatey Package Manager](https://chocolatey.org/) installed, you can install minikube if run as an Administrator:

```shell
choco install minikube
```

After it has installed, close the current CLI session and reopen it. minikube should have been added to your path automatically.
{{% /tab %}}
{{% /tabs %}}


## Hypervisor Setup

To check if virtualization is supported, run the following command on your Windows terminal or command prompt.

```shell
systeminfo
```
If you see the following output, virtualization is supported:

```shell
Hyper-V Requirements:     VM Monitor Mode Extensions: Yes
                          Virtualization Enabled In Firmware: Yes
                          Second Level Address Translation: Yes
                          Data Execution Prevention Available: Yes
```

If you see the following output, your system already has a Hypervisor installed and you can skip the next step.

```
Hyper-V Requirements:     A hypervisor has been detected.
                          Features required for Hyper-V will not be displayed.
```

{{% tabs %}}
{{% tab "VirtualBox" %}}
{{% readfile file="/docs/Getting started/_virtualbox.md" %}}
{{% /tab %}}
{{% tab "Hyper-V" %}}

If Hyper-V is active, you can start minikube with Hyper-V support using:

```shell
minikube start --vm-driver=hyperv
```

NOTE: If this fails due to networking issues, see the [Hyper-V driver documentation](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyper-v-driver) for further instructions.

To make hyperv the default for future invocations:

```shell
minikube config set vm-driver hyperv
```

{{% /tab %}}
{{% /tabs %}}

{{% readfile file="/docs/Getting started/_post_install.md" %}}