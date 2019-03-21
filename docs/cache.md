# Caching Images

Minikube supports caching non-minikube images using the `minikube cache` command. Images can be added to the cache by running `minikube cache add <img>`, and deleted by running `minikube cache delete <img>`.

Images in the cache will be loaded on `minikube start`. If you want to list all available cached images, you can use `minikube cache list` command to list. Below is an example of this functionality:

```shell
# cache a image into $HOME/.minikube/cache/images
$ minikube cache add ubuntu:16.04
$ minikube cache add redis:3

# list cached images
$ minikube cache list
redis:3
ubuntu:16.04

# delete cached images
$ minikube cache delete ubuntu:16.04
$ minikube cache delete $(minikube cache list)
```
