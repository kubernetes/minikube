---
title: "hyperkit"
linkTitle: "hyperkit"
weight: 1
date: 2018-08-08
description: >
  HyperKit driver
---

## Overview

[HyperKit](https://github.com/moby/hyperkit) is an open-source hypervisor for macOS hypervisor, optimized for lightweight virtual machines and container deployment.

{{% readfile file="/docs/Reference/Drivers/includes/hyperkit_usage.inc" %}}

## Special features

minikube start supports additional hyperkit specific flags:

* **`--hyperkit-vpnkit-sock`**: Location of the VPNKit socket used for networking. If empty, disables Hyperkit VPNKitSock, if 'auto' uses Docker for Mac VPNKit connection, otherwise uses the specified VSock
* **`--hyperkit-vsock-ports`**: List of guest VSock ports that should be exposed as sockets on the host
* **`--nfs-share`**: Local folders to share with Guest via NFS mounts
* **`--nfs-shares-root`**: Where to root the NFS Shares (default "/nfsshares")
* **`--uuid`**: Provide VM UUID to restore MAC address

## Issues

### Local DNS server conflict

If you are using `dnsmasq` and `minikube` fails, add `listen-address=192.168.64.1` to dnsmasq.conf.

If you are running other DNS servers, shut them off or specify an alternative bind address.

### Other

* [Full list of open 'hyperkit' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fhyperkit)

## Troubleshooting

* Run `docker-machine-driver-hyperkit version` to make sure the version matches minikube
* Run `minikube start --alsologtostderr -v=7` to debug crashes
