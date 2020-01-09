## istio Addon
[istio](https://istio.io/docs/setup/getting-started/) - Cloud platforms provide a wealth of benefits for the organizations that use them.

### Enabling istio
Propose to startup minikube with at least 8192 MB of memory and 4 CPUs to enable istio.
To enable this addon, simply run:

```shell script
minikube addons enable istio
```

In a minute or so istio default components will be installed into your cluster. You could run `kubectl get po -n istio-system` to see the progress for istio installation.

### Testing installation

```shell script
kubectl get po -n istio-system
```

If everything went well you shouldn't get any errors about istio being installed in your cluster. If you haven't deployed any releases `kubectl get po -n istio-system` won't return anything.

### Deprecation of istio
To disable this addon, simply run:
```shell script
minikube addons disable istio
```
