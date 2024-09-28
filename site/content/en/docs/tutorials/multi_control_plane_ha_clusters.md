---
title: "Using Multi-Control Plane - HA Clusters"
linkTitle: "Using Multi-Control Plane - HA Clusters"
weight: 1
date: 2024-03-10
---

## Overview

minikube implements Kubernetes highly available cluster topology using [stacked control plane nodes with colocated etcd nodes](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/ha-topology/#stacked-etcd-topology) using [kube-vip](https://kube-vip.io/) in [ARP](https://kube-vip.io/#arp) mode.

This tutorial will show you how to start and explore a multi-control plane - HA cluster on minikube.

## Prerequisites

- minikube > v1.32.0
- kubectl

### Optional

- ip_vs kernel modules: ip_vs, ip_vs_rr and nf_conntrack

rationale:

kube-vip supports the [control-plane load-balancing](https://kube-vip.io/docs/about/architecture/?query=load-balanc#control-plane-load-balancing) by distributing API requests across control-plane nodes using IPVS (IP Virtual Server) and Layer 4 (TCP-based) round-robin.

minikube will try to automatically enable control-plane load-balancing if these ip_vs kernel modules are available, whereas (see the [drivers]({{<ref "/docs/drivers/">}}) section for more details):

- for VM-based drivers (eg, kvm2 or qemu): minikube will automatically try to load ip_vs kernel modules
- for container-based or bare-metal-based drivers (eg, docker or "none"): minikube will only check if ip_vs kernel modules are already loaded, but will not try to load them automatically (to avoid unintentional modification of the underlying host's os/kernel), so it's up to the user to make them available, if applicable and desired

## Caveat

While a minikube HA cluster will continue to operate (although in degraded mode) after loosing any one control-plane node, keep in mind that there might be some components that are attached only to the primary control-plane node, like the storage-provisioner.

## Tutorial

- optional: if you plan on using a container-based or bare-metal-based driver on top of a Linux OS, check if the ip_vs kernel modules are already loaded

```shell
lsmod | grep ip_vs
```
```
ip_vs_rr               12288  1
ip_vs                 233472  3 ip_vs_rr
nf_conntrack          217088  9 xt_conntrack,nf_nat,nf_conntrack_tftp,nft_ct,xt_nat,nf_nat_tftp,nf_conntrack_netlink,xt_MASQUERADE,ip_vs
nf_defrag_ipv6         24576  2 nf_conntrack,ip_vs
libcrc32c              12288  5 nf_conntrack,nf_nat,btrfs,nf_tables,ip_vs
```

- Start a HA cluster with the driver and container runtime of your choice (here we chose docker driver and containerd container runtime):

```shell
minikube start --ha --driver=docker --container-runtime=containerd --profile ha-demo
```
```
😄  [ha-demo] minikube v1.32.0 on Opensuse-Tumbleweed 20240311
✨  Using the docker driver based on user configuration
📌  Using Docker driver with root privileges
👍  Starting "ha-demo" primary control-plane node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1710284843-18375 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔗  Configuring CNI (Container Networking Interface) ...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass

👍  Starting "ha-demo-m02" control-plane node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1710284843-18375 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
🌐  Found network options:
    ▪ NO_PROXY=192.168.49.2
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
    ▪ env NO_PROXY=192.168.49.2
🔎  Verifying Kubernetes components...

👍  Starting "ha-demo-m03" control-plane node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1710284843-18375 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
🌐  Found network options:
    ▪ NO_PROXY=192.168.49.2,192.168.49.3
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
    ▪ env NO_PROXY=192.168.49.2
    ▪ env NO_PROXY=192.168.49.2,192.168.49.3
🔎  Verifying Kubernetes components...
🏄  Done! kubectl is now configured to use "ha-demo" cluster and "default" namespace by default
```

- List your HA cluster nodes:

```shell
kubectl get nodes -owide
```
```
NAME          STATUS   ROLES           AGE     VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION    CONTAINER-RUNTIME
ha-demo       Ready    control-plane   4m21s   v1.28.4   192.168.49.2   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m02   Ready    control-plane   4m      v1.28.4   192.168.49.3   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m03   Ready    control-plane   3m37s   v1.28.4   192.168.49.4   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
```

- Check the status of your HA cluster:

```shell
minikube profile list
```
```
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
| Profile | VM Driver |  Runtime   |       IP       | Port | Version | Status | Nodes | Active |
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
| ha-demo | docker    | containerd | 192.168.49.254 | 8443 | v1.28.4 | HAppy  |     3 |        |
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
        last-update: Thu, 14 Mar 2024 21:38:07 GMT
        provider: minikube.sigs.k8s.io
        version: v1.32.0
      name: cluster_info
    server: https://192.168.49.254:8443
  name: ha-demo
contexts:
- context:
    cluster: ha-demo
    extensions:
    - extension:
        last-update: Thu, 14 Mar 2024 21:38:07 GMT
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
minikube ssh -p ha-demo -- 'find /var/lib/minikube/binaries -iname kubectl -exec sudo {} --kubeconfig=/var/lib/minikube/kubeconfig logs -n kube-system pod/kube-vip-ha-demo \; -quit'
```
```
time="2024-03-14T21:38:34Z" level=info msg="Starting kube-vip.io [v0.7.1]"
time="2024-03-14T21:38:34Z" level=info msg="namespace [kube-system], Mode: [ARP], Features(s): Control Plane:[true], Services:[false]"
time="2024-03-14T21:38:34Z" level=info msg="prometheus HTTP server started"
time="2024-03-14T21:38:34Z" level=info msg="Starting Kube-vip Manager with the ARP engine"
time="2024-03-14T21:38:34Z" level=info msg="Beginning cluster membership, namespace [kube-system], lock name [plndr-cp-lock], id [ha-demo]"
I0314 21:38:34.042420       1 leaderelection.go:250] attempting to acquire leader lease kube-system/plndr-cp-lock...
I0314 21:38:34.069509       1 leaderelection.go:260] successfully acquired lease kube-system/plndr-cp-lock
time="2024-03-14T21:38:34Z" level=info msg="Node [ha-demo] is assuming leadership of the cluster"
time="2024-03-14T21:38:34Z" level=info msg="Starting IPVS LoadBalancer"
time="2024-03-14T21:38:34Z" level=info msg="IPVS Loadbalancer enabled for 1.2.1"
time="2024-03-14T21:38:34Z" level=info msg="Gratuitous Arp broadcast will repeat every 3 seconds for [192.168.49.254]"
time="2024-03-14T21:38:34Z" level=info msg="Kube-Vip is watching nodes for control-plane labels"
time="2024-03-14T21:38:34Z" level=info msg="Added backend for [192.168.49.254:8443] on [192.168.49.3:8443]"
time="2024-03-14T21:38:48Z" level=info msg="Added backend for [192.168.49.254:8443] on [192.168.49.4:8443]"
```

```shell
minikube ssh -p ha-demo -- 'find /var/lib/minikube/binaries -iname kubectl -exec sudo {} --kubeconfig=/var/lib/minikube/kubeconfig logs -n kube-system pod/kube-vip-ha-demo-m02 \; -quit'
```
```
time="2024-03-14T21:38:25Z" level=info msg="Starting kube-vip.io [v0.7.1]"
time="2024-03-14T21:38:25Z" level=info msg="namespace [kube-system], Mode: [ARP], Features(s): Control Plane:[true], Services:[false]"
time="2024-03-14T21:38:25Z" level=info msg="prometheus HTTP server started"
time="2024-03-14T21:38:25Z" level=info msg="Starting Kube-vip Manager with the ARP engine"
time="2024-03-14T21:38:25Z" level=info msg="Beginning cluster membership, namespace [kube-system], lock name [plndr-cp-lock], id [ha-demo-m02]"
I0314 21:38:25.990817       1 leaderelection.go:250] attempting to acquire leader lease kube-system/plndr-cp-lock...
time="2024-03-14T21:38:34Z" level=info msg="Node [ha-demo] is assuming leadership of the cluster"
```

```shell
minikube ssh -p ha-demo -- 'find /var/lib/minikube/binaries -iname kubectl -exec sudo {} --kubeconfig=/var/lib/minikube/kubeconfig logs -n kube-system pod/kube-vip-ha-demo-m03 \; -quit'
```
```
time="2024-03-14T21:38:48Z" level=info msg="Starting kube-vip.io [v0.7.1]"
time="2024-03-14T21:38:48Z" level=info msg="namespace [kube-system], Mode: [ARP], Features(s): Control Plane:[true], Services:[false]"
time="2024-03-14T21:38:48Z" level=info msg="prometheus HTTP server started"
time="2024-03-14T21:38:48Z" level=info msg="Starting Kube-vip Manager with the ARP engine"
time="2024-03-14T21:38:48Z" level=info msg="Beginning cluster membership, namespace [kube-system], lock name [plndr-cp-lock], id [ha-demo-m03]"
I0314 21:38:48.856781       1 leaderelection.go:250] attempting to acquire leader lease kube-system/plndr-cp-lock...
time="2024-03-14T21:38:48Z" level=info msg="Node [ha-demo] is assuming leadership of the cluster"
```

- Overview of multi-etcd instances

```shell
minikube ssh -p ha-demo -- 'find /var/lib/minikube/binaries -iname kubectl -exec sudo {} --kubeconfig=/var/lib/minikube/kubeconfig exec -ti pod/etcd-ha-demo -n kube-system -- /bin/sh -c "ETCDCTL_API=3 etcdctl member list --write-out=table --cacert=/var/lib/minikube/certs/etcd/ca.crt --cert=/var/lib/minikube/certs/etcd/server.crt --key=/var/lib/minikube/certs/etcd/server.key" \; -quit'
```
```
+------------------+---------+-------------+---------------------------+---------------------------+------------+
|        ID        | STATUS  |    NAME     |        PEER ADDRS         |       CLIENT ADDRS        | IS LEARNER |
+------------------+---------+-------------+---------------------------+---------------------------+------------+
| 3c464e4a52eb93c5 | started | ha-demo-m03 | https://192.168.49.4:2380 | https://192.168.49.4:2379 |      false |
| 59bde6852118b2a5 | started | ha-demo-m02 | https://192.168.49.3:2380 | https://192.168.49.3:2379 |      false |
| aec36adc501070cc | started |     ha-demo | https://192.168.49.2:2380 | https://192.168.49.2:2379 |      false |
+------------------+---------+-------------+---------------------------+---------------------------+------------+
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
NAME          STATUS   ROLES           AGE     VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION    CONTAINER-RUNTIME
ha-demo       Ready    control-plane   7m16s   v1.28.4   192.168.49.2   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m03   Ready    control-plane   6m32s   v1.28.4   192.168.49.4   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
```
```shell
minikube profile list
```
```
|---------|-----------|------------|----------------|------|---------|----------|-------|--------|
| Profile | VM Driver |  Runtime   |       IP       | Port | Version |  Status  | Nodes | Active |
|---------|-----------|------------|----------------|------|---------|----------|-------|--------|
| ha-demo | docker    | containerd | 192.168.49.254 | 8443 | v1.28.4 | Degraded |     2 |        |
|---------|-----------|------------|----------------|------|---------|----------|-------|--------|
```

- Add a control-plane node

```shell
minikube node add --control-plane -p ha-demo
```
```
😄  Adding node m04 to cluster ha-demo as [worker control-plane]
👍  Starting "ha-demo-m04" control-plane node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1710284843-18375 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
🔎  Verifying Kubernetes components...
🏄  Successfully added m04 to ha-demo!
```
```shell
kubectl get nodes -owide
```
```
NAME          STATUS   ROLES           AGE     VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION    CONTAINER-RUNTIME
ha-demo       Ready    control-plane   8m34s   v1.28.4   192.168.49.2   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m03   Ready    control-plane   7m50s   v1.28.4   192.168.49.4   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m04   Ready    control-plane   36s     v1.28.4   192.168.49.5   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
```
```shell
minikube profile list
```
```
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
| Profile | VM Driver |  Runtime   |       IP       | Port | Version | Status | Nodes | Active |
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
| ha-demo | docker    | containerd | 192.168.49.254 | 8443 | v1.28.4 | HAppy  |     3 |        |
|---------|-----------|------------|----------------|------|---------|--------|-------|--------|
```

- Add a worker node

```shell
minikube node add -p ha-demo
```
```
😄  Adding node m05 to cluster ha-demo as [worker]
👍  Starting "ha-demo-m05" worker node in "ha-demo" cluster
🚜  Pulling base image v0.0.42-1710284843-18375 ...
🔥  Creating docker container (CPUs=2, Memory=5266MB) ...
📦  Preparing Kubernetes v1.28.4 on containerd 1.6.28 ...
🔎  Verifying Kubernetes components...
🏄  Successfully added m05 to ha-demo!
```

```shell
kubectl get nodes -owide
```
```
NAME          STATUS   ROLES           AGE     VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION    CONTAINER-RUNTIME
ha-demo       Ready    control-plane   9m35s   v1.28.4   192.168.49.2   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m03   Ready    control-plane   8m51s   v1.28.4   192.168.49.4   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m04   Ready    control-plane   97s     v1.28.4   192.168.49.5   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
ha-demo-m05   Ready    <none>          22s     v1.28.4   192.168.49.6   <none>        Ubuntu 22.04.4 LTS   6.7.7-1-default   containerd://1.6.28
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
NAME                     READY   STATUS    RESTARTS   AGE   IP           NODE          NOMINATED NODE   READINESS GATES
hello-7bf57d9696-64v6m   1/1     Running   0          18s   10.244.3.2   ha-demo-m04   <none>           <none>
hello-7bf57d9696-7gtlk   1/1     Running   0          18s   10.244.2.2   ha-demo-m03   <none>           <none>
hello-7bf57d9696-99qsw   1/1     Running   0          18s   10.244.0.4   ha-demo       <none>           <none>
```

- Look at our service, to know what URL to hit

```shell
minikube service list -p ha-demo
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
Hello from hello-7bf57d9696-99qsw (10.244.0.4)

curl  http://192.168.49.2:31000
Hello from hello-7bf57d9696-7gtlk (10.244.2.2)

curl  http://192.168.49.2:31000
Hello from hello-7bf57d9696-7gtlk (10.244.2.2)

curl  http://192.168.49.2:31000
Hello from hello-7bf57d9696-64v6m (10.244.3.2)
```
