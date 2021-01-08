---
title: "generic"
weight: 3
description: >
  Linux generic (remote) driver
aliases:
    - /docs/reference/drivers/generic
---

## Overview

This document is written for system integrators who wish to run minikube within a customized VM environment. The `generic` driver allows advanced minikube users to skip VM creation, allowing minikube to be run on a user-supplied VM.

{{% readfile file="/docs/drivers/includes/generic_usage.inc" %}}

## Issues

* [Full list of open 'generic' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fgeneric-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=4` to debug crashes
