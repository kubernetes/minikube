### Run localkube in a docker container (experimental)

**Warning:** This is experimental code at the moment.

#### How to build
From root minikube/ directory run:
```console
$ make localkube-image #optional env-vars: TAG=LOCALKUBE_VERSION REGISTRY=gcr.io/k8s-minikube
```

#### How to run

```console
$ docker run -d \
    --volume=/:/rootfs:ro \
    --volume=/sys:/sys:rw \
    --volume=/var/lib/docker:/var/lib/docker:rw \
    --volume=/var/lib/kubelet:/var/lib/kubelet:rw \
    --volume=/var/run:/var/run:rw \
    --net=host \
    --pid=host \
    --privileged \
    gcr.io/k8s-minikube/localkube-amd64:LOCALKUBE_VERSION \
    /localkube start \
    --apiserver-insecure-address=0.0.0.0 \
    --apiserver-insecure-port=8080 \
    --logtostderr=true \
    --containerized
```
Then to setup `kubectl` to use this cluster:
```console
kubectl config set-cluster localkube-image --server=http://127.0.0.1:8080 --api-version=v1
kubectl config set-context localkube-image --cluster=localkube-image
kubectl config use-context localkube-image
```
Now `kubectl` should be configured to properly access your local k8s environment
