---
title: "Troubleshooting"
linkTitle: "Troubleshoot"
weight: 20
date: 2019-08-01
description: >
  How to troubleshoot minikube issues
---

## Enabling debug logs

To debug issues with minikube (not *Kubernetes* but **minikube** itself), you can use the `-v` flag to see debug level info.  The specified values for `-v` will do the following (the values are all encompassing in that higher values will give you all lower value outputs as well):

* `--v=0` will output **INFO** level logs
* `--v=1` will output **WARNING** level logs
* `--v=2` will output **ERROR** level logs

* `--v=3` will output *libmachine* logging
* `--v=7` will output *libmachine --debug* level logging

Example:

`minikube start --v=7` will start minikube and output all the important debug logs to stdout.

## Gathering VM logs

To debug issues where Kubernetes failed to deploy, it is very useful to collect the Kubernetes pod and kernel logs:

```shell
minikube logs
```

## Viewing Pod Status

To view the deployment state of all Kubernetes pods, use:

```shell
kubectl get po -A
```

Example output:

```shell
NAMESPACE     NAME                        READY   STATUS    RESTARTS   AGE
kube-system   coredns-5c98db65d4-299md    1/1     Running   0          11m
kube-system   coredns-5c98db65d4-qlpkd    1/1     Running   0          11m
kube-system   etcd-minikube               1/1     Running   0          10m
kube-system   gvisor                      1/1     Running   0          11m
...
kube-system   storage-provisioner         1/1     Running   0          11m
```

To view more detailed information about a pod, use:

```shell
kubectl describe pod <name> -n <namespace>
```

## Debugging hung start-up

minikube will wait ~8 minutes before giving up on a Kubernetes deployment. If you want to see startup fails more immediately, consider using:

```shell
minikube logs --problems
```

This will attempt to surface known errors, such as invalid configuration flags. If nothing interesting shows up, try `minikube logs`.

