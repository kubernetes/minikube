---
title: "Host access"
date: 2017-01-05
weight: 9
description: >
  How to access host resources from a pod
aliases:
  - docs/tasks/accessing-host-resources/
---

{{% pageinfo %}}
This has only been tested on VirtualBox and Hyperkit. Instructions may differ for other VM drivers.
{{% /pageinfo %}}

### Prerequisites

The service running on your host must either be bound to all IP's (0.0.0.0) and interfaces, or to the IP and interface your VM is bridged against. If the service is bound only to localhost (127.0.0.1), this will not work.

### Getting the bridge IP

To access a resource, such as a MySQL service running on your host, from inside of a pod, the key information you need is the correct target IP to use.

```shell
minikube ssh "route -n | grep ^0.0.0.0 | awk '{ print \$2 }'"
```

Example output:

```
10.0.2.2
```

This is the IP your pods can connect to in order to access services running on the host.

### Validating connectivity

```shell
minikube ssh
telnet <ip> <port>
```

If you press enter, you may see an interesting string or error message pop up. This means that the connection is working.
