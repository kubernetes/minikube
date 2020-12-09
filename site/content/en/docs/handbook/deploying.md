---
title: "Deploying apps"
weight: 2
description: >
  How to deploy an application to minikube
aliases:
  - /docs/tasks/addons
  - /Handbook/addons
---

## kubectl

```shell
kubectl create deployment hello-minikube1 --image=k8s.gcr.io/echoserver:1.4
kubectl expose deployment hello-minikube1 --type=LoadBalancer --port=8080
```

## Addons

minikube has a built-in list of applications and services that may be easily deployed, such as Istio or Ingress. To list the available addons for your version of minikube:

```shell
minikube addons list
```

To enable an add-on, see:
```shell
minikube addons enable <name>
```

To enable an addon at start-up, where *--addons* option can be specified multiple times:

```shell
minikube start --addons <name1> --addons <name2>
```

For addons that expose a browser endpoint, you can quickly open them with:

```shell
minikube addons open <name>
```

To disable an addon:

```shell
minikube addons disable <name>
```
