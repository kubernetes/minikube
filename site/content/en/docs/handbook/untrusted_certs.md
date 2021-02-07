---
title: "Certificates"
weight: 7
date: 2019-08-15
description: >
  All about TLS certificates
---

## Untrusted Root Certificates

Many organizations deploy their own Root Certificate and CA service inside the corporate networks.
Internal websites, image repositories and other resources may install SSL server certificates issued by this CA service for security and privacy concerns.

You may install the Root Certificate into the minikube cluster to access these corporate resources within the cluster.

### Tutorial

You will need a corporate X.509 Root Certificate in PEM format. If it's in DER format, convert it:

```shell
openssl x509 -inform der -in my_company.cer -out my_company.pem
```

Copy the certificate into the certs directory:

```shell
mkdir -p $HOME/.minikube/certs
cp my_company.pem $HOME/.minikube/certs/my_company.pem
```

Then restart minikube with the `--embed-certs` flag to sync the certificates:

```shell
minikube start --embed-certs
```
