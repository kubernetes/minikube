# Alternative runtimes

## Using CRI-O

To use [CRI-O](https://github.com/kubernetes-sigs/cri-o) as the container runtime, run:

```shell
$ minikube start --container-runtime=cri-o
```

Or you can use the extended version:

```shell
$ minikube start --container-runtime=cri-o \
    --network-plugin=cni \
    --enable-default-cni \
    --cri-socket=/var/run/crio/crio.sock \
    --extra-config=kubelet.container-runtime=remote \
    --extra-config=kubelet.container-runtime-endpoint=unix:///var/run/crio/crio.sock \
    --extra-config=kubelet.image-service-endpoint=unix:///var/run/crio/crio.sock
```

## Using containerd

To use [containerd](https://github.com/containerd/containerd) as the container runtime, run:

```shell
$ minikube start --container-runtime=containerd
```

Or you can use the extended version:

```shell
$ minikube start --container-runtime=containerd \
    --network-plugin=cni \
    --enable-default-cni \
    --cri-socket=/run/containerd/containerd.sock \
    --extra-config=kubelet.container-runtime=remote \
    --extra-config=kubelet.container-runtime-endpoint=unix:///run/containerd/containerd.sock \
    --extra-config=kubelet.image-service-endpoint=unix:///run/containerd/containerd.sock
```
