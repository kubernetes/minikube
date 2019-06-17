# Dashboard

Minikube supports the [Kubernetes Dashboard](https://github.com/kubernetes/dashboard) out of the box.

## Accessing the UI

To access the dashboard:

```shell
minikube dashboard
```

This will enable the dashboard add-on, and open the proxy in the default web browser.

To stop the proxy (leaves the dashboard running), abort the started process (`Ctrl+C`).

## Individual steps

If the automatic command doesn't work for you for some reason, here are the steps:

```console
$ minikube addons enable dashboard
✅  dashboard was successfully enabled
```

If you have your kubernetes client configured for minikube, you can start the proxy:

```console
$ kubectl --context minikube proxy
Starting to serve on 127.0.0.1:8001
```

Access the dashboard at:

<http://localhost:8001/api/v1/namespaces/kube-system/services/http:kubernetes-dashboard:/proxy/>

For additional information, see [this page](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/).
