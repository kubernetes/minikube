## Requirements

* Run on Mac OS or Windows 
* Registry addon enabled
* Run Minikube on KIC driver.

## Start Minikube enabling registry addon

Open a shell console, and run the following command:

```bash
minikube start --addons=registry  --vm-driver={{docker | podman}}
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

* For more information, documentation ([Pushing to an in-cluster using Registry addon](https://minikube.sigs.k8s.io/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon))
