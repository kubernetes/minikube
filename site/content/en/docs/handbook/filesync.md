---
title: "File Sync"
weight: 12
description: >
  How to sync files into minikube
aliases:
  - /docs/tasks/sync/
  - /Handbook/sync/
---

## Built-in sync

minikube has a built-in file sync mechanism, but it only syncs when `minikube start` is run, though before Kubernetes is started. Examples where this may be useful are custom versions of system or Kubernetes configuration files, such as:

- DNS configuration
- SSL certificates
- Kubernetes service metadata

### Adding files

Place files to be synced in `$MINIKUBE_HOME/files`

For example, running the following will result in the deployment of a custom /etc/resolv.conf:

```shell
mkdir -p ~/.minikube/files/etc
echo nameserver 8.8.8.8 > ~/.minikube/files/etc/resolv.conf
minikube start
```

## Other approaches

With a bit of work, one could setup [Syncthing](https://syncthing.net) between the host and the guest VM for persistent file synchronization.

If you are looking for a solution tuned for iterative application development, consider using a Kubernetes tool that is known to work well with minikube:

- [Draft](https://draft.sh): see specific [minikube instructions](https://github.com/Azure/draft/blob/master/docs/install-minikube.md)
- [Okteto](https://github.com/okteto/okteto)
- [Skaffold](https://github.com/GoogleContainerTools/skaffold)
