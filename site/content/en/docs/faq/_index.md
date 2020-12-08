---
title: "FAQ"
linkTitle: "FAQ"
weight: 3
description: >
  Questions that come up regularly
---

## Does minikube support IPv6?

minikube currently doesn't support IPv6. However, it is on the [roadmap]({{< ref "/docs/contrib/roadmap.en.md" >}}).

## How can I prevent password prompts on Linux?

The easiest approach is to use the `docker` driver, as the backend service always runs as `root`.

`none` users may want to try `CHANGE_MINIKUBE_NONE_USER=true`,  where kubectl and such will still work: [see environment variables]({{< ref "/docs/handbook/config.md#environment-variables" >}})

Alternatively, configure `sudo` to never prompt for the commands issued by minikube.

## How to ignore system verification?

minikube's bootstrapper, [Kubeadm](https://github.com/kubernetes/kubeadm) verifies a list of features on the host system before installing Kubernetes. in case you get this error, and you still want to try minikube anyways despite your system's limitation you can skip the verification by starting minikube with this extra option:

```shell
minikube start --extra-config kubeadm.ignore-preflight-errors=SystemVerification
```

## what is the resource allocation for Knative Setup using minikube?

Please allocate sufficient resources for Knative setup using minikube, especially when you run a minikube cluster on your local machine. We recommend allocating at least 6 CPUs and 8G memory.

```shell
minikube start --cpus 6 --memory 8000
```

## Do I need to install kubectl locally?

No, minikube comes with built-in kubectl [see minikube's kubectl documentation]({{< ref "docs/handbook/kubectl.md" >}}). 
