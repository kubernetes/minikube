# minikube: Syncing files into the VM

## Syncing files during start up

As soon as a VM is created, minikube will populate the root filesystem with any files stored in $MINIKUBE_HOME (~/.minikube/files).

For example, running the following commands will result in /etc/OMG existing within the VM:

```
mkdir -p ~/.minikube/files/etc
touch ~/.minikube/files/etc/OMG
minikube start
```

This method of file synchronization can be useful for adding configuration files for apiserver, or adding HTTPS certificates.


