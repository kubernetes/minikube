# Persistent Volumes

Minikube supports [PersistentVolumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) of type `hostPath` out of the box.  These PersistentVolumes are mapped to a directory inside the running Minikube instance (usually a VM, unless you use `--vm-driver=none`).  For more information on how this works, read the Dynamic Provisioning section below.

## A note on mounts, persistence, and Minikube hosts

Minikube is configured to persist files stored under the following directories, which are made in the Minikube VM (or on your localhost if running on bare metal).  You may lose data from other directories on reboots.

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

In addition, minikube implements a very simple, canonical implementation of dynamic storage controller that runs alongside its deployment.  This manages provisioning of  *hostPath* volumes (rather then via the previous, in-tree hostPath provider).  

The default [Storage Provisioner Controller](https://github.com/kubernetes/minikube/blob/master/pkg/storage/storage_provisioner.go) is managed internally, in the minikube codebase, demonstrating how easy it is to plug a custom storage controller into kubernetes as a storage component of the system, and provides pods with dynamically, to test your pod's behaviour when persistent storage is mapped to it.

Note that this is not a CSI based storage provider, rather, it simply declares a PersistentVolume object of type hostpath dynamically when the controller see's that there is an outstanding storage request.
