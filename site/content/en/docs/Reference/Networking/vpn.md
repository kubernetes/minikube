---
title: "Host VPN"
linkTitle: "Host VPN"
weight: 6
date: 2019-08-01
description: >
  Using minikube on a host with a VPN installed
---

minikube requires access from the host to the following IP ranges:

* **192.168.99.0/24**: Used by the minikube VM. Configurable for some hypervisors via `--host-only-cidr`
* **192.168.39.0/24**: Used by the minikube kvm2 driver.
* **10.96.0.0/12**: Used by service cluster IP's. Configurable via  `--service-cluster-ip-range`

Unfortunately, many VPN configurations route packets to these destinations through an encrypted tunnel, rather than allowing the packets to go to the minikube VM. 

### Possible workarounds

1. If you have access, whitelist the above IP ranges in your VPN software
2. In your VPN software, select an option similar to "Allow local (LAN) access when using VPN" [(Cisco VPN example)](https://superuser.com/questions/987150/virtualbox-guest-os-through-vpn)
3. You may have luck selecting alternate values to the `--host-only-cidr` and `--service-cluster-ip-range` flags.
4. Turn off the VPN
