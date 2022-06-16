---
title: "hyperv"
weight: 2
aliases:
    - /docs/reference/drivers/hyperv
---
## Overview

[Hyper-V](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/) is a native hypervisor built in to modern versions of Microsoft Windows.

{{% readfile file="/docs/drivers/includes/hyperv_usage.inc" %}}

## Special features

The `minikube start` command supports additional hyperv specific flags:

* **`--hyperv-virtual-switch`**: Name of the virtual switch the minikube VM should use. Defaults to first found
* **`--hyperv-use-external-switch`**: Use external virtual switch over Default Switch if virtual switch not explicitly specified, creates a new one if not found. If the adapter is not specified, the driver first looks up LAN adapters before other adapters (WiFi, ...). Or the user may specify an adapter to attach to the external switch. Default false
* **`--hyperv-external-adapter`**:  External adapter on which the new external switch is created if no existing external switch is found. Since Windows 10 only allows one external switch for the same adapter, it finds the virtual switch before creating one. The external switch is created and named "minikube"

## Issues

Also see [co/hyperv open issues](https://github.com/kubernetes/minikube/labels/co%2Fhyperv)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
* While reinstalling minikube you may encounter error in starting minikube due to stuck.vmcx file from previous installation,a possible fix is:
    * Install [Handle Windows tool](https://docs.microsoft.com/en-us/sysinternals/downloads/handle), identify the process handling .vmcx file,kill it.

    * Run `minikube delete --all --purge` to remove the extra config files
