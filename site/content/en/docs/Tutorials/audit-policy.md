---
title: "Audit Policy"
linkTitle: "Audit Policy"
weight: 1
date: 2019-11-19
description: >
  Enabling audit policy for minikube
---

## Overview

[Auditing](https://kubernetes.io/docs/Handbook/debug-application-cluster/audit/) is not enabled in minikube by default.
This tutorial shows how to provide an [Audit Policy](https://kubernetes.io/docs/Handbook/debug-application-cluster/audit/#audit-policy) file to the minikube API server on startup.

## Tutorial

```shell
minikube stop

mkdir -p ~/.minikube/files/etc/ssl/certs

cat <<EOF > ~/.minikube/files/etc/ssl/certs/audit-policy.yaml
# Log all requests at the Metadata level.
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
EOF

minikube start \
  --extra-config=apiserver.audit-policy-file=/etc/ssl/certs/audit-policy.yaml \
  --extra-config=apiserver.audit-log-path=-

kubectl logs kube-apiserver-minikube -n  kube-system | grep audit.k8s.io/v1
```

The [Audit Policy](https://kubernetes.io/docs/Handbook/debug-application-cluster/audit/#audit-policy) used in this tutorial is very minimal and quite verbose. As a next step you might want to finetune the `audit-policy.yaml` file. To get the changes applied you need to stop and start minikube. Restarting minikube triggers the [file sync mechanism](https://minikube.sigs.k8s.io/Handbook/sync/) that copies the yaml file onto the minikube node and causes the API server to read the changed policy file.

Note: Currently there is no dedicated directory to store the `audit-policy.yaml` file in `~/.minikube/`. Using the `~/.minikube/files/etc/ssl/certs` directory is a workaround! This workaround works like this: By putting the file into a sub-directory of `~/.minikube/files/`, the [file sync mechanism](https://minikube.sigs.k8s.io/Handbook/sync/) gets triggered and copies the `audit-policy.yaml` file from the host onto the minikube node. When the API server container gets started by `kubeadm` I'll mount the `/etc/ssl/certs` directory from the minikube node into the container. This is the reason why the `audit-policy.yaml` file has to be stored in the ssl certs directory: It's one of the directories that get mounted from the minikube node into the container.
