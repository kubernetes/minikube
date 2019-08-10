---
title: "LoadBalancer access"
linkTitle: "LoadBalancer access"
weight: 6
date: 2018-08-02
description: >
  How to access a LoadBalancer service in minikube
---

## Overview

A LoadBalancer service is the standard way to expose a service to the internet. With this method, each service gets it's own IP address.


## Using `minikube tunnel`

Services of type `LoadBalancer` can be exposed via the `minikube tunnel` command. It will run until Ctrl-C is hit.

````shell
minikube tunnel
````
Example output:

```text
out/minikube tunnel
Password: *****
Status:
        machine: minikube
        pid: 59088
        route: 10.96.0.0/12 -> 192.168.99.101
        minikube: Running
        services: []
    errors:
                minikube: no errors
                router: no errors
                loadbalancer emulator: no errors
```


`minikube tunnel` runs as a separate daemon, creating a network route on the host to the service CIDR of the cluster using the cluster's IP address as a gateway.  The tunnel command exposes the external IP directly to any program running on the host operating system.

### DNS resolution (experimental)

If you are on macOS, the tunnel command also allows DNS resolution for Kubernetes services from the host.

### Cleaning up orphaned routes

If the `minikube tunnel` shuts down in an abrupt manner, it may leave orphaned network routes on your system. If this happens, the ~/.minikube/tunnels.json file will contain an entry for that tunnel. To remove orphaned routes, run:

````shell
minikube tunnel --cleanup
````

### Avoiding password prompts

Adding a route requires root privileges for the user, and thus there are differences in how to run `minikube tunnel` depending on the OS. If you want to avoid entering the root password, consider setting NOPASSWD for "ip" and "route" commands:

<https://superuser.com/questions/1328452/sudoers-nopasswd-for-single-executable-but-allowing-others>
