---
title: "Audit Policy"
linkTitle: "Audit Policy"
weight: 1
date: 2019-11-19
description: >
  Enabling audit policy for minikube
---

## Overview

[Auditing](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/) is not enabled in minikube by default.
This tutorial shows how to provide an [Audit Policy](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-policy) file
to the minikube API server on startup.

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

Putting the file into the `~/.minikube/files/` directory triggers the [file sync mechanism](https://minikube.sigs.k8s.io/docs/tasks/sync/) to copy the `audit-policy.yaml` file
from the host onto the minikube node. When the API server container starts I'll mount the `/etc/ssl/certs` directory from the minikube node and can thus read the audit policy file. 

You most likely want to tune the [Audit Policy](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-policy) next as the one used in this tutorial is quite verbose.
To use a new file you need to stop and start minikube.
