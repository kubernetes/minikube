---
title: "Ingress DNS"
linkTitle: "Minikube Ingress DNS"
weight: 1
date: 2021-06-03
---
DNS service for ingress controllers running on your minikube server

## Overview

### Problem
When running minikube locally you are highly likely to want to run your services on an ingress controller so that you
don't have to use minikube tunnel or NodePorts to access your services. While NodePort might be ok in a lot of
circumstances in order to test some features an ingress is necessary. Ingress controllers are great because you can
define your entire architecture in something like a helm chart and all your services will be available. There is only
1 problem. That is that your ingress controller works basically off of dns and while running minikube that means that
your local dns names like `myservice.test` will have to resolve to `$(minikube ip)` not really a big deal except the
only real way to do this is to add an entry for every service in your `/etc/hosts` file. This gets messy for obvious
reasons. If you have a lot of services running that each have their own dns entry then you have to set those up
manually. Even if you automate it you then need to rely on the host operating system storing configurations instead of
storing them in your cluster. To make it worse it has to be constantly maintained and updated as services are added,
remove, and renamed. I call it the `/ets/hosts` pollution problem.

### Solution
What if you could just access your local services magically without having to edit your `/etc/hosts` file? Well now you
can. This addon acts as a DNS service that runs inside your kubernetes cluster. All you have to do is install the
service and add the `$(minikube ip)` as a DNS server on your host machine. Each time the dns service is queried an
API call is made to the kubernetes master service for a list of all the ingresses. If a match is found for the name a
response is given with an IP address as the `$(minikube ip)`. So for example lets say my minikube ip address is
`192.168.99.106` and I have an ingress controller with the name of `myservice.test` then I would get a result like so:

```text
#bash:~$ nslookup myservice.test $(minikube ip)
Server:		192.168.99.169
Address:	192.168.99.169#53

Non-authoritative answer:
Name:	myservice.test $(minikube ip)
Address: 192.168.99.169
```

## Installation

### Start minikube
```
minikube start
```

### Install the kubernetes resources
```bash
minikube addons enable ingress-dns
```

### Add the minikube ip as a dns server

#### Mac OS
Create a file in `/etc/resolver/minikube-profilename-test`
```
domain test
nameserver 192.168.99.169
search_order 1
timeout 5
```
Replace `192.168.99.169` with your minikube ip and `profilename` is the name of the minikube profile for the
corresponding ip address

If you have multiple minikube ips you must configure multiple files

See https://www.unix.com/man-page/opendarwin/5/resolver/
Note that even though the `port` feature is documented. It does not actually work.

#### Linux
Update the file `/etc/resolvconf/resolv.conf.d/base` to have the following contents
```
search test
nameserver 192.168.99.169
timeout 5
```
Replace `192.168.99.169` with your minikube ip

If your linux OS uses `systemctl` run the following commands
```bash
sudo resolvconf -u
systemctl disable --now resolvconf.service
```

If your linux does not use `systemctl` run the following commands

TODO add supporting docs for Linux OS that do not use `systemctl`

See https://linux.die.net/man/5/resolver

When you are using Network Manager with the dnsmasq plugin, you can add an additional configuration file, but you need
to restart NetworkManager to activate the change.

```bash
echo "server=/test/$(minikube ip)" >/etc/NetworkManager/dnsmasq.d/minikube.conf
systemctl restart NetworkManager.service
```

Also see `dns=` in [NetworkManager.conf](https://developer.gnome.org/NetworkManager/stable/NetworkManager.conf.html).

#### Windows

TODO

## Testing

### Add the test ingress
```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/minikube/master/deploy/addons/ingress-dns/example/example.yaml
```
Note: Minimum Kubernetes version for example ingress is 1.19

### Validate DNS queries are returning A records
```bash
nslookup hello-john.test $(minikube ip)
nslookup hello-jane.test $(minikube ip)
```

### Validate domain names are resolving on host OS
```bash
ping hello-john.test
```
Expected results:
```text
PING hello-john.test (192.168.99.169): 56 data bytes
64 bytes from 192.168.99.169: icmp_seq=0 ttl=64 time=0.361 ms
```
```bash
ping hello-jane.test
```
```text
PING hello-jane.test (192.168.99.169): 56 data bytes
64 bytes from 192.168.99.169: icmp_seq=0 ttl=64 time=0.262 ms
```

### Curl the example server
```bash
curl http://hello-john.test
```
Expected result:
```text
Hello, world!
Version: 1.0.0
Hostname: hello-world-app-557ff7dbd8-64mtv
```
```bash
curl http://hello-jane.test
```
Expected result:
```text
Hello, world!
Version: 1.0.0
Hostname: hello-world-app-557ff7dbd8-64mtv
```

## Known issues

### .localhost domains will not resolve on chromium
.localhost domains will not correctly resolve on chromium since it is used as a loopback address. Instead use .test, .example, or .invalid

### .local is a reserved TLD
Do not use .local as this is a reserved TLD for mDNS and bind9 DNS servers

### Mac OS

#### mDNS reloading
Each time a file is created or a change is made to a file in `/etc/resolver` you may need to run the following to reload Mac OS mDNS resolver.
```bash
sudo launchctl unload -w /System/Library/LaunchDaemons/com.apple.mDNSResponder.plist
sudo launchctl load -w /System/Library/LaunchDaemons/com.apple.mDNSResponder.plist
```

## TODO
- Add a service that runs on the host OS which will update the files in `/etc/resolver` automatically
- Start this service when running `minikube addons enable ingress-dns` and stop the service when running
  `minikube addons disable ingress-dns`

## Contributors
- [Josh Woodcock](https://github.com/woodcockjosh)

## Images used in this plugin

| Image | Source | Owner |
| :---  | :---   | :---  |
| [ingress-nginx](https://quay.io/repository/kubernetes-ingress-controller/nginx-ingress-controller) | [ingress-nginx](https://github.com/kubernetes/ingress-nginx) | Kubernetes ingress-nginx
| [minikube-ingress-dns](https://hub.docker.com/r/cryptexlabs/minikube-ingress-dns) | [minikube-ingress-dns](https://gitlab.com/cryptexlabs/public/development/minikube-ingress-dns) | Cryptex Labs