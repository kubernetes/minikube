---
title: "DNS Domain"
linkTitle: "DNS Domain"
weight: 6
date: 2019-10-09
description: >
  Use configured DNS domain in bootstrapper kubeadm
---

minikube by default uses **cluster.local** if none is specified via the start flag --dns-domain. The configuration file used by kubeadm are found inside **/var/tmp/minikube/kubeadm.yaml** directory inside minikube.

Default DNS configuration will look like below

```
apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
localAPIEndpoint:
......
......
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
.....
.....
kubernetesVersion: v1.16.0
networking:
  dnsDomain: cluster.local
  podSubnet: ""
  serviceSubnet: 10.96.0.0/12
---
```

To change the dns pass the value when starting minikube 

```
minikube start --dns-domain bla.blah.blah   
```

the dns now changed to bla.blah.blah

```
apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
localAPIEndpoint:
......
......
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
.....
.....
kubernetesVersion: v1.16.0
networking:
  dnsDomain: bla.blah.blah
  podSubnet: ""
  serviceSubnet: 10.96.0.0/12
---
```