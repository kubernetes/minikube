---
title: "Troubleshooting"
weight: 20
description: >
  How to troubleshoot minikube issues
---

## Enabling debug logs

Pass `--alsologtostderr` to minikube commands to see detailed log output output. To increase the log verbosity, you can use:

* `-v=1`: verbose messages
* `-v=2`: really verbose messages
* `-v=8`: more log messages than you can possibly handle.

Example:

`minikube start --alsologtostderr --v=2` will start minikube and output all the important debug logs to stderr.

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

