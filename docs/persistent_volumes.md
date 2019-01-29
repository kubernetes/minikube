## Persistent Volumes
Minikube supports [PersistentVolumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) of type `hostPath` out of the box.  These PersistentVolumes are mapped to a directory inside the running Minikube instance (usually a VM, unless you use `--vm-driver=none`.

## A note on VMs

The Minikube VM boots into a tmpfs, so most directories will not be persisted across reboots (`minikube stop`).
However, Minikube is configured to persist files stored under the following directories in the Minikube VM:

* `/data`
* `/var/lib/minikube`
* `/var/lib/docker`
* `/tmp/hostpath_pv`
* `/tmp/hostpath-provisioner`

Here is an example PersistentVolume config to persist data in the '/data' directory:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv0001
spec:
  accessModes:
    - ReadWriteOnce
  capacity:
    storage: 5Gi
  hostPath:
    path: /data/pv0001/
```

You can also achieve persistence by creating a PV in a mounted host folder.

## Dynamic provisioning and CSI

In addition, minikube implements a very simple, canonical implementation of CSI as a controller that runs alongside its deployment.  This manages provisioning of  *hostPath* volumes via the CSI interface (rather then via the previous, in-tree hostPath provider).  

The default CSI [Storage Provisioner Controller](https://github.com/kubernetes/minikube/blob/master/pkg/storage/storage_provisioner.go) is managed internally, in the minikube codebase.  This controller provides pods with dynamically, CSI managed storage, which is a good way to experiment with CSI as well as to test your pod's behaviour when persistent storage is mapped to it.  You can see the running storage controller if you check the `kube-system` namespace.

