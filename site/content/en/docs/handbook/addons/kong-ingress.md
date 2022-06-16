---
title: "Using Kong Ingress Controller Addon"
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
> User-Agent: curl/7.77.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 404 Not Found
< Date: Tue, 25 Jan 2022 22:35:27 GMT
< Content-Type: application/json; charset=utf-8
< Connection: keep-alive
< Content-Length: 48
< X-Kong-Response-Latency: 0
< Server: kong/2.7.0
<
* Connection #0 to host localhost left intact
{"message":"no Route matched with those values"}%
````

## Creating Ingress object

Let's create a service.
As an example, we use `type=ExternalName` to point to https://httpbin.org

```bash
echo "
kind: Service
apiVersion: v1
metadata:
  name: proxy-to-httpbin
spec:
  ports:
  - protocol: TCP
    port: 80
  type: ExternalName
  externalName: httpbin.org
" | kubectl create -f -
```

Next, we will create the ingress object points to httpbin service.

```bash
echo '
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: proxy-from-k8s-to-httpbin
  annotations:
    konghq.com/strip-path: "true"
spec:
  ingressClassName: kong
  rules:
  - http:
      paths:
      - path: /foo
        pathType: ImplementationSpecific
        backend:
          service:
            name: proxy-to-httpbin
            port:
              number: 80
' | kubectl create -f -
```

Let's test our ingress object.

```bash
$ curl -i localhost/foo -H "Host: httpbin.org"


HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 4
Connection: keep-alive
X-App-Name:
X-App-Version: 0.2.4
Date: Tue, 25 Jan 2022 22:44:57 GMT
X-Kong-Upstream-Latency: 1
X-Kong-Proxy-Latency: 1
Via: kong/2.7.0

foo
```

## Next

**Note:** Read more about KIC and different use cases in official
[documentation](https://docs.konghq.com/kubernetes-ingress-controller/2.1.x/guides/overview/).
