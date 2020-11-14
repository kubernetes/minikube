---
title: "Using Ambassador Ingress Controller"
linkTitle: "Using Ambassador Ingress Controller"
weight: 1
date: 2020-05-14
description: >
  Using Ambassador Ingress Controller with Minikube
---

## Overview

[Ambassador](https://getambassador.io/) allows access to Kubernetes services running inside Minikube. Ambassador can be
configured via both, [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) resources and
[Mapping](https://www.getambassador.io/docs/latest/topics/using/intro-mappings/) resources.

## Prerequisites

- Minikube version higher than v1.10.1
- kubectl

## Configuring Ambassador

### Installing Ambassador

Ambassador is available as a Minikube
[addon]({{< ref "/docs/commands/addons" >}}) Install it by running -

```shell script
minikube addons enable ambassador
```

This will install Ambassador in the namespace `ambassador`.

### Accessing Ambassador via `minikube tunnel`

The service `ambassador` is of type `LoadBalancer`. To access this service, run a
[Minikube tunnel](https://minikube.sigs.k8s.io/docs/handbook/accessing/#using-minikube-tunnel) in a separate terminal.

```shell script
minikube tunnel
```

You can now access Ambassador at the external IP allotted to the `ambassador` service.
Get the external IP with the following command:
```shell script
kubectl get service ambassador -n ambassador
NAME         TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)                      AGE
ambassador   LoadBalancer   10.104.86.124   10.104.86.124   80:31287/TCP,443:31934/TCP   77m
```

### Configuring via Ingress resource

In this tutorial, we'll configure Ambassador via an Ingress resource. To configure via `IngressClass` resource, read
this [post](https://blog.getambassador.io/new-kubernetes-1-18-extends-ingress-c34abdc2f064).

First, let's create a Kubernetes deployment and service which we will talk to via Ambassador.

```shell script
kubectl create deployment hello-minikube --image=k8s.gcr.io/echoserver:1.4
kubectl expose deployment hello-minikube --port=8080
```

This service `hello-minikube` is of type `ClusterIP` and is not accessible from outside the cluster.

Now, create an Ingress resource which exposes this service at the path `/hello/`

**Note:** The Ingress resource must have the annotation `kubernetes.io/ingress.class: ambassador` for Ambassador to
pick it up.

`hello-ingress.yaml`
```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: ambassador
  name: test-ingress
spec:
  rules:
  - http:
      paths:
      - path: /hello/
        backend:
          serviceName: hello-minikube
          servicePort: 8080
```
Run the command:
```shell
kubectl apply -f hello-ingress.yaml
```

That's it! You can now access your service via Ambassador:
```shell script
curl http://<Ambassdor's External IP'/hello/>
```

**Note:** For more advanced Ingress configurations with Ambassador, like TLS termination and name-based virtual hosting,
see Ambassador's [documentation](https://www.getambassador.io/docs/latest/topics/running/ingress-controller/).

### Configuring via Mapping resource

While Ambassador understands the Ingress spec, the Ingress spec does not leverage all of Ambassador's features. The
[Mapping](https://www.getambassador.io/docs/latest/topics/using/intro-mappings/) resource is Ambassador's core resource
that maps a target backend service to a given host or prefix.

Let's create another Kubernetes deployment and service that we will expose via Ambassador -
```shell script
kubectl create deployment mapping-minikube --image=k8s.gcr.io/echoserver:1.4
kubectl expose deployment mapping-minikube --port=8080
```

This service `mapping-minikube` is of type `ClusterIP` and is not accessible from outside the cluster.

Now, let's create a mapping that exposes this service via Ambassador at the path `/hello-mapping/`

`hello-mapping.yaml`
```yaml
apiVersion: getambassador.io/v2
kind:  Mapping
metadata:
  name:  mapping-minikube
spec:
  prefix: /hello-mapping/
  service: mapping-minikube.default:8080
```
Run the command:
```shell
kubectl apply -f hello-mapping.yaml
```

That's it! You can now access your service via Ambassador:
```shell script
curl http://<Ambassdor's External IP'/hello-mapping/>
```

**Note:** Read more about mappings in Ambassador's
[documentation](https://www.getambassador.io/docs/latest/topics/using/mappings/).
