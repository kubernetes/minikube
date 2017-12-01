### Run localkube in a docker container (experimental)

**Warning:** This is experimental code at the moment.

#### How to build
From root minikube/ directory run:
```console
$ make localkube-dind-image #optional env-vars: TAG=LOCALKUBE_VERSION REGISTRY=gcr.io/k8s-minikube
```

#### How to run

##### Linux
```console
$ docker run -it \
  --privileged \
  -p 127.0.0.1:8080:8080 \
  -v /boot:/boot \
  -v /lib/modules:/lib/modules \
  gcr.io/k8s-minikube/localkube-dind-image:v1.7.0 \
  /start.sh
```

Then to setup `kubectl` to use this cluster:
```console
kubectl config set-cluster localkube-image --server=http://127.0.0.1:8080 --api-version=v1
kubectl config set-context localkube-image --cluster=localkube-image
kubectl config use-context localkube-image
```
Now `kubectl` should be configured to properly access your local k8s environment
