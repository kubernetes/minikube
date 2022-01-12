---
title: "ExternalDNS Addon"
linkTitle: "ExternalDNS Addon"
weight: 1
date: 2022-01-12
---
DNS service for ingress controllers, services of type `LoadBalancer`, Istio Gateways &
VirtualServices and much more.

## Overview

When running minikube locally, you may want to:

- access your services using an ingress controller
- access services directly using type `LoadBalancer`
- access your services using Istio Gateway and VirtualService resources
- or something similar

On minikube, this is quite challenging, because you have to manually edit your hosts file for every
hostname you want to resolve to your minikube IP or LoadBalancer IP. 

### Solution

This addon deploys a DNS server as well as an [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) 
instance inside your minikube Kubernetes cluster. ExternalDNS observes the Kubernetes API server for
Ingress, Service and Istio resources and updates the DNS server accordingly. All you have to do is
add the `minikube ip` as a DNS server on your host machine.

## Installation

<h3 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">1</strong></span>Start minikube</h2>

```bash
minikube start
```

<h3 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">2</strong></span>(optional) Configure the addon</h2>

You can only configure hostnames ending with `.demo` because the DNS zone is set to
`demo.` by default. If you want to change the DNS zone, you can do so by configuring the addon.

```bash
minikube addons configure external-dns
```

<h3 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">3</strong></span>Enable the addon</h2>

```bash
minikube addons enable external-dns
```

<h3 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">4</strong></span>Add the `minikube ip` as a DNS server</h2>

{{% tabs %}}
{{% linuxtab %}}

## Linux OS with resolvconf

Update the file `/etc/resolvconf/resolv.conf.d/base` to have the following contents.

```
search demo
nameserver 192.168.99.169
timeout 5
```

Replace `192.168.99.169` with your `minikube ip`.

If your Linux OS uses `systemctl`, run the following commands.

```bash
sudo resolvconf -u
systemctl disable --now resolvconf.service
```

See https://linux.die.net/man/5/resolver

## Linux OS with Network Manager

Network Manager can run integrated caching DNS server - `dnsmasq` plugin and can be configured to use separate nameservers per domain.

Edit /etc/NetworkManager/NetworkManager.conf and set `dns=dnsmasq`

```
[main]
dns=dnsmasq
```
Also see `dns=` in [NetworkManager.conf](https://developer.gnome.org/NetworkManager/stable/NetworkManager.conf.html).

Configure dnsmasq to handle .test domain

```bash
sudo mkdir /etc/NetworkManager/dnsmasq.d/
echo "server=/test/$(minikube ip)" >/etc/NetworkManager/dnsmasq.d/minikube.conf
```

Restart Network Manager
```
systemctl restart NetworkManager.service
```
Ensure your /etc/resolv.conf  contains only single nameserver
```bash
cat /etc/resolv.conf | grep nameserver
nameserver 127.0.0.1
```

{{% /linuxtab %}}

{{% mactab %}}

Create a file in `/etc/resolver/minikube-test` with the following contents.

```
domain demo
nameserver 192.168.99.169
search_order 1
timeout 5
```

Replace `192.168.99.169` with your `minikube ip`.

If you have multiple minikube IPs, you must configure a file for each.

See https://www.unix.com/man-page/opendarwin/5/resolver/

Note that the `port` feature does not work as documented.

{{% /mactab %}}

{{% windowstab %}}

Open `Powershell` as Administrator and execute the following.
```sh
Add-DnsClientNrptRule -Namespace ".test" -NameServers "$(minikube ip)"
```

The following will remove any matching rules before creating a new one. This is useful for updating the `minikube ip`.
```sh
Get-DnsClientNrptRule | Where-Object {$_.Namespace -eq '.demo'} | Remove-DnsClientNrptRule -Force; Add-DnsClientNrptRule -Namespace ".demo" -NameServers "$(minikube ip)"
```

{{% /windowstab %}}
{{% /tabs %}}

<h3 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">5</strong></span>(optional) Configure in-cluster DNS server to resolve local DNS names inside cluster</h2>

Sometimes it's useful to access other applications inside cluster via ingress and by their local DNS
name - microservices/APIs/tests. 
In such case the DNS server of the external-dns addon should be used by the in-cluster DNS server -
[CoreDNS](https://coredns.io/) to resolve local DNS names.

Edit your CoreDNS config
```sh
kubectl edit configmap coredns -n kube-system
```
and add block for your local domain
```
    demo:53 {
            errors
            cache 30
            forward . 192.168.99.169
    }

```
Replace `192.168.99.169` with your `minikube ip`.

The final ConfigMap should look like:
```yaml
apiVersion: v1
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
...
    }
    demo:53 {
            errors
            cache 30
            forward . 192.168.99.169
    }
kind: ConfigMap
metadata:
...
```

See https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/
