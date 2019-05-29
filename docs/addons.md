# Add-ons

Minikube has a set of built in addons that can be used enabled, disabled, and opened inside of the local k8s environment. Below is an example of this functionality for the `heapster` addon:

```shell
$ minikube addons list
- registry: disabled
- registry-creds: disabled
- freshpod: disabled
- addon-manager: enabled
- dashboard: enabled
- heapster: disabled
- efk: disabled
- ingress: disabled
- default-storageclass: enabled
- storage-provisioner: enabled
- storage-provisioner-gluster: disabled
- nvidia-driver-installer: disabled
- nvidia-gpu-device-plugin: disabled

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
* [Heapster](https://github.com/kubernetes/heapster): [Troubleshooting Guide](https://github.com/kubernetes/heapster/blob/master/docs/influxdb.md) Note:You will need to login to Grafana as admin/admin in order to access the console
* [EFK](https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/fluentd-elasticsearch)
* [Registry](https://github.com/kubernetes/minikube/tree/master/deploy/addons/registry)
* [Registry Credentials](https://github.com/upmc-enterprises/registry-creds)
* [Ingress](https://github.com/kubernetes/ingress-nginx)
* [Freshpod](https://github.com/GoogleCloudPlatform/freshpod)
* [nvidia-driver-installer](https://github.com/GoogleCloudPlatform/container-engine-accelerators/tree/master/nvidia-driver-installer/minikube)
* [nvidia-gpu-device-plugin](https://github.com/GoogleCloudPlatform/container-engine-accelerators/tree/master/cmd/nvidia_gpu)
* [logviewer](https://github.com/ivans3/minikube-log-viewer)
* [gvisor](../deploy/addons/gvisor/README.md)
* [storage-provisioner-gluster](../deploy/addons/storage-provisioner-gluster/README.md)

If you would like to have minikube properly start/restart custom addons, place the addon(s) you wish to be launched with minikube in the `.minikube/addons` directory. Addons in this folder will be moved to the minikube VM and launched each time minikube is started/restarted.

If you have a request for an addon in minikube, please open an issue with the name and preferably a link to the addon with a description of its purpose and why it should be added.  You can also attempt to add the addon to minikube by following the guide at [Adding an Addon](contributors/adding_an_addon.md)

**Note:** If you want to have a look at the default configuration for the addons, see [deploy/addons](https://github.com/kubernetes/minikube/tree/master/deploy/addons).
