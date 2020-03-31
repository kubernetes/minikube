---
title: "Sudo prompts"
linkTitle: "Sudo prompts"
weight: 1
date: 2020-03-26
description: >
  Disabling sudo prompts when using minikude start/stop/status, kubectl cluster-info, ... 
---

## Use the `docker` driver

Use the `docker` driver rather than the `none` driver. `docker` driver should be used unless it does not meet requirements for some reason.

## For `none` users

For `none` users, `CHANGE_MINIKUBE_NONE_USER=true`, kubectl and such will still work: [see environment variables](https://minikube.sigs.k8s.io/docs/reference/environment_variables/)

## Otherwise deal with `sudo`

Configure `sudo` to never prompt for the commands issued by minikube.
