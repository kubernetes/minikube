---
title: "How to use custom TLS certificate with ingress addon"
linkTitle: "Using custom TLS certificate with ingress addon"
weight: 1
date: 2020-11-30
---

## Overview

- This tutorial will show you how to configure custom TLS certificatate for ingress addon.

## Tutorial

- Start minikube
```
$ minikube start
```

- Create TLS secret which contains custom certificate and private key
```
$ kubectl -n kube-system create secret tls mkcert --key key.pem --cert cert.pem
```

- Configure ingress addon
```
$ minikube addons configure ingress
-- Enter custom cert(format is "namespace/secret"): kube-system/mkcert
✅  ingress was successfully configured
```

- Enable ingress addon (disable first when already enabled)
```
$ minikube addons disable ingress
🌑  "The 'ingress' addon is disabled

$ minikube addons enable ingress
🔎  Verifying ingress addon...
🌟  The 'ingress' addon is enabled
```
- Verify if custom certificate was enabled
```
$ kubectl -n ingress-nginx get deployment ingress-nginx-controller -o yaml | grep "kube-system"
- --default-ssl-certificate=kube-system/mkcert
```
