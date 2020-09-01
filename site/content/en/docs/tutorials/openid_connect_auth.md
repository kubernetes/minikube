---
title: "OpenID Connect Authentication"
linkTitle: "OpenID Connect Authentication"
weight: 1
date: 2018-01-02
description: >
  Configuring minikube to use OpenID Connect Authentication
---

The `kube-apiserver` in minikube can be configured to support OpenID Connect Authentication.

Read more about OpenID Connect Authentication for Kubernetes here: <https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens>

## Configuring the API Server

Configuration values can be passed to the API server using the `--extra-config` flag on the `minikube start` command. See [configuring_kubernetes.md]({{< ref "/docs/handbook/config.md#kubernetes-configuration" >}}) for more details.

The following example configures your Minikube cluster to support RBAC and OIDC:

```shell
minikube start \
  --extra-config=apiserver.authorization-mode=RBAC \
  --extra-config=apiserver.oidc-issuer-url=https://example.com \
  --extra-config=apiserver.oidc-username-claim=email \
  --extra-config=apiserver.oidc-client-id=kubernetes-local
```

Note that as stated in the Kubernetes [documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server), for `--extra-config=apiserver.oidc-issuer-url` flag, only URLs which use the `https://` scheme are accepted. Otherwise `kube-apiserver` will not start.

## Configuring kubectl

You can use the kubectl `oidc` authenticator to create a kubeconfig as shown in the Kubernetes docs: <https://kubernetes.io/docs/reference/access-authn-authz/authentication/#option-1-oidc-authenticator>

`minikube start` already creates a kubeconfig that includes a `cluster`, in order to use it with your `oidc` authenticator kubeconfig, you can run:

```shell
kubectl config set-context kubernetes-local-oidc --cluster=minikube --user username@example.com
Context "kubernetes-local-oidc" created.
kubectl config use-context kubernetes-local-oidc
```

For the new context to work you will need to create, at the very minimum, a `Role` and a `RoleBinding` in your cluster to grant permissions to the `subjects` included in your `oidc-username-claim`.
