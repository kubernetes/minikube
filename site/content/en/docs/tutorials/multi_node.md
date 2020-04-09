---
title: "Using Multi-Node Clusters (Experimental)"
linkTitle: "Using multi-node clusters"
weight: 1
date: 2019-11-24
---

## Overview

- This tutorial will show you how to start a multi-node clusters on minikube and deploy a service to it.

## Prerequisites

- minikube 1.9.0 or higher
- kubectl

## Tutorial

- Start a cluster with 2 nodes in the driver of your choice (the extra parameters are to make our chosen CNI, flannel, work while we're still experimental):
```
minikube start --nodes 2 -p multinode-demo --network-plugin=cni --extra-config=kubeadm.pod-network-cidr=10.244.0.0/16
😄  [multinode-demo] minikube v1.9.2 on Darwin 10.14.6
✨  Automatically selected the hyperkit driver
👍  Starting control plane node m01 in cluster multinode-demo
🔥  Creating hyperkit VM (CPUs=2, Memory=4000MB, Disk=20000MB) ...
🐳  Preparing Kubernetes v1.18.0 on Docker 19.03.8 ...
🌟  Enabling addons: default-storageclass, storage-provisioner

👍  Starting node m02 in cluster multinode-demo
🔥  Creating hyperkit VM (CPUs=2, Memory=4000MB, Disk=20000MB) ...
🌐  Found network options:
    ▪ NO_PROXY=192.168.64.213
🐳  Preparing Kubernetes v1.18.0 on Docker 19.03.8 ...
🏄  Done! kubectl is now configured to use "multinode-demo"
```

- Get the list of your nodes:
```
kubectl get nodes
NAME                 STATUS   ROLES    AGE     VERSION
multinode-demo       Ready    master   9m58s   v1.18.0
multinode-demo-m02   Ready    <none>   9m5s    v1.18.0
```

- Install a CNI (e.g. flannel):
NOTE: This currently needs to be done manually after the apiserver is running, the multi-node feature is still experimental as of 1.9.2.
```
kubectl apply -f kube-flannel.yaml
podsecuritypolicy.policy/psp.flannel.unprivileged created
clusterrole.rbac.authorization.k8s.io/flannel created
clusterrolebinding.rbac.authorization.k8s.io/flannel created
serviceaccount/flannel created
configmap/kube-flannel-cfg created
daemonset.apps/kube-flannel-ds-amd64 created
daemonset.apps/kube-flannel-ds-arm64 created
daemonset.apps/kube-flannel-ds-arm created
daemonset.apps/kube-flannel-ds-ppc64le created
daemonset.apps/kube-flannel-ds-s390x created
```

- Deploy our hello world deployment:
```
kubectl apply -f hello-deployment.yaml
deployment.apps/hello created

kubectl rollout status deployment/hello
deployment "hello" successfully rolled out
```


- Deploy our hello world service, which just spits back the IP address the request was served from:
{{% readfile file="/docs/tutorials/includes/hello-svc.yaml" %}}
```
kubectl apply -f hello-svc.yml
service/hello created
```


- Check out the IP addresses of our pods, to note for future reference
```
kubectl get pods -o wide
NAME                    READY   STATUS    RESTARTS   AGE   IP           NODE             NOMINATED NODE   READINESS GATES
hello-c7b8df44f-qbhxh   1/1     Running   0          31s   10.244.0.3   multinode-demo   <none>           <none>
hello-c7b8df44f-xv4v6   1/1     Running   0          31s   10.244.0.2   multinode-demo   <none>           <none>
```

- Look at our service, to know what URL to hit
```
minikube service list
|-------------|------------|--------------|-----------------------------|
|  NAMESPACE  |    NAME    | TARGET PORT  |             URL             |
|-------------|------------|--------------|-----------------------------|
| default     | hello      |           80 | http://192.168.64.226:31000 |
| default     | kubernetes | No node port |                             |
| kube-system | kube-dns   | No node port |                             |
|-------------|------------|--------------|-----------------------------|
```

- Let's hit the URL a few times and see what comes back
```
curl  http://192.168.64.226:31000
Hello from hello-c7b8df44f-qbhxh (10.244.0.3)

curl  http://192.168.64.226:31000
Hello from hello-c7b8df44f-qbhxh (10.244.0.3)

curl  http://192.168.64.226:31000
Hello from hello-c7b8df44f-xv4v6 (10.244.0.2)

curl  http://192.168.64.226:31000
Hello from hello-c7b8df44f-xv4v6 (10.244.0.2)
```

- Multiple nodes!


- Referenced YAML files
{{% tabs %}}
{{% tab kube-flannel.yaml %}}
```
{{% readfile file="/docs/tutorials/includes/kube-flannel.yaml" %}}
```
{{% /tab %}}
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
