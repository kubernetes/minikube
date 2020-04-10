---
title: "FAQ"
linkTitle: "FAQ"
weight: 3
description: >
  Questions that come up regularly
---

## Operating-systems

## Linux

### Preventing password prompts

The easiest approach is to use the `docker` driver, as the backend service always runs as `root`.

`none` users may want to try `CHANGE_MINIKUBE_NONE_USER=true`,  where kubectl and such will still work: [see environment variables](https://minikube.sigs.k8s.io/reference/environment_variables/)

Alternatively, configure `sudo` to never prompt for the commands issued by minikube.
