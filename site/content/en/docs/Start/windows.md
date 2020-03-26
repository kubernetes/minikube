---
title: "Windows"
linkTitle: "Windows"
weight: 3
---

### Prerequisites

* Windows 8 or above
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



{{% tabs %}}
{{% tab "Docker" %}}
{{% readfile file="/docs/Reference/Drivers/includes/docker_usage.inc" %}}
{{% /tab %}}

{{% tab "Hyper-V" %}}
## Check Hypervisor 
{{% readfile file="/docs/Reference/Drivers/includes/check_virtualization_windows.inc" %}}

{{% readfile file="/docs/Reference/Drivers/includes/hyperv_usage.inc" %}}
{{% /tab %}}
{{% tab "VirtualBox" %}}
## Check Hypervisor 
{{% readfile file="/docs/Reference/Drivers/includes/check_virtualization_windows.inc" %}}

{{% readfile file="/docs/Reference/Drivers/includes/virtualbox_usage.inc" %}}
{{% /tab %}}
{{% /tabs %}}

{{% readfile file="/docs/Start/includes/post_install.inc" %}}
