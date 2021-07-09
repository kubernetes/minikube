---
title: "Using Multi-Node Clusters"
linkTitle: "Using multi-node clusters"
weight: 1
date: 2019-11-24
---

## Overview

- This tutorial will show you how to start a multi-node clusters on minikube and deploy a service to it.

## Prerequisites

- minikube 1.10.1 or higher
- kubectl

## Tutorial

- Start a cluster with 2 nodes in the driver of your choice:

```shell
minikube start --nodes 2 -p multinode-demo
```
```
😄  [multinode-demo] minikube v1.18.1 on Opensuse-Tumbleweed 
✨  Automatically selected the docker driver
👍  Starting control plane node multinode-demo in cluster multinode-demo
🔥  Creating docker container (CPUs=2, Memory=8000MB) ...
🐳  Preparing Kubernetes v1.20.2 on Docker 20.10.3 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔗  Configuring CNI (Container Networking Interface) ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass

👍  Starting node multinode-demo-m02 in cluster multinode-demo
🔥  Creating docker container (CPUs=2, Memory=8000MB) ...
🌐  Found network options:
    ▪ NO_PROXY=192.168.49.2
🐳  Preparing Kubernetes v1.20.2 on Docker 20.10.3 ...
    ▪ env NO_PROXY=192.168.49.2
🔎  Verifying Kubernetes components...
🏄  Done! kubectl is now configured to use "multinode-demo" cluster and "default" namespace by default
```

- Get the list of your nodes:

```shell
kubectl get nodes
```
```
NAME                 STATUS   ROLES                  AGE   VERSION
multinode-demo       Ready    control-plane,master   99s   v1.20.2
multinode-demo-m02   Ready    <none>                 73s   v1.20.2
```

- You can also check the status of your nodes:

```shell
minikube status -p multinode-demo
```

```
multinode-demo
type: Control Plane
host: Running
kubelet: Running
apiserver: Running
kubeconfig: Configured

multinode-demo-m02
type: Worker
host: Running
kubelet: Running
```

- Deploy our hello world deployment:

```shell
kubectl apply -f hello-deployment.yaml
```
```
deployment.apps/hello created
```
```shell
kubectl rollout status deployment/hello
```
```
deployment "hello" successfully rolled out
```

- Deploy our hello world service, which just spits back the IP address the request was served from:

```shell
kubectl apply -f hello-svc.yaml
```
```
service/hello created
```

- Check out the IP addresses of our pods, to note for future reference

```shell
kubectl get pods -o wide
```
```
NAME                     READY   STATUS    RESTARTS   AGE   IP           NODE                 NOMINATED NODE   READINESS GATES
hello-695c67cf9c-bzrzk   1/1     Running   0          22s   10.244.1.2   multinode-demo-m02   <none>           <none>
hello-695c67cf9c-frcvw   1/1     Running   0          22s   10.244.0.3   multinode-demo       <none>           <none>
```

- Look at our service, to know what URL to hit

```shell
minikube service list -p multinode-demo
```
```
|-------------|------------|--------------|---------------------------|
|  NAMESPACE  |    NAME    | TARGET PORT  |            URL            |
|-------------|------------|--------------|---------------------------|
| default     | hello      |           80 | http://192.168.49.2:31000 |
| default     | kubernetes | No node port |                           |
| kube-system | kube-dns   | No node port |                           |
|-------------|------------|--------------|---------------------------|
```

- Let's hit the URL a few times and see what comes back

```shell
curl  http://192.168.49.2:31000
```
```
Hello from hello-695c67cf9c-frcvw (10.244.0.3)

curl  http://192.168.49.2:31000
Hello from hello-695c67cf9c-bzrzk (10.244.1.2)

curl  http://192.168.49.2:31000
Hello from hello-695c67cf9c-bzrzk (10.244.1.2)

curl  http://192.168.49.2:31000
Hello from hello-695c67cf9c-frcvw (10.244.0.3)
```

- Multiple nodes!

- Referenced YAML files
{{% tabs %}}
{{% tab hello-deployment.yaml %}}

```
{{% readfile file="/docs/tutorials/includes/hello-deployment.yaml" %}}
```
{{% /tab %}}
{{% tab hello-svc.yaml %}}
```
{{% readfile file="/docs/tutorials/includes/hello-svc.yaml" %}}
```
{{% /tab %}}
{{% /tabs %}}
