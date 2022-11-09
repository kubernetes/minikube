---
title: "Setting Up Docker Desktop, Minikube and addons on WSL2"                                 
linkTitle: "Setting Up Docker Desktop, Minikube and addons on WSL2"
weight: 1
date: 2022-11-09
---

## Overview

- This guide will show you how to setup Minikube and addons on WS2.

## Prerequisite

Choose one of the dcumentation below for enabling WSL2 and adding Docker
1. [Install WSL](https://docs.microsoft.com/en-us/windows/wsl/install)
2. Install and setup [Docker Desktop for Windows](https://docs.docker.com/desktop/windows/wsl/#download)

## Kubernetes installation for WSL2

### 1-To avoid problems

```console
sudo mkdir /sys/fs/cgroup/systemd && mount -t cgroup -o none,name=systemd cgroup /sys/fs/cgroup/systemd
sudo rm -rf ~/.minikube
```

### 2-Make a clean docker installation

#### Prerequesite

```console
sudo apt-get update
sudo apt-get remove -t docker docker-engine docker.io containerd runc
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg-agent software-properties-common
```
    
#### Installation

```console
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
sudo apt-get install -y docker-ce docker-ce-cli containerd.io
sudo service docker start
sudo groupadd docker
sudo usermod -aG docker $USER
```

### 3-Install Minikube

```console
curl -sLo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
chmod +x minikube
sudo mkdir -p /usr/local/bin/
sudo install minikube /usr/local/bin/
minikube start --driver=docker
```

### 4-Restart Minikube (after ubuntu reboot)

```console
service docker start
minikube start --driver=docker
```

### 5-Configuration

#### Addons

```console
minikube addons list
```

| ADDON NAME | PROFILE |STATUS|
|--|--|--|
| ambassador | minikube | disabled |
| dashboard | minikube | disabled |
| default-storageclass | minikube | enabled ✅ |
| efk | minikube | disabled |
| freshpod | minikube | disabled |
| gvisor | minikube | disabled |
| helm-tiller | minikube | disabled |
| ingress | minikube | disabled |
| ingress-dns | minikube | disabled |
| istio| minikube | disabled |
| istio-provisioner | minikube | disabled |
| logviewer | minikube | disabled |
| metallb | minikube | disabled |
| metrics-server | minikube | disabled |
| nvidia-driver-installer | minikube | disabled |
| nvidia-gpu-device-plugin | minikube | disabled |
| olm | minikube | disabled |
| registry | minikube | disabled |
| registry-aliases | minikube | disabled |
| registry-creds | minikube | disabled |
| storage-provisioner | minikube | enabled ✅ |
| storage-provisioner-gluster | minikube | disabled |

The next will concentrer on adding following addons :

- Kubernetes Dashoard
- Helm : Package manager
- Metrics : Intern Metrics Collection

#### Kubernetes Dashboards

Enable addon :  `minikube addons enable dashboard`

Available via namespace :  `kubernetes-dashboard`

Let's check : `kubectl get -n kubernetes-dashboard pod,deployment,service`

> NAME                                             READY   STATUS   
> RESTARTS   AGE pod/dashboard-metrics-scraper-7b64584c5c-nz682   1/1   
> Running   0          18m pod/kubernetes-dashboard-5b48b67b68-f4w9s    
> 1/1     Running   0          18m
> 
> NAME                                        READY   UP-TO-DATE  
> AVAILABLE   AGE deployment.apps/dashboard-metrics-scraper   1/1     1 
> 1           18m deployment.apps/kubernetes-dashboard        1/1     1 
> 1           18m
> 
> NAME                                TYPE        CLUSTER-IP     
> EXTERNAL-IP   PORT(S)    AGE service/dashboard-metrics-scraper  
> ClusterIP   10.110.233.80   <none>        8000/TCP   18m
> service/kubernetes-dashboard        ClusterIP   10.96.57.22     <none>
> 80/TCP     18m
> 
> Connectons-nous à l’interface visualle du dashboard :
> 
> kubectl port-forward -n kubernetes-dashboard svc/kubernetes-dashboard
> 8080:80
> 
> Forwarding from 127.0.0.1:8080 -> 9090 Forwarding from [::1]:8080 ->
> 9090 Handling connection for 8080

Now it is available at this address http://localhost:8080/

![Ecran d'acceuil et apercu de l'application Kubernetes Dashboard](https://d3uyj2gj5wa63n.cloudfront.net/wp-content/uploads/2020/06/image-1024x772.png)

[Kubernetes Dashboard official site](https://kubernetes.io/fr/docs/tasks/access-application-cluster/web-ui-dashboard/)

#### Helm

Enable addon :  `minikube addons enable dashboard`

Check availability : `kubectl get pod -n kube-system -l name=tiller`

To install Helm client last release

```console
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
sudo apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
```

[Helm Official documentation](https://helm.sh/docs/)

#### Supervision des métriques Kubernetes

Enable addon : `minikube addons enable metrics-server`
Check availability : `kubectl get pod,deploy,svc -n kube-system -l k8s-app=metrics-server`

Visualize used ressources : `kubectl top nodes`

> NAME      CPU(cores)  CPU%  MEMORY(bytes)  MEMORY%
> minikube  215m        2%    6Mi            0%

Visualize used ressources (detailed) : `kubectl top pods -A`

> NAMESPACE     NAME                               CPU(cores)  
> MEMORY(bytes) kube-system   coredns-5644d7b6d9-tmskc           4m     
> 14Mi kube-system   coredns-5644d7b6d9-x5vjn           4m          
> 12Mi kube-system   etcd-minikube                      27m         
> 34Mi kube-system   kube-apiserver-minikube            70m         
> 283Mi kube-system   kube-controller-manager-minikube   32m         
> 44Mi kube-system   kube-proxy-qhdts                   2m          
> 16Mi kube-system   kube-scheduler-minikube            2m          
> 15Mi kube-system   metrics-server-6754dbc9df-rtsqx    0m          
> 10Mi kube-system   storage-provisioner                1m          
> 19Mi

[More about metrics-server](https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-metrics-pipeline/#metrics-server)

#### Additional

[Source of this tuto and additional tuto for other addons](https://blog.ineat-group.com/2020/06/utiliser-kubernetes-en-local-avec-minikube-sous-windows-10/)
