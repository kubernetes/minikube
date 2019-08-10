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
No documentation is available yet.
{{% /tab %}}
{{% /tabs %}}

## Issues

* [Full list of open 'vmware' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fvmware)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
