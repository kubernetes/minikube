---
title: "parallels"
weight: 4
aliases:
    - /docs/reference/drivers/parallels
---

## Overview

The Parallels driver is particularly useful for users who own Parallels Desktop for Mac, as it does not require VT-x hardware support.

{{% readfile file="/docs/drivers/includes/parallels_usage.inc" %}}

## Issues

* [Full list of open 'parallels' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fparallels)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
