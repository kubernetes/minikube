### Using rkt container engine

To use [rkt](https://github.com/coreos/rkt) as the container runtime run:

```shell
$ minikube start \
    --network-plugin=cni \
    --container-runtime=rkt
```


### Using CRI-O

To use [CRI-O](https://github.com/kubernetes-incubator/cri-o) as the container runtime, run:

```shell
$ minikube start \
    --network-plugin=cni \
    --container-runtime=cri-o \
    --bootstrapper=kubeadm
```

Or you can use the extended version:

```shell
$ minikube start \
    --network-plugin=cni \
    --extra-config=kubelet.container-runtime=remote \
    --extra-config=kubelet.container-runtime-endpoint=/var/run/crio/crio.sock \
    --extra-config=kubelet.image-service-endpoint=/var/run/crio/crio.sock \
    --bootstrapper=kubeadm
```

### Using containerd

To use [containerd](https://github.com/containerd/containerd) as the container runtime, run:

```shell
$ minikube start \
    --network-plugin=cni \
    --container-runtime=containerd \
    --bootstrapper=kubeadm
```

Or you can use the extended version:

```shell
$ minikube start \
    --network-plugin=cni \
    --extra-config=kubelet.container-runtime=remote \
    --extra-config=kubelet.container-runtime-endpoint=unix:///run/containerd/containerd.sock \
    --extra-config=kubelet.image-service-endpoint=unix:///run/containerd/containerd.sock \
    --bootstrapper=kubeadm
```