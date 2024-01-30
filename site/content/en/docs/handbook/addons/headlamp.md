---
title: "Using the Headlamp Addon"
linkTitle: "Headlamp"
weight: 1
date: 2022-06-08
---

## Headlamp Addon

[Headlamp](https://kinvolk.github.io/headlamp) is an easy-to-use and extensible Kubernetes web UI.

### Enable Headlamp on minikube

To enable this addon, simply run:
```shell script
minikube addons enable headlamp
```

Once the addon is enabled, you can access the Headlamp's web UI using the following command.
```shell script
minikube service headlamp -n headlamp
```

To authenticate in Headlamp, fetch the Authentication Token using the following command:

```shell script
export SECRET=$(kubectl get secrets --namespace headlamp -o custom-columns=":metadata.name" | grep "headlamp-token")
kubectl get secret $SECRET --namespace headlamp --template=\{\{.data.token\}\} | base64 --decode
``` 

Headlamp can display more detailed information when metrics-server is installed. To install it, run:

```shell script
minikube addons enable metrics-server	
```		

### Testing installation

```shell script
kubectl get pods -n headlamp
```

If everything went well, there should be no errors about Headlamp's installation in your minikube cluster.

### Disable headlamp

To disable this addon, simply run:

```shell script
minikube addons disable headlamp
```
