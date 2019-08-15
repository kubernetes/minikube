---
title: "NodePort access"
linkTitle: "NodePort access"
weight: 6
date: 2018-08-02
description: >
  How to access a NodePort service in minikube
---

A NodePort service is the most basic way to get external traffic directly to your service. NodePort, as the name implies, opens a specific port, and any traffic that is sent to this port is forwarded to the service.

### Getting the NodePort using the service command

We also have a shortcut for fetching the minikube IP and a service's `NodePort`:

`minikube service --url $SERVICE`

## Getting the NodePort using kubectl

The minikube VM is exposed to the host system via a host-only IP address, that can be obtained with the `minikube ip` command. Any services of type `NodePort` can be accessed over that IP address, on the NodePort.

To determine the NodePort for your service, you can use a `kubectl` command like this (note that `nodePort` begins with lowercase `n` in JSON output):

`kubectl get service $SERVICE --output='jsonpath="{.spec.ports[0].nodePort}"'`

### Increasing the NodePort range

By default, minikube only exposes ports 30000-32767. If this does not work for you, you can adjust the range by using:

`minikube start --extra-config=apiserver.service-node-port-range=1-65535`

This flag also accepts a comma separated list of ports and port ranges.

