---
title: "Using the Registry Aliases Addon"
linkTitle: "Registry Aliases"
weight: 1
date: 2020-03-07
---

## Registry Aliases Addon

An addon to minikube that can help push and pull from the minikube registry using custom domain names. The custom domain names will be made resolveable from with in cluster and at minikube node.

## How to use ?

### Start minikube

```shell
minikube start -p demo
```
This addon depends on `registry` addon, it need to be enabled before the alias addon is installed:

### Enable internal registry

```shell
minikube addons enable registry
```

Verifying the registry deployment

```shell
watch kubectl get pods -n kube-system
```

```shell
NAME                           READY   STATUS    RESTARTS   AGE
coredns-6955765f44-kpbzt       1/1     Running   0          16m
coredns-6955765f44-lzlsv       1/1     Running   0          16m
etcd-demo                      1/1     Running   0          16m
kube-apiserver-demo            1/1     Running   0          16m
kube-controller-manager-demo   1/1     Running   0          16m
kube-proxy-q8rb9               1/1     Running   0          16m
kube-scheduler-demo            1/1     Running   0          16m
*registry-4k8zs*              1/1     Running   0          40s
registry-proxy-vs8jt           1/1     Running   0          40s
storage-provisioner            1/1     Running   0          16m
```

```shell
kubectl get svc -n kube-system
```

```shell
NAME       TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                  AGE
kube-dns   ClusterIP   10.96.0.10     <none>        53/UDP,53/TCP,9153/TCP   17m
registry   ClusterIP   10.97.247.75   <none>        80/TCP                   94s
```

>
> **NOTE:**
> Please make a note of the CLUSTER-IP of `registry` service

### Enable registry aliases addon

```shell
minikube addons enable registry-aliases
🌟  The 'registry-aliases' addon is enabled
```

You can check the mikikube vm's `/etc/hosts` file for the registry aliases entries:

```shell
watch minikube ssh -- cat /etc/hosts
```

```shell
127.0.0.1       localhost
127.0.1.1 demo
10.97.247.75    example.org
10.97.247.75    example.com
10.97.247.75    test.com
10.97.247.75    test.org
```

The above output shows that the Daemonset has added the `registryAliases` from the ConfigMap pointing to the internal registry's __CLUSTER-IP__.

### Update CoreDNS

The coreDNS would have been automatically updated by the patch-coredns. A successful job run will have coredns ConfigMap updated like:

```yaml
apiVersion: v1
data:
  Corefile: |-
    .:53 {
        errors
        health
        rewrite name example.com registry.kube-system.svc.cluster.local
        rewrite name example.org registry.kube-system.svc.cluster.local
        rewrite name test.com registry.kube-system.svc.cluster.local
        rewrite name test.org registry.kube-system.svc.cluster.local
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        proxy . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
kind: ConfigMap
metadata:
  name: coredns
```

To verify it run the following command:

```shell
kubectl get cm -n kube-system coredns -o yaml
```

Once you have successfully patched you can now push and pull from the registry using suffix `example.com`, `example.org`,`test.com` and `test.org`.

The successful run will show the following extra pods (Daemonset, Job) in `kube-system` namespace:

```shell
NAME                                    READY   STATUS      RESTARTS   AGE
registry-aliases-hosts-update-995vx     1/1     Running     0          47s
registry-aliases-patch-core-dns-zsxfc   0/1     Completed   0          47s
```

## Verify with sample application

You can verify the deployment end to end using the example [application](https://github.com/kameshsampath/minikube-registry-aliases-demo).

```shell
git clone https://github.com/kameshsampath/minikube-registry-aliases-demo
cd minikube-registry-aliases-demo
```

Make sure you set the docker context using `eval $(minikube -p demo docker-env)`

Deploy the application using [Skaffold](https://skaffold.dev):

```shell
skaffold dev --port-forward
```

Once the application is running try doing `curl localhost:8080` to see the `Hello World` response

You can also update [skaffold.yaml](./skaffold.yaml) and [app.yaml](.k8s/app.yaml), to use `test.org`, `test.com` or `example.org` as container registry urls, and see all the container image names resolves to internal registry, resulting in successful build and deployment.

> **NOTE**:
>
> You can also update [skaffold.yaml](./skaffold.yaml) and [app. yaml](.k8s/app.yaml), to use `test.org`, `test.com` or > `example.org` as container registry urls, and see all the > container image names resolves to internal registry, resulting in successful build and deployment.
