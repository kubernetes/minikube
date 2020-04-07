
---
title: "Basic controls"
weight: 2
description: >
  See minikube in action!
---

Start a cluster by running:

`minikube start`

Access the Kubernetes Dashboard running within the minikube cluster:

`minikube dashboard`

Once started, you can interact with your cluster using `kubectl`, just like any other Kubernetes cluster. For instance, starting a server:

`kubectl create deployment hello-minikube --image=k8s.gcr.io/echoserver:1.4`

Exposing a service as a NodePort

`kubectl expose deployment hello-minikube --type=NodePort --port=8080`

minikube makes it easy to open this exposed endpoint in your browser:

`minikube service hello-minikube`

Upgrade your cluster:

`minikube start --kubernetes-version=latest`

Start a second local cluster (_note: This will not work if minikube is using the bare-metal/none driver_):

`minikube start -p cluster2`

Stop your local cluster:

`minikube stop`

Delete your local cluster:

`minikube delete`

Delete all local clusters and profiles

`minikube delete --all`
