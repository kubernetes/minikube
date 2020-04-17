---

## Table for different drivers.
The registry addon works in all OS, to use it depends on the driver you built your cluster with, If you are using container system (Docker, Podman) in Mac or Windows OS it has some limitation, is this case you will have to use the forwarded port for 5000.

Here is a table to help you with registry addon:

| Drivers   	|  Port 	|
|---	|---	|---	|---	|---	|
|  Docker driver on Linux	| 5000 	|
|  Docker driver on Mac and windows | Random 	|
|  All other drivers | 5000 	|


* note1 : The default port for linux drivers is 5000.
* note2 : Drivers for Docker driver is random.

---

## Requirements

* Registry addon enabled
* Run minikube on KIC driver (see table).

## Start minikube enabling registry addon

Open a shell console, and run the following command:

```bash
minikube start --addons=registry  --vm-driver={{docker | podman}}
```

---
## 1. Usage Registry for drivers that dont need port forward

Build docker image and tag it appropriately:

```shell
docker build --tag $(minikube ip):5000/test-img .
```
Push docker image to minikube registry:

```shell
docker push $(minikube ip):5000/test-img
```

---
## 2. Use Registry for docker on mac and windows

Build docker image and tag it appropriately:

#### Accept pushed images on Mac

In order to make docker accept pushing images to this registry, we have to redirect port 5000 on the docker virtual machine over to port 5000 on the minikube machine. We can (ab)use docker’s network configuration to instantiate a container on the docker’s host, and run socat there:

```shell
docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000"
```

#### Accept pushing images on Windows

In order to make docker accept pushing images to this registry, we have to redirect port 5000 on the docker virtual machine over to port 5000 on the minikube machine. Unfortunately, the docker vm cannot directly see the IP address of the minikube vm. To fix this, you will have to add one more level of redirection.

Use kubectl port-forward to map your local workstation to the minikube vm

```shell
kubectl port-forward --namespace kube-system <name of the registry vm> 5000:5000
```

On your local machine you should now be able to reach the minikube registry by using curl `http://localhost:5000/v2/_catalog`

From this point we can (ab)use docker’s network configuration to instantiate a container on the docker’s host, and run socat there to redirect traffic going to the docker vm’s port 5000 to port 5000 on your host workstation.

```shell
docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:host.docker.internal:5000"
```


#### Push docker image to minikube registry:

```shell
docker push $(minikube ip):$(docker inspect --format '{{(index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort}}' <container_id>/test-img
```

After the image is pushed, refer to it by `localhost:5000/{name}` in kubectl specs.

* For more information, documentation ([Pushing to an in-cluster using Registry addon](https://minikube.sigs.k8s.io/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon))