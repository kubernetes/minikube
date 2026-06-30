---
title: "Using the Traefik Ingress Controller Addon"
linkTitle: "Traefik"
weight: 1
date: 2025-06-27
---

## Overview

[Traefik](https://traefik.io/traefik/) is a modern, cloud-native reverse proxy and ingress controller that makes
deploying and routing services easy. The minikube Traefik addon installs Traefik via its
[official Helm chart](https://github.com/traefik/traefik-helm-chart) into the `kube-system` namespace.

> **Note:** This is a Helm-based addon. minikube automatically installs Helm inside the node and manages the
> Traefik lifecycle (install, upgrade, uninstall) via `helm upgrade --install` and `helm uninstall`.

## Prerequisites

- Minikube v1.38.0 or higher
- kubectl

## Enabling the Addon

```shell script
minikube addons enable traefik
```

The first time you enable it, minikube will install Helm (if not present) and then deploy Traefik.
This may take a few minutes while images are pulled.

```
🌟  The 'traefik' addon is enabled
💡  To verify Traefik is running:
        kubectl get pods -n kube-system -l app.kubernetes.io/name=traefik
💡  To access the Traefik dashboard:
        kubectl port-forward -n kube-system deployment/traefik 9000:8080
    Then open http://localhost:9000/dashboard/ in your browser.
💡  To expose Traefik's LoadBalancer service, run in a separate terminal:
        minikube tunnel
```

## Verifying the Installation

Check that the Traefik pod is running:

```shell script
kubectl get pods -n kube-system -l app.kubernetes.io/name=traefik
```
Example output:
```
NAME                       READY   STATUS    RESTARTS   AGE
traefik-6b5d4cb8f7-x9z2k   1/1     Running   0          2m
```

Check the Traefik service:

```shell script
kubectl get svc -n kube-system -l app.kubernetes.io/name=traefik
```
Example output:
```
NAME      TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
traefik   LoadBalancer   10.96.123.45   <pending>     80:31080/TCP,443:31443/TCP   2m
```

## Accessing the Traefik Dashboard

Traefik ships with a built-in dashboard that shows your routers, services, and middleware at a glance.

There are two ways to access the Traefik dashboard from your host machine:

### Option A: Using `minikube service`

Because the Traefik service is of type `LoadBalancer` and exposes port `8080` (admin/dashboard port), you can run:

```shell script
minikube service -n kube-system traefik
```

It will print a table of URLs matching the exposed service ports. Look for the row with `traefik/8080` under the **TARGET PORT** column:

```
|   NAMESPACE   |  NAME   |  TARGET PORT  |            URL            |
| ------------- | ------- | ------------- | ------------------------- |
|  kube-system  | traefik | traefik/8080  | http://192.168.49.2:31781 |
|               |         | web/80        | http://192.168.49.2:30419 |
|               |         | websecure/443 | http://192.168.49.2:31862 |
```

Open the URL corresponding to the `traefik/8080` target port (in the example above, `http://192.168.49.2:31781`) and append `/dashboard/` to the end (e.g. `http://192.168.49.2:31781/dashboard/`).


### Option B: Using `kubectl port-forward`

```shell script
kubectl port-forward -n kube-system deployment/traefik 9000:8080
```

Then open [http://localhost:9000/dashboard/](http://localhost:9000/dashboard/) in your browser.

> **Important:** The trailing slash in `/dashboard/` is required for both options, without it you'll get a 404.

### Dashboard Overview

Below is the main overview of the Traefik dashboard displaying the active HTTP features and service statistics:

![Traefik Dashboard](/images/addons/traefik_dashboard.png)

### HTTP Routers Status

Below is the list of active HTTP routers showing the status and routing rules defined for services:

![Traefik HTTP Routers](/images/addons/traefik_routers.png)


## Usage Example: Deploying an App with Ingress

This example deploys a simple web application and exposes it through Traefik using a standard Kubernetes
[Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) resource.

### Step 1: Deploy the application

Create a file called `traefik-app.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: traefik-app
  labels:
    app: traefik-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: traefik-app
  template:
    metadata:
      labels:
        app: traefik-app
    spec:
      containers:
      - name: echo-server
        image: kicbase/echo-server:1.0
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: traefik-app
spec:
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app: traefik-app
```

Apply it:

```shell script
kubectl apply -f traefik-app.yaml
```

Wait for the pod to be ready:

```shell script
kubectl wait --for=condition=ready pod -l app=traefik-app --timeout=60s
```

Check the service and pods status:

```shell script
kubectl get svc
```
Example output:
```
NAME          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
traefik-app   ClusterIP   10.99.33.121   <none>        8080/TCP   1m
kubernetes    ClusterIP   10.96.0.1      <none>        443/TCP    5m
```

```shell script
kubectl get pods
```
Example output:
```
NAME                           READY   STATUS    RESTARTS   AGE
traefik-app-7748457cf9-gq295   1/1     Running   0          1m
```


### Step 2: Create an Ingress

Create a file called `traefik-ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: traefik-ingress
spec:
  ingressClassName: traefik
  rules:
  - host: traefik.example
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: traefik-app
            port:
              number: 8080
```

Apply it:

```shell script
kubectl apply -f traefik-ingress.yaml
```

### Step 3: Test the Ingress

From inside the minikube node:

```shell script
minikube ssh -- curl -s -H "Host: traefik.example" http://127.0.0.1
```
Example output:
```
Request served by traefik-app-7b9bf5c4d6-abc12
...
```

Or, if you're running `minikube tunnel` in another terminal:

```shell script
curl -H "Host: traefik.example" http://127.0.0.1
```
Example output:
```
Request served by traefik-app-7b9bf5c4d6-abc12
...
```


## Using `minikube tunnel`

Traefik's main service is of type `LoadBalancer`. To get an external IP assigned, run in a separate terminal:

```shell script
minikube tunnel
```

This assigns `127.0.0.1` as the external IP, allowing you to reach Traefik at `http://127.0.0.1`.

> On macOS and Windows with Docker/Podman driver, `minikube tunnel` is required for LoadBalancer access.

## Minikube-Specific Notes

- **Helm-based addon:** Unlike most minikube addons that use static YAML manifests, Traefik is installed
  via Helm. The Helm binary is automatically installed inside the minikube node.

- **Namespace:** Traefik is deployed in the `kube-system` namespace.

- **Avoid port conflicts:** If you have the `ingress` addon (nginx) enabled and Traefik enabled at the same
  time, both will try to bind ports 80 and 443. Either disable one or configure different node ports.

- **Traefik CRDs:** The Helm chart automatically installs Traefik's Custom Resource Definitions (CRDs),
  including `IngressRoute`, `Middleware`, `TLSOption`, etc. You can use these for more advanced routing
  beyond standard Kubernetes Ingress resources.

## Disabling the Addon

```shell script
minikube addons disable traefik
```

This runs `helm uninstall traefik -n kube-system` inside the node and removes all Traefik resources.

## Further Reading

- [Traefik Documentation](https://doc.traefik.io/traefik/)
- [Traefik Helm Chart](https://github.com/traefik/traefik-helm-chart)
- [Traefik Kubernetes Ingress](https://doc.traefik.io/traefik/providers/kubernetes-ingress/)
- [Traefik IngressRoute CRD](https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/)
- [Traefik Dashboard](https://doc.traefik.io/traefik/operations/dashboard/)
