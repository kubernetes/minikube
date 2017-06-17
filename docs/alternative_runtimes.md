### Using rkt container engine

To use [rkt](https://github.com/coreos/rkt) as the container runtime run:

```shell
$ minikube start \
    --network-plugin=cni \
    --container-runtime=rkt
```
