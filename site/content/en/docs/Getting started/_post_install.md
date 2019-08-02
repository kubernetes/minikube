### Getting to know Kubernetes

Once started, you can use any regular Kubernetes command to interact with your minikube cluster. For example, you can see the pod states by running:

```shell
 kubectl get po -A
```

### Increasing memory allocation

minikube only allocates a 2GB of RAM to Kubernetes, which is only enough for basic deployments. If you run into stability issues, increase this value if your system has the resources available. You will need to recreate the VM using `minikube delete` for this to take effect.

```shell
minikube config set memory 4096
```

### Where to go next?

Visit the [examples](/docs/examples) page to get an idea of what you can do with minikube.
