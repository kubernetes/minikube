## gVisor Addon
[gVisor](https://github.com/google/gvisor/blob/master/README.md), a sandboxed container runtime, allows users to securely run pods with untrusted workloads within Minikube.

### Starting Minikube
gVisor depends on the containerd runtime to run in Minikube.
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

Within one minute, the addon manager should pick up the change and you should see the `gvisor` pod:

```
$ kubectl get pod gvisor -n kube-system
NAME      READY     STATUS    RESTARTS   AGE
gvisor    1/1       Running   0          3m
```

Once the pod has status `Running`, gVisor is enabled in Minikube. 

### Running pods in gVisor
To run a pod in gVisor, add this annotation to the Kubernetes yaml:

```
io.kubernetes.cri.untrusted-workload: "true"
```

An example Pod is shown below:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-untrusted
  annotations:
    io.kubernetes.cri.untrusted-workload: "true"
spec:
  containers:
  - name: nginx
    image: nginx
```

_Note: this annotation will not be necessary once the RuntimeClass Kubernetes feature is available broadly._

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

_Note: Once gVisor is disabled, any pod with the `io.kubernetes.cri.untrusted-workload` annotation will fail with a FailedCreatePodSandBox error._
