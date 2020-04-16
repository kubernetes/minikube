---

## Table for different drivers.
The registry addon works in all OS, to use it depends on the driver you built your cluster with, If you are using container system (Docker, Podman) in Mac or Windows OS it has some limitation, is this case you will have to use the forwarded port for 5000.

Here is a table to help you with registry addon:

| Drivers   	| Supported Status  	|  Drivers to port forward 	|  Ports 	|
|---	|---	|---	|---	|---	|
|  [Linux drivers](/docs/drivers)	|   works fine	|  	|  5000 	|
|  [Mac drivers](/docs/drivers)	|   works fine	| [Docker](/docs/drivers/docker/)  |  random 	|
|  [Windows drivers](/docs/drivers) 	|  works fine 	| [Docker](docs/drivers/docker/)   	|  random 	|


* note1 : The default port for linux drivers is 5000.
* note2 : Drivers for Docker driver is random like (87362).

---

## Requirements

* Registry addon enabled
* Run Minikube on KIC driver (see table).

## Start Minikube enabling registry addon

Open a shell console, and run the following command:

```bash
minikube start --addons=registry  --vm-driver={{docker | podman}}
```

## Before usage Registry

Make sure the port you have to use, you can see it as a result of the previous command, or seeing the port forward running(in case of docker driver on Mac or Windows):

```shell
docker ps
```

## Usage Registry

Build docker image and tag it appropriately:

```shell
docker build --tag $(minikube ip):5000/test-img .
```
Push docker image to minikube registry:

```shell
docker push $(minikube ip):5000/test-img
```

After the image is pushed, refer to it by `localhost:5000/{name}` in kubectl specs.

* For more information, documentation ([Pushing to an in-cluster using Registry addon](https://minikube.sigs.k8s.io/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon))
