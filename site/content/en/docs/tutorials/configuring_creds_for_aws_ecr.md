---
title: "Configure credentials for AWS Elastic Container Registry using registry-creds addon"
linkTitle: "Configure creds for AWS ECR using registry-creds"
weight: 1
date: 2020-03-25
description: >
  How to configure credentials for AWS ECR using the registry-creds addon for a minikube cluster
---

## Overview

The minikube [registry-creds addon](https://github.com/kubernetes/minikube/tree/master/deploy/addons/registry-creds) enables developers to setup credentials for pulling images from AWS ECR from inside their minikube cluster.

The addon automagically refreshes the service account token for the `default` service account in the `default` namespace.

## Prerequisites

- a working minikube cluster
- a container image in AWS ECR that you would like to use
- AWS access keys that can be used to pull the above image
- AWS account number of the account hosting the registry

## Configuring and enabling the registry-creds addon

### Configure the registry-creds addon

Configure the minikube registry-creds addon with the following command:

Note: In this tutorial, we will focus only on the AWS ECR.

```shell
minikube addons configure registry-creds
```

Follow the prompt and enter `y` for AWS ECR. Provide the requested information. It should look like this -
```shell
$ minikube addons configure registry-creds

Do you want to enable AWS Elastic Container Registry? [y/n]: y
-- Enter AWS Access Key ID: <put_access_key_here>
-- Enter AWS Secret Access Key: <put_secret_access_key_here>
-- (Optional) Enter AWS Session Token:
-- Enter AWS Region: us-west-2
-- Enter 12 digit AWS Account ID (Comma separated list): <account_number>
-- (Optional) Enter ARN of AWS role to assume:

Do you want to enable Google Container Registry? [y/n]: n

Do you want to enable Docker Registry? [y/n]: n

Do you want to enable Azure Container Registry? [y/n]: n
✅  registry-creds was successfully configured

```

### Enable the registry-creds addon

Enable the minikube registry-creds addon with the following command:

```shell
minikube addons enable registry-creds
```

### Create a deployment that uses an image in AWS ECR

This tutorial will use a vanilla alpine image that has been already uploaded into a repository in AWS ECR.

Let's use this alpine deployment that is setup to use the alpine image from ECR. Make sure you update the `image` field with a valid URI.

`alpine-deployment.yaml`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alpine-deployment
  labels:
    app: alpine
spec:
  replicas: 1
  selector:
    matchLabels:
      app: alpine
  template:
    metadata:
      labels:
        app: alpine
    spec:
      containers:
      - name: alpine
        image: <<aws_account_number>>.dkr.ecr.<<aws_region>>.amazonaws.com/alpine:latest
        command: ['sh', '-c', 'echo Container is Running ; sleep 3600']
```

Create a file called `alpine-deployment.yaml` and paste the contents above. Be sure to replace <<aws_account_number>> and <<aws_region>> with your actual account number and aws region. Then create the alpine deployment with the following command:

```shell
kubectl apply -f alpine-deployment.yaml
```

### Test your deployment

Describe the pod and verify the image pull was successful:

```shell
kubectl describe pods << alpine-deployment-pod-name >>
```

You should see an event like this:

```text
Successfully pulled image "<<account_number>>.dkr.ecr.<<aws_region>>.amazonaws.com/alpine:latest"
```

If you do not see that event, look at the troubleshooting section.

## Review

In the above tutorial, we configured the `registry-creds` addon to refresh the credentials for AWS ECR so that we could pull private container images onto our minikube cluster. We ultimately created a deployment that used an image in a private AWS ECR repository.

## Troubleshooting

- Check if you have a secret called `awsecr-cred` in the `default` namespace by running `kubectl get secrets`.
- Check if the image path is valid.
- Check if the registry-creds addon is enabled by using `minikube addons list`.

## Caveats

The service account token for the `default` service account in the `default` namespace is kept updated by the addon. If you create your deployment in a different namespace, the image pull will not work.

## Related articles

- [registry-creds addon](https://github.com/kubernetes/minikube/tree/master/deploy/addons/registry-creds)
