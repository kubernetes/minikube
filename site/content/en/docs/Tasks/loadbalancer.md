---
title: "LoadBalancer access"
linkTitle: "LoadBalancer access"
weight: 6
date: 2018-08-02
description: >
  How to access a LoadBalancer service in minikube
---

## Overview

A LoadBalancer service is the standard way to expose a service to the internet. With this method, each service gets its own IP address.


## Using `minikube tunnel`

Services of type `LoadBalancer` can be exposed via the `minikube tunnel` command. It will run in a separate terminal until Ctrl-C is hit.

## Example

#### Run tunnel in a separate terminal
it will ask for password.

```
minikube tunnel
```

`minikube tunnel` runs as a separate daemon, creating a network route on the host to the service CIDR of the cluster using the cluster's IP address as a gateway.  The tunnel command exposes the external IP directly to any program running on the host operating system.


<details>
<summary>
tunnel output example
</summary>
<pre>
Password: 
Status:	
	machine: minikube
	pid: 39087
	route: 10.96.0.0/12 -> 192.168.64.194
	minikube: Running
	services: [hello-minikube]
    errors: 
		minikube: no errors
		router: no errors
		loadbalancer emulator: no errors
...
...
...
</pre>
</details>


#### Create a kubernetes deployment 
```
kubectl create deployment hello-minikube1 --image=k8s.gcr.io/echoserver:1.4
```
#### Create a kubernetes service type LoadBalancer
```
kubectl expose deployment hello-minikube1 --type=LoadBalancer --port=8080
```

### Check external IP 
```
kubectl get svc
```
<pre>
$ kc get svc
NAME              TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)          AGE
hello-minikube1   LoadBalancer   10.96.184.178   10.96.184.178   8080:30791/TCP   40s
</pre>


note that without minikube tunnel, kubernetes would be showing external IP as "pending".

### Try in your browser
open in your browser (make sure there is no proxy set)
```
http://REPLACE_WITH_EXTERNAL_IP:8080
```


Each service will get its own external ip.

----
### DNS resolution (experimental)

If you are on macOS, the tunnel command also allows DNS resolution for Kubernetes services from the host.

### Cleaning up orphaned routes

If the `minikube tunnel` shuts down in an abrupt manner, it may leave orphaned network routes on your system. If this happens, the ~/.minikube/tunnels.json file will contain an entry for that tunnel. To remove orphaned routes, run:

````shell
minikube tunnel --cleanup
````

### Avoiding password prompts

Adding a route requires root privileges for the user, and thus there are differences in how to run `minikube tunnel` depending on the OS. If you want to avoid entering the root password, consider setting NOPASSWD for "ip" and "route" commands:

<https://superuser.com/questions/1328452/sudoers-nopasswd-for-single-executable-but-allowing-others>
