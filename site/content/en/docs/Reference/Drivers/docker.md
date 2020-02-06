---
title: "docker"
linkTitle: "docker"
weight: 3
date: 2020-02-05
description: >
  Docker driver (EXPERIMENTAL)
---

## Overview

The Docker driver is an experimental VM-free driver that ships with minikube v1.7.

This driver was inspired by the [kind project](https://kind.sigs.k8s.io/), and uses a modified version of its base image.

## Special features

No hypervisor required when run on Linux.

## Limitations

As an experimental driver, not all commands are supported on all platforms. Notably: `mount,` `service`, `tunnel`, and others. Most of these limitations will be addressed by minikube v1.8 (March 2020)

## Issues

* [Full list of open 'kic-driver' issues](https://github.com/kubernetes/minikube/labels/co%2Fkic-driver)

## Troubleshooting

* Run `minikube start --alsologtostderr -v=1` to debug crashes
