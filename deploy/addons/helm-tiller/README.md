## helm-tiller Addon
[Kubernetes Helm](https://helm.sh) - The Kubernetes Package Manager

### Enabling helm-tiller
To enable this addon, simply run:

```shell script
minikube addons enable helm-tiller
```

In a minute or so tiller will be installed into your cluster. You could run `helm init` each time you create a new minikube instance or you could just enable this addon.
Each time you start a new minikube instance, tiller will be automatically installed. 

### Testing installation

```shell script
helm ls
```

If everything went well you shouldn't get any errors about tiller being installed in your cluster. If you haven't deployed any releases `helm ls` won't return anything.

### Deprecation of Tiller
When tiller is finally deprecated this addon won't be necessary anymore. If your version of helm doesn't use tiller, you don't need this addon.
