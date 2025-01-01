---
title: "vfkit"
weight: 2
aliases:
    - /docs/reference/drivers/vfkit
---

## Overview

[VFKit](https://github.com/crc-org/vfkit) is an open-source program for macOS virtualization, optimized for lightweight virtual machines and container deployment.

## Issues

### Other

* [Full list of open 'vfkit' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fvfkit)

## Troubleshooting

### Run with logs

Run `minikube start --driver vfkit --alsologtostderr -v=7` to debug crashes

### Upgrade VFKit

New updates to macOS often require an updated vfkit driver. To upgrade:

* If Podman Desktop is installed, it also bundles `vfkit`
* If you have Brew Package Manager, run: `brew upgrade vfkit`
* As a final alternative, you install the latest VFKit from [GitHub](https://github.com/crc-org/vfkit/releases)
* To check your current version, run: `vfkit -v`
* If the version didn't change after upgrading verify the correct VFKit is in the path. run: `which vfkit`

