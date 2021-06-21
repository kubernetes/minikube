---
title: "Proxies and VPNs"
weight: 6
description: >
  How to use minikube with a VPN or HTTP/HTTPS Proxy
aliases:
  - /docs/reference/networking/vpn
  - /docs/reference/networking/proxy
---

minikube requires access to the internet via HTTP, HTTPS, and DNS protocols.

## Proxy

If a HTTP proxy is required to access the internet, you may need to pass the proxy connection information to both minikube and Docker using environment variables:

* `HTTP_PROXY` - The URL to your HTTP proxy
* `HTTPS_PROXY` - The URL to your HTTPS proxy
* `NO_PROXY` - A comma-separated list of hosts which should not go through the proxy.

The NO_PROXY variable here is important: Without setting it, minikube may not be able to access resources within the VM. minikube uses two IP ranges, which should not go through the proxy:

* **192.168.99.0/24**: Used by the minikube VM. Configurable for some hypervisors via `--host-only-cidr`
* **192.168.39.0/24**: Used by the minikube kvm2 driver.
* **192.168.49.0/24**: Used by the minikube docker driver's first cluster.
* **10.96.0.0/12**: Used by service cluster IP's. Configurable via  `--service-cluster-ip-range`

One important note: If NO_PROXY is required by non-Kubernetes applications, such as Firefox or Chrome, you may want to specifically add the minikube IP to the comma-separated list, as they may not understand IP ranges ([#3827](https://github.com/kubernetes/minikube/issues/3827)).

## Example Usage

### macOS and Linux

```shell
export HTTP_PROXY=http://<proxy hostname:port>
export HTTPS_PROXY=https://<proxy hostname:port>
export NO_PROXY=localhost,127.0.0.1,10.96.0.0/12,192.168.99.0/24,192.168.39.0/24

minikube start
```

To make the exported variables permanent, consider adding the declarations to ~/.bashrc or wherever your user-set environment variables are stored.

### Windows

```shell
set HTTP_PROXY=http://<proxy hostname:port>
set HTTPS_PROXY=https://<proxy hostname:port>
set NO_PROXY=localhost,127.0.0.1,10.96.0.0/12,192.168.99.0/24,192.168.39.0/24

minikube start
```

To set these environment variables permanently, consider adding these to your [system settings](https://support.microsoft.com/en-au/help/310519/how-to-manage-environment-variables-in-windows-xp) or using [setx](https://stackoverflow.com/questions/5898131/set-a-persistent-environment-variable-from-cmd-exe)

### Troubleshooting

#### unable to cache ISO... connection refused

```text
Unable to start VM: unable to cache ISO: https://storage.googleapis.com/minikube/iso/minikube.iso:
failed to download: failed to download to temp file: download failed: 5 error(s) occurred:

* Temporary download error: Get https://storage.googleapis.com/minikube/iso/minikube.iso:
proxyconnect tcp: dial tcp <host>:<port>: connect: connection refused
```

This error indicates that the host:port combination defined by HTTPS_PROXY or HTTP_PROXY is incorrect, or that the proxy is unavailable.

#### Unable to pull images..Client.Timeout exceeded while awaiting headers

```text
Unable to pull images, which may be OK:

failed to pull image "k8s.gcr.io/kube-apiserver:v1.13.3": output: Error response from daemon:
Get https://k8s.gcr.io/v2/: net/http: request canceled while waiting for connection
(Client.Timeout exceeded while awaiting headers)
```

This error indicates that the container runtime running within the VM does not have access to the internet. Verify that you are passing the appropriate value to `--docker-env HTTPS_PROXY`.

#### x509: certificate signed by unknown authority

```text
[ERROR ImagePull]: failed to pull image k8s.gcr.io/kube-apiserver:v1.13.3:
output: Error response from daemon:
Get https://k8s.gcr.io/v2/: x509: certificate signed by unknown authority
```

This is because minikube VM is stuck behind a proxy that rewrites HTTPS responses to contain its own TLS certificate. The [solution](https://github.com/kubernetes/minikube/issues/3613#issuecomment-461034222) is to install the proxy certificate into a location that is copied to the VM at startup, so that it can be validated.

Ask your IT department for the appropriate PEM file, and add it to:

`~/.minikube/files/etc/ssl/certs`

or

`~/.minikube/certs`

Then run `minikube delete` and `minikube start`.

#### downloading binaries: proxyconnect tcp: tls: oversized record received with length 20527

The supplied value of `HTTPS_PROXY` is probably incorrect. Verify that this value is not pointing to an HTTP proxy rather than an HTTPS proxy.

## VPN

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
