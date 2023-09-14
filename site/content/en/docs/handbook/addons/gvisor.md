---
title: "Using the gVisor Addon"
linkTitle: "gVisor"
weight: 1
date: 2018-01-02
---

## gVisor Addon
[gVisor](https://gvisor.dev/), a sandboxed container runtime, allows users to securely run pods with untrusted workloads within minikube.

### Starting minikube
gVisor depends on the containerd runtime to run in minikube.
When starting minikube, specify the following flags, along with any additional desired flags:

```shell
$ minikube start --container-runtime=containerd  \
    --docker-opt containerd=/var/run/containerd/containerd.sock
```

### Enabling gVisor
To enable this addon, simply run:

```
$ minikube addons enable gvisor
```

Within one minute, the addon manager should pick up the change and you should
see the `gvisor` pod and `gvisor` [Runtime Class](https://kubernetes.io/docs/concepts/containers/runtime-class/):

```
$ kubectl get pod,runtimeclass gvisor -n kube-system
NAME         READY   STATUS    RESTARTS   AGE
pod/gvisor   1/1     Running   0          2m52s

NAME                              CREATED AT
runtimeclass.node.k8s.io/gvisor   2019-06-15T04:35:09Z
```

Once the pod has status `Running`, gVisor is enabled in minikube.

### Running pods in gVisor

To run a pod in gVisor, add the `gvisor` runtime class to the Pod spec in your
Kubernetes yaml:

```
runtimeClassName: gvisor
```

An example Pod is shown below:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-untrusted
spec:
  runtimeClassName: gvisor
  containers:
  - name: nginx
    image: nginx
```

### Disabling gVisor

To disable gVisor, run:

```
$ minikube addons disable gvisor
```

Within one minute, the addon manager should pick up the change.
Once the `gvisor` pod has status `Terminating`, or has been deleted, the gvisor addon should be disabled.

```
$ kubectl get pod gvisor -n kube-system
NAME      READY     STATUS        RESTARTS   AGE
gvisor    1/1       Terminating   0          5m
```

_Note: Once gVisor is disabled, any pod with the `gvisor` Runtime Class will fail with a FailedCreatePodSandBox error._
