---
title: "Using the Kong Ingress Controller Addon"
linkTitle: "Kong Ingress"
weight: 1
date: 2022-01-25
---
Kong Ingress Controller (KIC) running on your minikube server.

1. Start `minikube`

   ```bash
   minikube start
   ```

   It will take a few minutes to get all resources provisioned.

   ```bash
   kubectl get nodes
   ```

## Deploy the Kong Ingress Controller

Enable Kong Ingress Controller via `minikube` command.

```bash
$ minikube addons enable kong

🌟  The 'kong' addon is enabled
```

> Note: this process could take up to five minutes the first time.

## Setup environment variables

Next, we will set up an environment variable with the IP address at which
Kong is accessible.
We can use it to send requests into the Kubernetes cluster.

```bash
$ export PROXY_IP=$(minikube service -n kong kong-proxy --url | head -1)
$ echo $PROXY_IP
http://192.168.99.100:32728
```

Alternatively, you can use `minikube tunnel` command.

```bash

# open another terminal window and run
minikube tunnel

# you may need to enter an admin password because minikube need to use ports 80 and 443 
```

Let's test if KIC is up and running.

```bash
$ curl -v localhost

*   Trying 127.0.0.1:80...
* Connected to localhost (127.0.0.1) port 80 (#0)
> GET / HTTP/1.1
> Host: localhost
> User-Agent: curl/7.86.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 404 Not Found
< Date: Wed, 03 May 2023 01:34:31 GMT
< Content-Type: application/json; charset=utf-8
< Connection: keep-alive
< Content-Length: 48
< X-Kong-Response-Latency: 0
< Server: kong/3.2.2
<
* Connection #0 to host localhost left intact
{"message":"no Route matched with those values"}%
````

## Creating Ingress object

To proxy requests, you need an upstream application to proxy to. 
Deploying this echo server provides a simple application that returns information about the Pod it’s running in:

```bash
echo "
apiVersion: v1
kind: Service
metadata:
  labels:
    app: echo
  name: echo
spec:
  ports:
  - port: 1025
    name: tcp
    protocol: TCP
    targetPort: 1025
  - port: 1026
    name: udp
    protocol: TCP
    targetPort: 1026
  - port: 1027
    name: http
    protocol: TCP
    targetPort: 1027
  selector:
    app: echo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: echo
  name: echo
spec:
  replicas: 1
  selector:
    matchLabels:
      app: echo
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: echo
    spec:
      containers:
      - image: kong/go-echo:latest
        name: echo
        ports:
        - containerPort: 1027
        env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
        resources: {}
" | kubectl apply -f -
```

Next, we will create routing configuration to proxy `/echo` requests to the echo server:

```bash
echo "
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: echo
  annotations:
    konghq.com/strip-path: 'true'
spec:
  ingressClassName: kong
  rules:
  - host: kong.example
    http:
      paths:
      - path: /echo
        pathType: ImplementationSpecific
        backend:
          service:
            name: echo
            port:
              number: 1027
" | kubectl apply -f -
```

Let's test our ingress object.

```bash
$ curl -i localhost/echo -H "Host: kong.example"

HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 133
Connection: keep-alive
Date: Wed, 03 May 2023 01:59:25 GMT
X-Kong-Upstream-Latency: 1
X-Kong-Proxy-Latency: 1
Via: kong/3.2.2

Welcome, you are connected to node minikube.
Running on Pod echo-f4fdf987c-qdv7s.
In namespace default.
With IP address 10.244.0.6.
```

## Next

**Note:** Read more about KIC and different use cases in official
[documentation](https://docs.konghq.com/kubernetes-ingress-controller/latest/).
