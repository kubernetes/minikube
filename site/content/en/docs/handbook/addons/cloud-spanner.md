---
title: "Using the Cloud Spanner Addon"
linkTitle: "Cloud Spanner"
weight: 1
date: 2022-10-14
---

## Cloud Spanner Addon

[Cloud Spanner](https://cloud.google.com/spanner) is a fully managed relational database. The Cloud Spanner addon provides a local emulator to test your local application without incurring the cost of an actual spanner instance. 

### Enable Cloud Spanner on minikube

To enable this addon, simply run:
```shell script
minikube addons enable cloud-spanner
```

### Cloud Spanner Endpoints
Cloud Spanner provides two different ports, HTTP and GRPC. List Cloud Spanner emulator urls by running:
``` shell
minikube service cloud-spanner-emulator

####################Sample Output#########################
|-----------|------------------------|-------------|---------------------------|
| NAMESPACE |          NAME          | TARGET PORT |            URL            |
|-----------|------------------------|-------------|---------------------------|
| default   | cloud-spanner-emulator | http/9020   | http://192.168.49.2:30233 |
|           |                        | grpc/9010   | http://192.168.49.2:30556 |
|-----------|------------------------|-------------|---------------------------|
[default cloud-spanner-emulator http/9020
grpc/9010 http://192.168.49.2:30233
http://192.168.49.2:30556]
```

### Using Cloud Spanner within a cluster
Cloud Spanner emulator can be used via endpoint `cloud-spanner-emulator:9020` for http clients and `cloud-spanner-emulator:9010` for grpc clients respectively. If you're using the standard client library for Cloud Spanner then set `SPANNER_EMULATOR_HOST` to the GRPC endpoint `cloud-spanner-emulator:9010`.

### Testing installation

```shell script
kubectl get pods -n cloud-spanner-emulator
```

If everything went well, there should be no errors about Cloud Spanner's installation in your minikube cluster.

### Disable Cloud Spanner

To disable this addon, simply run:

```shell script
minikube addons disable cloud-spanner
```
