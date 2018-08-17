## Persistent Volumes
Minikube supports [PersistentVolumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) of type `hostPath`.
These PersistentVolumes are mapped to a directory inside the Minikube VM.

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
