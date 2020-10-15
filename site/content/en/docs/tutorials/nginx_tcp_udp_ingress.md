---
title: "Ingress nginx for TCP and UDP services"
linkTitle: "Ingress nginx for TCP and UDP services"
weight: 1
date: 2019-08-15
description: >
  How to set up a minikube ingress for TCP and UDP services
---

## Overview

The minikube [ingress addon](https://github.com/kubernetes/minikube/tree/master/deploy/addons/ingress) enables developers 
to route traffic from their host (Laptop, Desktop, etc) to a Kubernetes service running inside their minikube cluster.
The ingress addon uses the [ingress nginx](https://github.com/kubernetes/ingress-nginx) controller which by default
is only configured to listen on ports 80 and 443. TCP and UDP services listening on other ports can be enabled.

## Prerequisites

- Latest minikube binary and ISO
- Telnet command line tool
- [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command line tool
- A text editor

## Configuring TCP and UDP services with the nginx ingress controller

### Enable the ingress addon

Enable the minikube ingress addon with the following command:

```shell
minikube addons enable ingress
```

### Update the TCP and/or UDP services configmaps

Borrowing from the tutorial on [configuring TCP and UDP services with the ingress nginx controller](https://kubernetes.github.io/ingress-nginx/user-guide/exposing-tcp-udp-services/)
we will need to edit the configmap which is installed by default when enabling the minikube ingress addon.

There are 2 configmaps, 1 for TCP services and 1 for UDP services. By default they look like this: 

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tcp-services
  namespace: ingress-nginx
```

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: udp-services
  namespace: ingress-nginx
```

Since these configmaps are centralized and may contain configurations, it is best if we only patch them rather than completely overwrite them. 

Let's use this redis deployment as an example:

`redis-deployment.yaml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-deployment
  namespace: default
  labels:
    app: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - image: redis
        imagePullPolicy: Always
        name: redis
        ports:
        - containerPort: 6379
          protocol: TCP
```

Create a file `redis-deployment.yaml` and paste the contents above. Then install the redis deployment with the following command:

```shell
kubectl apply -f redis-deployment.yaml
```

Next we need to create a service that can route traffic to our pods:

`redis-service.yaml`
```yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: default
spec:
  selector:
    app: redis
  type: ClusterIP
  ports:
    - name: tcp-port
      port: 6379
      targetPort: 6379
      protocol: TCP
```

Create a file `redis-service.yaml` and paste the contents above. Then install the redis service with the following command:

```shell
kubectl apply -f redis-service.yaml
```

To add a TCP service to the nginx ingress controller you can run the following command:

```shell
kubectl patch configmap tcp-services -n kube-system --patch '{"data":{"6379":"default/redis-service:6379"}}'
```

Where:

- `6379` : the port your service should listen to from outside the minikube virtual machine
- `default` : the namespace that your service is installed in
- `redis-service` : the name of the service

We can verify that our resource was patched with the following command: 

```shell
kubectl get configmap tcp-services -n kube-system -o yaml
```

We should see something like this: 

```yaml
apiVersion: v1
data:
  "6379": default/redis-service:6379
kind: ConfigMap
metadata:
  creationTimestamp: "2019-10-01T16:19:57Z"
  labels:
    addonmanager.kubernetes.io/mode: EnsureExists
  name: tcp-services
  namespace: kube-system
  resourceVersion: "2857"
  selfLink: /api/v1/namespaces/kube-system/configmaps/tcp-services
  uid: 4f7fac22-e467-11e9-b543-080027057910
```

The only value you need to validate is that there is a value under the `data` property that looks like this: 

```yaml
  "6379": default/redis-service:6379
```

### Patch the ingress-nginx-controller

There is one final step that must be done in order to obtain connectivity from the outside cluster.
We need to patch our nginx controller so that it is listening on port 6379 and can route traffic to your service. To do
this we need to create a patch file.

`ingress-nginx-controller-patch.yaml`
```yaml
spec:
  template:
    spec:
      containers:
      - name: controller
        ports:
         - containerPort: 6379
           hostPort: 6379
```

Create a file called `ingress-nginx-controller-patch.yaml` and paste the contents above.

Next apply the changes with the following command:

```shell
kubectl patch deployment ingress-nginx-controller --patch "$(cat ingress-nginx-controller-patch.yaml)" -n kube-system
```

### Test your connection

Test that you can reach your service with telnet via the following command:

```shell
telnet $(minikube ip) 6379
```

You should see the following output:

```text
Trying 192.168.99.179...
Connected to 192.168.99.179.
Escape character is '^]'
```

To exit telnet enter the `Ctrl` key and `]` at the same time. Then type `quit` and press enter.

If you were not able to connect please review your steps above.

## Review

In the above example we did the following:

- Created a redis deployment and service in the `default` namespace
- Patched the `tcp-services` configmap in the `kube-system` namespace
- Patched the `ingress-nginx-controller` deployment in the `kube-system` namespace
- Connected to our service from the host via port 6379

You can apply the same steps that were applied to `tcp-services` to the `udp-services` configmap as well if you have a 
service that uses UDP and/or TCP

## Caveats

With the exception of ports 80 and 443, each minikube instance can only be configured for exactly 1 service to be listening 
on any particular port. Multiple TCP and/or UDP services listening on the same port in the same minikube instance is not supported 
and can not be supported until an update of the ingress spec is released. 
Please see [this document](https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/edit#) 
for the latest info on these potential changes.

## Related articles

- [Routing traffic multiple services on ports 80 and 443 in minikube with the Kubernetes Ingress resource](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/)
- [Use port forwarding to access applications in a cluster](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/)

