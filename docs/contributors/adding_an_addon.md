# Adding a New Addon

To add a new addon to minikube the following steps are required:

* For the new addon's .yaml file(s):
  * Put the required .yaml files for the addon in the `minikube/deploy/addons` directory.
  * Add the `kubernetes.io/minikube-addons: <NEW_ADDON_NAME>` label to each piece of the addon (ReplicationController, Service, etc.)
  * Also, `addonmanager.kubernetes.io/mode` annotation is needed so that your resources are picked up by the `addon-manager` minikube addon.
  * In order to have `minikube addons open <NEW_ADDON_NAME>` work properly, the `kubernetes.io/minikube-addons-endpoint: <NEW_ADDON_NAME>` label must be added to the appropriate endpoint service (what the user would want to open/interact with).  This service must be of type NodePort.

* To add the addon into minikube commands/VM:
  * Add the addon with appropriate fields filled into the `Addon` dictionary, see this [commit](https://github.com/kubernetes/minikube/commit/41998bdad0a5543d6b15b86b0862233e3204fab6#diff-e2da306d559e3f019987acc38431a3e8R133) and example.

  ```go
  // cmd/minikube/cmd/config/config.go
  var settings = []Setting{
    ...,
    // add other addon setting
    {
      name:        "efk",
      set:         SetBool,
      validations: []setFn{IsValidAddon},
      callbacks:   []setFn{EnableOrDisableAddon},
    },
  }
  ```

  * Add the addon to settings list, see this [commit](https://github.com/kubernetes/minikube/commit/41998bdad0a5543d6b15b86b0862233e3204fab6#diff-07ad0c54f98b231e68537d908a214659R89) and example.

  ```go
  // pkg/minikube/assets/addons.go
  var Addons = map[string]*Addon{
    ...,
    // add other addon asset
    "efk": NewAddon([]*BinDataAsset{
      NewBinDataAsset(
        "deploy/addons/efk/efk-configmap.yaml",
        constants.AddonsPath,
        "efk-configmap.yaml",
        "0640"),
      NewBinDataAsset(
        "deploy/addons/efk/efk-rc.yaml",
        constants.AddonsPath,
        "efk-rc.yaml",
        "0640"),
      NewBinDataAsset(
        "deploy/addons/efk/efk-svc.yaml",
        constants.AddonsPath,
        "efk-svc.yaml",
        "0640"),
    }, false, "efk"),
  }
  ```

* Rebuild minikube using make out/minikube.  This will put the addon's .yaml binary files into the minikube binary using go-bindata.
* Test addon using `minikube addons enable <NEW_ADDON_NAME>` command to start service.
