### Run localkube in a docker container (experimental)

**Warning:** This is experimental code at the moment.

#### How to build
From root minikube/ directory run:
```console
$ make localkube-image #optional env-vars: TAG=LOCALKUBE_VERSION REGISTRY=gcr.io/k8s-minikube
```

#### How to run

##### Linux
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
    gcr.io/k8s-minikube/localkube-image:${LOCALKUBE_VERSION:-v1.5.3} \
    /localkube start \
    --apiserver-insecure-address=0.0.0.0 \
    --apiserver-insecure-port=8080 \
    --logtostderr=true \
    --containerized
```

##### Docker for Mac/Windows
```console
# Fix mounting, need to run every time Docker VM boots
$ docker run --rm --volume=/:/rootfs:rw --pid=host --privileged gcr.io/k8s-minikube/localkube-image:${LOCALKUBE_VERSION:-v1.5.3} nsenter --mount=/proc/1/ns/mnt sh -c "if ! df | grep /var/lib/kubelet > /dev/null; then mkdir -p /var/lib/kubelet && mount --bind /var/lib/kubelet /var/lib/kubelet && mount --make-shared /var/lib/kubelet; fi"
$ docker run -d \
    --volume=/sys:/sys:ro \
    --volume=/var/lib/docker:/var/lib/docker:rw \
    --volume=/var/lib/kubelet:/var/lib/kubelet:shared \
    --volume=/var/run:/var/run:rw \
    --volume="${HOME}"/.minikube/certs:/var/lib/localkube/certs:rw \
    --volume="${HOME}"/.minikube/etcd:/var/lib/localkube/etcd:rw \
    --volume="${HOME}"/.minikube/manifests:/etc/kubernetes/manifests:ro \
    --name=minikube \
    --net=host \
    --pid=host \
    --privileged \
    gcr.io/k8s-minikube/localkube-amd64:${LOCALKUBE_VERSION:-v1.5.3} \
    /localkube start \
    --apiserver-insecure-address=127.0.0.1 \
    --apiserver-insecure-port=8000 \
    --logtostderr=true
# Fix loopback
$ docker exec minikube sh -c 'echo 127.0.0.1 ${HOSTNAME} >> /etc/hosts'
```

###### Manifests
Copy manifests to ${HOME}/.minikube/manifests
* kube-addon-manager
Put addons in ${HOME}/.minikube/addons
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: kube-addon-manager
  namespace: kube-system
  labels:
    component: kube-addon-manager
    version: v6.4
    kubernetes.io/minikube-addons: addon-manager
spec:
  hostNetwork: true
  containers:
  - name: kube-addon-manager
    image: gcr.io/google-containers/kube-addon-manager:v6.4-beta.2
    imagePullPolicy: IfNotPresent
    resources:
      requests:
        cpu: 5m
        memory: 50Mi
    volumeMounts:
    - mountPath: /etc/kubernetes/addons
      name: addons
      readOnly: true
    - mountPath: /etc/kubernetes/admission-controls
      name: admission-controls
      readOnly: true
  volumes:
  - hostPath:
      path: ${HOME}/.minikube/addons
    name: addons
  - hostPath:
      path: ${HOME}/.minikube/admission-controls
    name: admission-controls
```

* Docker for Mac/Windows does not support port forwarding with host networking, so kube-apiserver traffic will need to be proxied. The following pod will do that for us:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: socat
  namespace: kube-system
spec:
  containers:
    - name: busybox
      image: calpicow/alpine-socat:latest
      imagePullPolicy: IfNotPresent
      command:
        - socat
        - tcp-listen:8001,reuseaddr,fork
        - tcp:10.0.0.1:443
      ports:
        - containerPort: 8001
          hostPort: 8001
```

###### Issues
* We need to enable `--insecure-skip-tls-verify` since localkube does not generate a cert for localhost

#### kubectl configuration
Then to setup `kubectl` to use this cluster:
```console
kubectl config set-cluster localkube-image --server=http://127.0.0.1:8080 --api-version=v1
kubectl config set-context localkube-image --cluster=localkube-image
kubectl config use-context localkube-image
```
Or for Mac/Windows:
```console
kubectl config set-cluster localkube-image --server=https://localhost:8001 --insecure-skip-tls-verify=true
kubectl config set-credentials localkube-image --certificate-authority="${HOME}"/.minikube/certs/ca.crt --client-certificate="${HOME}"/.minikube/certs/apiserver.crt --client-key="${HOME}"/.minikube/certs/apiserver.key
kubectl config set-context localkube-image --cluster=localkube-image --user=localkube-image
```
Now `kubectl` should be configured to properly access your local k8s environment
