#### Adding a New Addon
To add a new addon to minikube the following steps are required:

* For the new addon's .yaml file(s):
  * Put the required .yaml files for the addon in the minikube/deploy/addons directory.
  * Add the `kubernetes.io/minikube-addons: <NEW_ADDON_NAME>` label to each piece of the addon (ReplicationController, Service, etc.)
  * In order to have `minikube open addons <NEW_ADDON_NAME>` work properly, the `kubernetes.io/minikube-addons-endpoint: <NEW_ADDON_NAME>` label must be added to the appropriate endpoint service (what the user would want to open/interact with).  This service must be of type NodePort.

* To add the addon into minikube commands/VM:
  * Add the addon with appropriate fields filled into the `Addon` dictionary, see this [commit](https://github.com/kubernetes/minikube/commit/41998bdad0a5543d6b15b86b0862233e3204fab6#diff-e2da306d559e3f019987acc38431a3e8R133).
  * Add the addon to settings list, see this [commit](https://github.com/kubernetes/minikube/commit/41998bdad0a5543d6b15b86b0862233e3204fab6#diff-07ad0c54f98b231e68537d908a214659R89).
* Rebuild minikube using make out/minikube.  This will put the addon .yaml binary files into the minikube binary using go-bindata.
