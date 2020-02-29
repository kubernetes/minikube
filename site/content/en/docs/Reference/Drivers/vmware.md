---
title: "vmware"
linkTitle: "vmware"
weight: 6
date: 2018-08-08
description: >
  VMware driver
---

## Overview

The vmware driver supports virtualization across all VMware based hypervisors.

{{% tabs %}}
{{% tab "macOS" %}}
{{% readfile file="/docs/Reference/Drivers/includes/vmware_macos_usage.inc" %}}
{{% /tab %}}
{{% tab "Linux" %}}
No documentation is available yet.
{{% /tab %}}
{{% tab "Windows" %}}
1. Download docker-machine-driver-vmware from below link.
https://github.com/machine-drivers/docker-machine-driver-vmware/releases/download/v0.1.0/docker-machine-driver-vmware_windows_amd64.exe
2. Rename it to docker-machine-driver-vmware (without .exe)
3. Copy it to system32 folder or add existing folder(where file is) to environment variable.
4. Add C:\Program Files (x86)\VMware\VMware Workstation to environment variable (otherwise you get vmrun error)
5. now open new command prompt and run "minikube start --vm-driver=vmware"
That's it
{{% /tab %}}
{{% /tabs %}}

## Issues

* [Full list of open 'vmware' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fvmware)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
