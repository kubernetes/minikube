---
title: "Using Static Token file"
linkTitle: "Using Static Token file"
weight: 1
date: 2021-08-04
description: >
  Using a static token file in Minikube
---

## Overview

A [static token file](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#static-token-file) can be used to ensure only authenticated users access the API server. As minikube nodes are run in VMs/containers, this adds a complication to ensuring this token file is accessible to the node. This tutorial explains how to configure a static token file.

## Tutorial

This must be done before creating the minikube cluster.

```shell
# Create the folder that will be copied into the control plane.
mkdir -p ~/.minikube/files/etc/ca-certificates/

# Copy the token file into the folder.
cp token.csv ~/.minikube/files/etc/ca-certificates/token.csv

# Start minikube with the token auth file specified.
minikube start \
  --extra-config=apiserver.token-auth-file=/etc/ca-certificates/token.csv
```

Placing files in `~/.minikube/files/` automatically copies them to the specified path in each minikube node. This means once we call `minikube start`, it is able to access the token file since it is locally present in the node.
