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
Download and run the [minikube installer](https://storage.googleapis.com/minikube/releases/latest/minikube-installer.exe)
{{% /tab %}}

{{% tab "Chocolatey" %}}

If the [Chocolatey Package Manager](https://chocolatey.org/) is installed, use it to install minikube:

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

```shell
Hyper-V Requirements:     A hypervisor has been detected.
```

{{% tabs %}}
{{% tab "Hyper-V" %}}
{{% readfile file="/docs/Reference/Drivers/includes/hyperv_usage.inc" %}}
{{% /tab %}}
{{% tab "VirtualBox" %}}
{{% readfile file="/docs/Reference/Drivers/includes/virtualbox_usage.inc" %}}
{{% /tab %}}
{{% /tabs %}}

{{% readfile file="/docs/Start/includes/post_install.inc" %}}
