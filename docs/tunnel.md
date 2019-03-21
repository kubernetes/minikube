# Minikube Tunnel Design Doc

## Background

Minikube today only exposes a single IP address for all cluster and VM communication.
This effectively requires users to connect to any running Pods, Services or LoadBalancers over ClusterIPs, which can require modifications to workflows when compared to developing against a production cluster.

A main goal of Minikube is to minimize the differences required in code and configuration between development and production, so this is not ideal.
If all cluster IP addresses and Load Balancers were made available on the minikube host machine, these modifications would not be necessary and users would get the "magic" experience of developing from inside a cluster.

Tools like telepresence.io, sshuttle, and the OpenVPN chart provide similar capabilities already.

Also, Steve Sloka has provided a very detailed guide on how to setup a similar configuration [manually](https://stevesloka.com/2017/06/12/access-minikube-service-from-linux-host/).

Elson Rodriguez has provided a similar guide, including a Minikube [external LB controller](https://github.com/elsonrodriguez/minikube-lb-patch).

## Example usage

```shell
$ minikube tunnel
Starting minikube tunnel process. Press Ctrl+C to exit.
All cluster IPs and load balancers are now available from your host machine.
```

## Overview

We will introduce a new command, `minikube tunnel`, that must be run with root permissions.
This command will:

* Establish networking routes from the host into the VM for all IP ranges used by Kubernetes.
* Enable a cluster controller that allocates IPs to services external `LoadBalancer` IPs.
* Clean up routes and IPs when stopped, or when `minikube` stops.

Additionally, we will introduce a Minikube LoadBalancer controller that manages a CIDR of IPs and assigns them to services of type `LoadBalancer`.
These IPs will also be made available on the host machine.

## Network Routes

Minikube drivers usually establish "host-only" IP addresses (192.168.1.1, for example) that route into the running VM
from the host.

The new `minikube tunnel` command will create a static routing table entry that maps the CIDRs used by Pods, Services and LoadBalancers to the host-only IP, obtainable via the `minikube ip` command.

The commands below detail adding routes for the entire `/8` block, we should probably add individual entries for each CIDR we manage instead.

### Linux

Route entries for the entire 10.* block can be added via:

```shell
sudo ip route add 10.0.0.0/8 via $(minikube ip)
```

and deleted via:

```shell
sudo ip route delete 10.0.0.0/8
```

The routing table can be queried with `netstat -nr -f inet`

### OSX

Route entries can be added via:

```shell
sudo route -n add 10.0.0.0/8 $(minikube ip)
```

and deleted via:

```shell
sudo route -n delete 10.0.0.0/8

```

The routing table can be queried with `netstat -nr -f inet`

### Windows

Route entries can be added via:

```shell
route ADD 10.0.0.0 MASK 255.0.0.0 <minikube ip>
```

and deleted via:

```shell
route DELETE 10.0.0.0
```

The routing table can be queried with `route print -4`

### Handling unclean shutdowns

Unclean shutdowns of the tunnel process can result in partially executed cleanup process, leaving network routes in the routing table.
We will keep track of the routes created by each tunnel in a centralized location in the main minikube config directory.
This list serves as a registry for tunnels containing information about

- machine profile
- process ID
- and the route that was created

The cleanup command cleans the routes from both the routing table and the registry for tunnels that are not running:

```shell
minikube tunnel --cleanup
```

Updating the tunnel registry and the routing table is an atomic transaction:

- create route in the routing table + create registry entry if both are successful, otherwise rollback
- delete route in the routing table + remove registry entry if both are successful, otherwise rollback

*Note*: because we don't support currently real multi cluster setup (due to overlapping CIDRs), the handling of running/not-running processes is not strictly required however it is forward looking.

### Handling routing table conflicts

A routing table conflict happens when a destination CIDR of the route required by the tunnel overlaps with an existing route.
Minikube tunnel will warn the user if this happens and should not create the rule.
There should not be any automated removal of conflicting routes.

*Note*: If the user removes the minikube config directory, this might leave conflicting rules in the network routing table that will have to be cleaned up manually.

## Load Balancer Controller

In addition to making IPs routable, minikube tunnel will assign an external IP (the ClusterIP) to all services of type `LoadBalancer`.

The logic of this controller will be, roughly:

```python
for service in services:
  if service.type == "LoadBalancer" and len(service.ingress) == 0:
    add_ip_to_service(ClusterIP, service)
sleep
```

Note that the Minikube ClusterIP can change over time (during system reboots) and this loop should also handle reconciliation of those changes.

## Handling multiple clusters

Multiple clusters are currently not supported due to our inability to specify ServiceCIDR.
This causes conflicting routes having the same destination CIDR.
