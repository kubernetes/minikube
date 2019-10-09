---
title: "Insecure"
linkTitle: "Insecure"
weight: 6
date: 2019-08-1
description: >
  How to enable insecure registry support within minikube
---

minikube allows users to configure the docker engine's `--insecure-registry` flag. 

You can use the `--insecure-registry` flag on the
`minikube start` command to enable insecure communication between the docker engine and registries listening to requests from the CIDR range.

One nifty hack is to allow the kubelet running in minikube to talk to registries deployed inside a pod in the cluster without backing them
with TLS certificates. Because the default service cluster IP is known to be available at 10.0.0.1, users can pull images from registries
deployed inside the cluster by creating the cluster with `minikube start --insecure-registry "10.0.0.0/24"`.
