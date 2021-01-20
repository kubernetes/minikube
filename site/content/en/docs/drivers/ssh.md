---
title: "ssh"
weight: 3
description: >
  Linux ssh (remote) driver
aliases:
    - /docs/reference/drivers/ssh
---

## Overview

This document is written for system integrators who wish to run minikube within a customized VM environment. The `ssh` driver allows advanced minikube users to skip VM creation, allowing minikube to be run on a user-supplied VM.

{{% readfile file="/docs/drivers/includes/ssh_usage.inc" %}}

## Issues

* [Full list of open 'ssh' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fssh-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=4` to debug crashes
