---
title: "How to use custom TLS certificate with ingress addon"
linkTitle: "Using Custom TLS certificate with Ingress Addon"
weight: 1
date: 2020-11-30
---

## Overview

- This tutorial will show you how to configure custom TLS certificatate for ingress addon.  
- [mkcert](https://github.com/FiloSottile/mkcert) is a simple tool for making locally-trusted development certificates. It requires no configuration.

## Tutorial

- Start minikube
```shell
$ minikube start
```

- Create TLS secret which contains custom certificate and private key
```shell
$ kubectl -n kube-system create secret tls mkcert --key key.pem --cert cert.pem
```

- Configure ingress addon
```shell
$ minikube addons configure ingress
-- Enter custom cert(format is "namespace/secret"): kube-system/mkcert
✅  ingress was successfully configured
```

- Enable ingress addon (disable first when already enabled)
```shell
$ minikube addons disable ingress
🌑  "The 'ingress' addon is disabled

$ minikube addons enable ingress
🔎  Verifying ingress addon...
🌟  The 'ingress' addon is enabled
```
- Verify if custom certificate was enabled
```shell
$ kubectl -n ingress-nginx get deployment ingress-nginx-controller -o yaml | grep "kube-system"
- --default-ssl-certificate=kube-system/mkcert
```
