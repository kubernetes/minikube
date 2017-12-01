## Caching Images

Minikube supports caching non-minikube images using the `minikube cache` command. Images can be added to the cache by running `minikube cache add <img>`, and deleted by running `minikube cache delete <img>`.

Images in the cache will be loaded on `minikube start`.