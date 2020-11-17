# Static IP for minikube containers KIC Drivers (docker/podman)

design by @medyagh

## Prototype [PR 9294](https://github.com/kubernetes/minikube/pull/9294):

Each minikube cluster will get it is own network, instead of using docker’s owned default network which is unmodifiable by the user.

on “minikube start”, minikube will try to create a docker network at a default subnet
If a user happens to have an existing network in that exact subnet, minikube will try X times, by incrementing the sunbet by 10 each time.

once the network is created, it will calculate container IP inside the network by this formula:

```
gateway IP = first ip of the subnet
container IP = gateway IP + 1 + container index
```

container_index is calculated from this formula:

default is 1 for single node (and for multinde is the order the machine created based on it is name, for example the index for minikube-m02 will be 2)



### Alternative design :

once network is found or created, minikube can look up list of all the container IPs in the network, and try to 


### Considerations:

if can’t create network, should fall back to docker’s default bridge IP
The existing running clusters should not break by using newer version of minikube


### Advantages:

since we own the created network we can assign and calculate the IP.
By avoiding using the default docker bridge we can isolate minikube network problems to not be affected by user’s other containers.
User will have the same minikube IP across computer restart and hibernations, which will save user time on the next start. (each time IP changes there will be a tax of 22 seconds of setting up the cluster by kubeadm)
