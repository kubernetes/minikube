---
title: "vmware"
weight: 6
aliases:
    - /docs/reference/drivers/vmware
---

## Overview

The vmware driver supports virtualization across all VMware based hypervisors.

{{% tabs %}}
{{% mactab %}}
{{% readfile file="/docs/drivers/includes/vmware_macos_usage.inc" %}}
{{% /mactab %}}
{{% linuxtab %}}
No documentation is available yet.
{{% /linuxtab %}}
{{% windowstab %}}
No documentation is available yet.
{{% /windowstab %}}
{{% /tabs %}}

## Issues

* [Full list of open 'vmware-driver' issues](https://github.com/kubernetes/minikube/labels/co%2Fvmware-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
