---
title: "Kubectl"
weight: 2
description: >
  Use kubectl inside minikube
aliases:
  - /docs/kubectl/
---

By default, [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) gets configured to access the kubernetes cluster control plane
inside minikube when the `minikube start` command is executed.

```shell
kubectl <kubectl commands>
```

However if `kubectl` is not installed locally, minikube already includes kubectl which can be used like this:

```shell
minikube kubectl -- <kubectl commands>
```

{{% tabs %}}

{{% linuxtab %}}
You can also alias kubectl for easier usage.

```shell
alias kubectl="minikube kubectl --"
```

Alternatively, you can create a symbolic link to minikube's binary named 'kubectl'.

```shell
ln -s $(which minikube) /usr/local/bin/kubectl
```

{{% /linuxtab %}}

{{% mactab %}}
You can also alias kubectl for easier usage.

```shell
alias kubectl="minikube kubectl --"
```

Alternatively, you can create a symbolic link to minikube's binary named 'kubectl'.

```shell
ln -s $(which minikube) /usr/local/bin/kubectl
```

{{% /mactab %}}

{{% windowstab %}}
You can also alias kubectl for easier usage.

Powershell.

```shell
function kubectl { minikube kubectl -- $args }
```

Command Prompt.

```shell
doskey kubectl=minikube kubectl $*
```


{{% /windowstab %}}

{{% /tabs %}}

Get pods

```shell
minikube kubectl -- get pods
```

Creating a deployment inside kubernetes cluster

```shell
minikube kubectl -- create deployment hello-minikube --image=kicbase/echo-server:1.0
```

Exposing the deployment with a NodePort service

```shell
minikube kubectl -- expose deployment hello-minikube --type=NodePort --port=8080
```

For more help

```shell
minikube kubectl -- --help
```

Documentation

<https://kubernetes.io/docs/reference/kubectl/>

### Shell autocompletion

After applying the alias or the symbolic link you can follow https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#enable-shell-autocompletion to enable shell-autocompletion.
