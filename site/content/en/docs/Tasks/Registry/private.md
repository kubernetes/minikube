---
title: "Private"
linkTitle: "Private"
weight: 6
date: 2020-01-14
description: >
  How to use a private registry within minikube
---


**GCR/ECR/ACR/Docker**: minikube has an addon, `registry-creds` which maps credentials into minikube to support pulling from Google Container Registry (GCR), Amazon's EC2 Container Registry (ECR), Azure Container Registry (ACR), and Private Docker registries.  You will need to run `minikube addons configure registry-creds` and `minikube addons enable registry-creds` to get up and running.  An example of this is below:

```shell
$ minikube addons configure registry-creds
Do you want to enable AWS Elastic Container Registry? [y/n]: n

Do you want to enable Google Container Registry? [y/n]: y
-- Enter path to credentials (e.g. /home/user/.config/gcloud/application_default_credentials.json):/home/user/.config/gcloud/application_default_credentials.json

Do you want to enable Docker Registry? [y/n]: n

Do you want to enable Azure Container Registry? [y/n]: n
registry-creds was successfully configured
$ minikube addons enable registry-creds
```

For additional information on private container registries, see [this page](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

We recommend you use _ImagePullSecrets_, but if you would like to configure access on the minikube VM you can place the `.dockercfg` in the `/home/docker` directory or the `config.json` in the `/var/lib/kubelet` directory. Make sure to restart your kubelet (for kubeadm) process with `sudo systemctl restart kubelet`.

