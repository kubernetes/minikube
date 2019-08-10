---
title: "parallels"
linkTitle: "parallels"
weight: 4
date: 2018-08-08
description: >
  Parallels driver
---

## Overview

The Parallels driver is particularly useful for users who own Parallels Desktop, as it does not require VT-x hardware support.

{{% readfile file="/docs/Reference/Drivers/includes/parallels_usage.inc" %}}

## Issues

* [Full list of open 'parallels' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fparallels)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=7` to debug crashes
