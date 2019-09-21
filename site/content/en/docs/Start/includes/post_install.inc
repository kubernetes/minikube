### Getting to know Kubernetes

Once started, you can use any regular Kubernetes command to interact with your minikube cluster. For example, you can see the pod states by running:

```shell
 kubectl get po -A
```

### Increasing memory allocation

minikube only allocates 2GB of RAM by default, which is only enough for trivial deployments. For larger 
deployments, increase the memory allocation using the `--memory` flag, or make the setting persistent using:

```shell
minikube config set memory 4096
```

### Where to go next?

Visit the [examples](/docs/examples) page to get an idea of what you can do with minikube.

📣😀 **Please fill out our [fast 5-question survey](https://forms.gle/Gg3hG5ZySw8c1C24A)** so that we can learn how & why you use minikube, and what improvements we should make. Thank you! 💃🏽🎉