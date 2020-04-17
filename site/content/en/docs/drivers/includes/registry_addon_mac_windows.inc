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
* Run Minikube on KIC driver (see table).

## Start Minikube enabling registry addon

Open a shell console, and run the following command:

```bash
minikube start --addons=registry  --vm-driver={{docker | podman}}
```

## 1. Usage Registry for drivers that dont need port forward

Build docker image and tag it appropriately:

```shell
docker build --tag $(minikube ip):5000 <container_id>)/test-img .
```
Push docker image to minikube registry:

```shell
docker push $(minikube ip):5000 <container_id>/test-img
```

## 2. Use Registry for docker on mac and windows

Build docker image and tag it appropriately:

```shell
docker build --tag $(minikube ip):$(docker inspect --format '{{(index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort}}' <container_id>)/test-img .
```
Push docker image to minikube registry:

```shell
docker push $(minikube ip):$(docker inspect --format '{{(index (index .NetworkSettings.Ports "5000/tcp") 0).HostPort}}' <container_id>/test-img
```

After the image is pushed, refer to it by `localhost:5000/{name}` in kubectl specs.

* For more information, documentation ([Pushing to an in-cluster using Registry addon](https://minikube.sigs.k8s.io/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon))