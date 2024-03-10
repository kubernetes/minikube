---
title: "Using Multi-Control Plane - HA Clusters"
linkTitle: "Using Multi-Control Plane - HA Clusters"
weight: 1
date: 2024-03-10
---

## Overview

minikube implements Kubernetes highly available cluster topology using [stacked control plane and etcd nodes](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/ha-topology/#stacked-etcd-topology) using [kube-vip](https://kube-vip.io/) in [ARP](https://kube-vip.io/#arp) mode.

This tutorial will show you how to start and explore a multi-control plane - HA cluster on minikube.

## Prerequisites

- minikube > v1.32.0
- kubectl

## Caveat

While a minikube HA cluster will continue to operate (although degraded) after loosing any one control-plane node, keep in mind that there might be some components that are attached only to the primary control-plane node, like the storage-provisioner.

## Tutorial

- Start a HA cluster with the driver and container runtime of your choice:

```shell
minikube start --ha --container-runtime=containerd --profile ha-demo
```
```
😄  [ha-demo] minikube v1.32.0 on Opensuse-Tumbleweed 20240307
✨  Automatically selected the docker driver. Other choices: kvm2, qemu2, virtualbox, ssh
📌  Using Docker driver with root privileges
👍  Starting "ha-demo" primary control-plane node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1708944392-18244 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔗  Configuring CNI (Container Networking Interface) ...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass

👍  Starting "ha-demo-m02" control-plane node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1708944392-18244 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
🌐  Found network options:
    ▪ NO_PROXY=192.168.67.2
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
    ▪ env NO_PROXY=192.168.67.2
🔎  Verifying Kubernetes components...

👍  Starting "ha-demo-m03" control-plane node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1708944392-18244 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
🌐  Found network options:
    ▪ NO_PROXY=192.168.67.2,192.168.67.3
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
    ▪ env NO_PROXY=192.168.67.2
    ▪ env NO_PROXY=192.168.67.2,192.168.67.3
🔎  Verifying Kubernetes components...
🏄  Done! kubectl is now configured to use "ha-demo" cluster and "default" namespace by default
```

- List your HA cluster nodes:

```shell
kubectl get nodes -owide
```
```
NAME          STATUS   ROLES           AGE     VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION    CONTAINER-RUNTIME
ha-demo       Ready    control-plane   3m27s   v1.28.4   192.168.67.2   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m02   Ready    control-plane   3m9s    v1.28.4   192.168.67.3   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m03   Ready    control-plane   2m38s   v1.28.4   192.168.67.4   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
```

- Check the status of your HA cluster:

```shell
minikube profile list
```
```
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
| Profile | VM Driver |  Runtime   |       IP       | Port | Version | Status | Nodes | Active |
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
| ha-demo | docker    | containerd | 192.168.67.254 | 8443 | v1.28.4 | HAppy  |     3 |        |
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
```

- Check the status of your HA cluster nodes:

```shell
minikube status -p ha-demo
```
```
ha-demo
type: Control Plane
host: Running
kubelet: Running
apiserver: Running
kubeconfig: Configured

ha-demo-m02
type: Control Plane
host: Running
kubelet: Running
apiserver: Running
kubeconfig: Configured

ha-demo-m03
type: Control Plane
host: Running
kubelet: Running
apiserver: Running
kubeconfig: Configured
```

- For a HA cluster, kubeconfig points to the Virual Kubernetes API Server IP

```shell
kubectl config view --context ha-demo
```
```
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/prezha/.minikube/ca.crt
    extensions:
    - extension:
        last-update: Sat, 09 Mar 2024 23:13:20 GMT
        provider: minikube.sigs.k8s.io
        version: v1.32.0
      name: cluster_info
    server: https://192.168.67.254:8443
  name: ha-demo
contexts:
- context:
    cluster: ha-demo
    extensions:
    - extension:
        last-update: Sat, 09 Mar 2024 23:13:20 GMT
        provider: minikube.sigs.k8s.io
        version: v1.32.0
      name: context_info
    namespace: default
    user: ha-demo
  name: ha-demo
current-context: ha-demo
kind: Config
preferences: {}
users:
- name: ha-demo
  user:
    client-certificate: /home/prezha/.minikube/profiles/ha-demo/client.crt
    client-key: /home/prezha/.minikube/profiles/ha-demo/client.key
```

- Overview of the current leader and follower API servers

```shell
minikube ssh -p ha-demo -- 'sudo /var/lib/minikube/binaries/v1.28.4/kubectl --kubeconfig=/var/lib/minikube/kubeconfig logs -n kube-system pod/kube-vip-minikube-ha'
```
```
time="2024-03-09T23:13:46Z" level=info msg="Starting kube-vip.io [v0.7.1]"
time="2024-03-09T23:13:46Z" level=info msg="namespace [kube-system], Mode: [ARP], Features(s): Control Plane:[true], Services:[false]"
time="2024-03-09T23:13:46Z" level=info msg="prometheus HTTP server started"
time="2024-03-09T23:13:46Z" level=info msg="Starting Kube-vip Manager with the ARP engine"
time="2024-03-09T23:13:46Z" level=info msg="Beginning cluster membership, namespace [kube-system], lock name [plndr-cp-lock], id [ha-demo]"
I0309 23:13:46.190758       1 leaderelection.go:250] attempting to acquire leader lease kube-system/plndr-cp-lock...
E0309 23:13:49.522895       1 leaderelection.go:332] error retrieving resource lock kube-system/plndr-cp-lock: etcdserver: leader changed
I0309 23:13:50.639076       1 leaderelection.go:260] successfully acquired lease kube-system/plndr-cp-lock
time="2024-03-09T23:13:50Z" level=info msg="Node [ha-demo] is assuming leadership of the cluster"
time="2024-03-09T23:13:50Z" level=info msg="Starting IPVS LoadBalancer"
time="2024-03-09T23:13:50Z" level=info msg="IPVS Loadbalancer enabled for 1.2.1"
time="2024-03-09T23:13:50Z" level=info msg="Gratuitous Arp broadcast will repeat every 3 seconds for [192.168.67.254]"
time="2024-03-09T23:13:50Z" level=info msg="Kube-Vip is watching nodes for control-plane labels"
time="2024-03-09T23:13:50Z" level=info msg="Added backend for [192.168.67.254:8443] on [192.168.67.3:8443]"
time="2024-03-09T23:14:07Z" level=info msg="Added backend for [192.168.67.254:8443] on [192.168.67.4:8443]"
```

```shell
minikube ssh -p ha-demo -- 'sudo /var/lib/minikube/binaries/v1.28.4/kubectl --kubeconfig=/var/lib/minikube/kubeconfig logs -n kube-system pod/kube-vip-ha-demo-m02'
```
```
time="2024-03-09T23:13:36Z" level=info msg="Starting kube-vip.io [v0.7.1]"
time="2024-03-09T23:13:36Z" level=info msg="namespace [kube-system], Mode: [ARP], Features(s): Control Plane:[true], Services:[false]"
time="2024-03-09T23:13:36Z" level=info msg="prometheus HTTP server started"
time="2024-03-09T23:13:36Z" level=info msg="Starting Kube-vip Manager with the ARP engine"
time="2024-03-09T23:13:36Z" level=info msg="Beginning cluster membership, namespace [kube-system], lock name [plndr-cp-lock], id [ha-demo-m02]"
I0309 23:13:36.656894       1 leaderelection.go:250] attempting to acquire leader lease kube-system/plndr-cp-lock...
E0309 23:13:46.663237       1 leaderelection.go:332] error retrieving resource lock kube-system/plndr-cp-lock: Get "https://kubernetes:8443/apis/coordination.k8s.io/v1/namespaces/kube-system/leases/plndr-cp-lock?timeout=10s": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
time="2024-03-09T23:13:50Z" level=info msg="Node [ha-demo] is assuming leadership of the cluster"
```

```shell
minikube ssh -p ha-demo -- 'sudo /var/lib/minikube/binaries/v1.28.4/kubectl --kubeconfig=/var/lib/minikube/kubeconfig logs -n kube-system pod/kube-vip-ha-demo-m03'
```
```
time="2024-03-09T23:14:06Z" level=info msg="Starting kube-vip.io [v0.7.1]"
time="2024-03-09T23:14:06Z" level=info msg="namespace [kube-system], Mode: [ARP], Features(s): Control Plane:[true], Services:[false]"
time="2024-03-09T23:14:06Z" level=info msg="prometheus HTTP server started"
time="2024-03-09T23:14:06Z" level=info msg="Starting Kube-vip Manager with the ARP engine"
time="2024-03-09T23:14:06Z" level=info msg="Beginning cluster membership, namespace [kube-system], lock name [plndr-cp-lock], id [ha-demo-m03]"
I0309 23:14:06.972839       1 leaderelection.go:250] attempting to acquire leader lease kube-system/plndr-cp-lock...
time="2024-03-09T23:14:08Z" level=info msg="Node [ha-demo] is assuming leadership of the cluster"
```


- Overview of multi-etcd instances

```shell
minikube ssh -p ha-demo -- 'sudo /var/lib/minikube/binaries/v1.28.4/kubectl --kubeconfig=/var/lib/minikube/kubeconfig exec -ti pod/etcd-ha-demo -n kube-system -- /bin/sh -c "ETCDCTL_API=3 etcdctl member list --write-out=table --cacert=/var/lib/minikube/certs/etcd/ca.crt --cert=/var/lib/minikube/certs/etcd/server.crt --key=/var/lib/minikube/certs/etcd/server.key"'
```
```
+------------------+---------+-------------+---------------------------+---------------------------+------------+
|        ID        | STATUS  |    NAME     |        PEER ADDRS         |       CLIENT ADDRS        | IS LEARNER |
+------------------+---------+-------------+---------------------------+---------------------------+------------+
|  798ac3f6077b24a | started | ha-demo-m02 | https://192.168.67.3:2380 | https://192.168.67.3:2379 |      false |
| 7930f2d13748db77 | started | ha-demo-m03 | https://192.168.67.4:2380 | https://192.168.67.4:2379 |      false |
| 8688e899f7831fc7 | started |     ha-demo | https://192.168.67.2:2380 | https://192.168.67.2:2379 |      false |
+------------------+---------+-------------+---------------------------+---------------------------+------------+
```

- Load balancer configuration

```shell
minikube ssh -p ha-demo -- 'sudo apt update && sudo apt install ipvsadm -y'
```
```shell
minikube ssh -p ha-demo -- 'sudo ipvsadm -ln'
```
```
IP Virtual Server version 1.2.1 (size=4096)
Prot LocalAddress:Port Scheduler Flags
  -> RemoteAddress:Port           Forward Weight ActiveConn InActConn
TCP  192.168.67.254:8443 rr
  -> 192.168.67.2:8443            Local   1      4          23
  -> 192.168.67.3:8443            Local   1      3          24
  -> 192.168.67.4:8443            Local   1      1          24
```

- Load balancer connections

```shell
minikube ssh -p ha-demo -- 'sudo ipvsadm -lnc | grep -i established'
```
```
TCP 14:58  ESTABLISHED 192.168.67.4:36468 192.168.67.254:8443 192.168.67.3:8443
TCP 14:40  ESTABLISHED 192.168.67.3:54706 192.168.67.254:8443 192.168.67.3:8443
TCP 14:58  ESTABLISHED 192.168.67.4:36464 192.168.67.254:8443 192.168.67.3:8443
TCP 14:57  ESTABLISHED 192.168.67.3:54260 192.168.67.254:8443 192.168.67.2:8443
TCP 14:55  ESTABLISHED 192.168.67.254:44480 192.168.67.254:8443 192.168.67.2:8443
TCP 14:59  ESTABLISHED 192.168.67.1:35728 192.168.67.254:8443 192.168.67.4:8443
TCP 14:58  ESTABLISHED 192.168.67.254:37856 192.168.67.254:8443 192.168.67.2:8443
```
- Loosing a control-plane node - degrades cluster, but not a problem!

```shell
minikube node delete m02 -p ha-demo
```
```
🔥  Deleting node m02 from cluster ha-demo
✋  Stopping node "ha-demo-m02"  ...
🛑  Powering off "ha-demo-m02" via SSH ...
🔥  Deleting "ha-demo-m02" in docker ...
💀  Node m02 was successfully deleted.
```
```shell
kubectl get nodes -owide
```
```
NAME          STATUS   ROLES           AGE   VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION    CONTAINER-RUNTIME
ha-demo       Ready    control-plane   70m   v1.28.4   192.168.67.2   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m03   Ready    control-plane   69m   v1.28.4   192.168.67.4   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
```
```shell
minikube profile list
```
```
|---------|-----------|------------|----------------|------|---------|----------|-------|--------|
| Profile | VM Driver |  Runtime   |       IP       | Port | Version |  Status  | Nodes | Active |
|---------|-----------|------------|----------------|------|---------|----------|-------|--------|
| ha-demo | docker    | containerd | 192.168.67.254 | 8443 | v1.28.4 | Degraded |     2 |        |
|---------|-----------|------------|----------------|------|---------|----------|-------|--------|
```


- Add a control-plane node

```shell
minikube node add --control-plane -p ha-demo
```
```
😄  Adding node m04 to cluster ha-demo as [worker control-plane]
👍  Starting "ha-demo-m04" control-plane node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1708944392-18244 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
🔎  Verifying Kubernetes components...
🏄  Successfully added m04 to ha-demo!
```
```shell
kubectl get nodes -owide
```
```
NAME          STATUS   ROLES           AGE   VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION    CONTAINER-RUNTIME
ha-demo       Ready    control-plane   72m   v1.28.4   192.168.67.2   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m03   Ready    control-plane   71m   v1.28.4   192.168.67.4   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m04   Ready    control-plane   28s   v1.28.4   192.168.67.5   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
```
```shell
minikube profile list
```
```
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
| Profile | VM Driver |  Runtime   |       IP       | Port | Version | Status | Nodes | Active |
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
| ha-demo | docker    | containerd | 192.168.67.254 | 8443 | v1.28.4 | HAppy  |     3 |        |
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
```

- Add a worker node

```shell
minikube node add -p ha-demo
```
```
😄  Adding node m05 to cluster ha-demo as [worker]
👍  Starting "ha-demo-m05" worker node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1708944392-18244 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
🔎  Verifying Kubernetes components...
🏄  Successfully added m05 to ha-demo!
```

```shell
kubectl get nodes -owide
```
```
NAME          STATUS   ROLES           AGE    VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION    CONTAINER-RUNTIME
ha-demo       Ready    control-plane   73m    v1.28.4   192.168.67.2   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m03   Ready    control-plane   73m    v1.28.4   192.168.67.4   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m04   Ready    control-plane   110s   v1.28.4   192.168.67.5   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m05   Ready    <none>          23s    v1.28.4   192.168.67.6   <none>        Ubuntu 22.04.3 LTS   6.7.7-1-default   containerd://1.6.28
```

- Test by deploying a hello service, which just spits back the IP address the request was served from:

```shell
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 100%
  selector:
    matchLabels:
      app: hello
  template:
    metadata:
      labels:
        app: hello
    spec:
      affinity:
        # ⬇⬇⬇ This ensures pods will land on separate hosts
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchExpressions: [{ key: app, operator: In, values: [hello] }]
              topologyKey: "kubernetes.io/hostname"
      containers:
        - name: hello-from
          image: pbitty/hello-from:latest
          ports:
            - name: http
              containerPort: 80
      terminationGracePeriodSeconds: 1
---
apiVersion: v1
kind: Service
metadata:
  name: hello
spec:
  type: NodePort
  selector:
    app: hello
  ports:
    - protocol: TCP
      nodePort: 31000
      port: 80
      targetPort: http
EOF
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

- Check out the IP addresses of our pods, to note for future reference

```shell
kubectl get pods -owide
```
```
NAME                     READY   STATUS    RESTARTS   AGE     IP           NODE          NOMINATED NODE   READINESS GATES
hello-7bf57d9696-2zrmx   1/1     Running   0          3m56s   10.244.0.4   ha-demo       <none>           <none>
hello-7bf57d9696-hd75l   1/1     Running   0          3m56s   10.244.1.2   ha-demo-m04   <none>           <none>
hello-7bf57d9696-vn6jd   1/1     Running   0          3m56s   10.244.4.2   ha-demo-m05   <none>           <none>
```

- Look at our service, to know what URL to hit

```shell
minikube service list -p ha-demo
```
```
|-------------|------------|--------------|---------------------------|
|  NAMESPACE  |    NAME    | TARGET PORT  |            URL            |
|-------------|------------|--------------|---------------------------|
| default     | hello      |           80 | http://192.168.67.2:31000 |
| default     | kubernetes | No node port |                           |
| kube-system | kube-dns   | No node port |                           |
|-------------|------------|--------------|---------------------------|
```

- Let's hit the URL a few times and see what comes back

```shell
curl  http://192.168.67.2:31000
```
```
Hello from hello-7bf57d9696-vn6jd (10.244.4.2)

curl  http://192.168.67.2:31000
Hello from hello-7bf57d9696-2zrmx (10.244.0.4)

curl  http://192.168.67.2:31000
Hello from hello-7bf57d9696-hd75l (10.244.1.2)

curl  http://192.168.67.2:31000
Hello from hello-7bf57d9696-vn6jd (10.244.4.2)
```
