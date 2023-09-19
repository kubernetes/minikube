---
title: "Using the Istio Addon"
linkTitle: "Istio"
weight: 1
date: 2019-12-25
---

## istio Addon
[istio](https://istio.io/docs/setup/getting-started/) - Cloud platforms provide a wealth of benefits for the organizations that use them.

### Enable istio on minikube
Make sure to start minikube with at least 8192 MB of memory and 4 CPUs.
See official [Platform Setup](https://istio.io/docs/setup/platform-setup/) documentation.

```shell script
minikube start --memory=8192mb --cpus=4
```

To enable this addon, simply run:
```shell script
minikube addons enable istio-provisioner
minikube addons enable istio
```

In a minute or so istio default components will be installed into your cluster. You could run `kubectl get po -n istio-system` to see the progress for istio installation.

### Testing installation

```shell script
kubectl get po -n istio-system
```

If everything went well you shouldn't get any errors about istio being installed in your cluster. If you haven't deployed any releases `kubectl get po -n istio-system` won't return anything.

### Disable istio
To disable this addon, simply run:
```shell script
minikube addons disable istio-provisioner
minikube addons disable istio
```
