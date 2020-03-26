---
title: "Untrusted Root Certificate"
linkTitle: "Untrusted Root Certificate"
weight: 1
date: 2019-08-15
description: >
  Using minikube with Untrusted Root Certificate
---

## Overview

Most organizations deploy their own Root Certificate and CA service inside the corporate networks.
Internal websites, image repositories and other resources may install SSL server certificates issued by this CA service for security and privacy concerns.
You may install the Root Certificate into the minikube cluster to access these corporate resources within the cluster.

## Prerequisites

- Corporate X.509 Root Certificate
- Latest minikube binary and ISO

## Tutorial

* The certificate must be in PEM format. You may use `openssl` to convert from DER format.

```
openssl x509 -inform der -in my_company.cer -out my_company.pem
```

* You may need to delete existing minikube cluster

```shell
minikube delete
```

* Copy the certificate before creating the minikube cluster

```shell
mkdir -p $HOME/.minikube/certs
cp my_company.pem $HOME/.minikube/certs/my_company.pem

minikube start
```
