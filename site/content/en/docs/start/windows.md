---
title: "Windows"
linkTitle: "Windows"
weight: 3
aliases:
  - /docs/start/windows/
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
{{% readfile file="/docs/drivers/includes/docker_usage.inc" %}}
{{% /tab %}}

{{% tab "Hyper-V" %}}

## Check Hypervisor 

{{% readfile file="/docs/drivers/includes/check_virtualization_windows.inc" %}}

{{% readfile file="/docs/drivers/includes/hyperv_usage.inc" %}}
{{% /tab %}}
{{% tab "VirtualBox" %}}

## Check Hypervisor

{{% readfile file="/docs/drivers/includes/check_virtualization_windows.inc" %}}

{{% readfile file="/docs/drivers/includes/virtualbox_usage.inc" %}}
{{% /tab %}}
{{% /tabs %}}

{{% readfile file="/docs/start/includes/post_install.inc" %}}
