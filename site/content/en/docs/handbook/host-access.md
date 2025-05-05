---
title: "Host access"
date: 2017-01-05
weight: 9
description: >
  How to access host resources from a pod
aliases:
  - docs/tasks/accessing-host-resources/
---
### Prerequisites

The service running on your host must either be bound to all IP's (0.0.0.0) and interfaces, or to the IP and interface your VM is bridged against. If the service is bound only to localhost (127.0.0.1), this will not work.

### `host.minikube.internal`

To make it easier to access your host, minikube v1.10 adds a hostname entry `host.minikube.internal` to `/etc/hosts`. The IP which `host.minikube.internal` resolves to is different across drivers, and may be different across clusters.

### Validating connectivity

You can use `minikube ssh` to confirm connectivity:

```sh
                         _             _
            _         _ ( )           ( )
  ___ ___  (_)  ___  (_)| |/')  _   _ | |_      __  
/' _ ` _ `\| |/' _ `\| || , <  ( ) ( )| '_`\  /'__`\
| ( ) ( ) || || ( ) || || |\`\ | (_) || |_) )(  ___/
(_) (_) (_)(_)(_) (_)(_)(_) (_)`\___/'(_,__/'`\____)

$ ping host.minikube.internal
PING host.minikube.internal (192.168.64.1): 56 data bytes
64 bytes from 192.168.64.1: seq=0 ttl=64 time=0.225 ms
```

To test connectivity to a specific TCP service listening on your host, use `nc -vz host.minikube.internal <port>`:

```sh
$ nc -vz host.minikube.internal 8000
Connection to host.minikube.internal 8000 port [tcp/*] succeeded!
```

Here are how to interpret the different messages:
* `Connection succeeded`: You are connected!
* `Connection refused`: the service is not listening on the port, at least not across all interfaces

{{% alert title="Note" color="primary" %}}
When using an older version of minikube, you may have to manually install tools like `ping` and `netcat` within the minikube image:
```sh
sudo apt install iputils-ping netcat-openbsd
```
{{% /alert %}}
