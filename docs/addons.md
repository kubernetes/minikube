## Add-ons

Minikube has a set of built in addons that can be used enabled, disabled, and opened inside of the local k8s environment. Below is an example of this functionality for the `heapster` addon:
```shell
$ minikube addons list
- kube-dns: enabled
- registry: disabled
- registry-creds: disabled
- freshpod: disabled
- addon-manager: enabled
- dashboard: enabled
- coredns: disabled
- heapster: disabled
- efk: disabled
- ingress: disabled
- default-storageclass: enabled
- storage-provisioner: enabled

# minikube must be running for these commands to take effect
$ minikube addons enable heapster
heapster was successfully enabled

$ minikube addons open heapster # This will open grafana (interacting w/ heapster) in the browser
Waiting, endpoint for service is not ready yet...
Waiting, endpoint for service is not ready yet...
Created new window in existing browser session.
```
The currently supported addons include:

* [Kubernetes Dashboard](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dashboard)
* [Kube-dns](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/dns)
* [Heapster](https://github.com/kubernetes/heapster): [Troubleshooting Guide](https://github.com/kubernetes/heapster/blob/master/docs/influxdb.md) Note:You will need to login to Grafana as admin/admin in order to access the console
* [EFK](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/fluentd-elasticsearch)
* [Registry](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/registry)
* [Registry Credentials](https://github.com/upmc-enterprises/registry-creds)
* [CoreDNS](https://github.com/coredns/deployment/tree/master/kubernetes)
* [Ingress](https://github.com/kubernetes/ingress-nginx)
* [Freshpod](https://github.com/GoogleCloudPlatform/freshpod)

If you would like to have minikube properly start/restart custom addons, place the addon(s) you wish to be launched with minikube in the `.minikube/addons` directory. Addons in this folder will be moved to the minikube VM and launched each time minikube is started/restarted.

If you have a request for an addon in minikube, please open an issue with the name and preferably a link to the addon with a description of its purpose and why it should be added.  You can also attempt to add the addon to minikube by following the guide at [Adding an Addon](contributors/adding_an_addon.md)
