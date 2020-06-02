---
title: "Setup minikube as CI step in github actions"
linkTitle: "minikube in github actions"
weight: 1
date: 2020-06-02
description: >
  How to use minikube in github actions for testing your app?
---

To install and start a minikube cluster, add the following step to your [github action workflow](https://help.github.com/en/actions/configuring-and-managing-workflows/configuring-a-workflow).


```yaml
    steps:
    - name: start minikube
      id: minikube
      uses: medyagh/setup-minikube@master
```

for more information checkout github actions marketplace :[setup-minikube](https://github.com/marketplace/actions/setup-minikube).


## Full example: build image and deploy to minikube on each PR

Requirements:

- a valid Dockerfile
- a valid deployment.yaml  (make sure image pull policy is set to never see bellow for example

Steps:

- Copy the wokryaml to `.github/workflows/pr.yml` in your github repo.
- Make a PR to your repo and see the result in github actions.

```yaml
name: CI
on:
  - pull_request
jobs:
  job1:
    runs-on: ubuntu-latest
    name: build example and deploy to minikbue
    steps:
    - uses: actions/checkout@v2
    - name: Start minikube
      uses: medyagh/setup-minikube@master
    - name: Try the cluster !
      run: kubectl get pods -A
    - name: Build image
      run: |
        export SHELL=/bin/bash
        eval $(minikube -p minikube docker-env)
        docker build -f ./Dockerfile -t local/example .
        echo -n "verifying images:"
        docker images
    - name: Deploy to minikube
      run:
        kubectl apply -f deploy-to-minikube.yaml
    - name: Test service URLs
      run: |
        minikube service list
        minikube service example --url
        echo "------------------opening the service------------------"
        curl $(minikube service example --url)/version
```

In this example, the above workflow yaml, will do the following steps on each coming PR:

1- Checks out the the source code
2- Installs and starts minikube
3- Trying out the cluster just by running kubectl get pods -A
4- Build the docker image using minikube's docker-env feature
5- Apply the yaml deployment yaml file minikube
6- Check the service been created in minikube

### example deployment yaml for minikube

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
    name: example
spec:
    selector:
        matchLabels:
            app: example
    replicas: 2
    template:
        metadata:
            labels:
                app: example
        spec:
            containers:
                - name: example-api
                  imagePullPolicy: Never
                  image: local/example:latest
                  resources:
                      limits:
                          cpu: 50m
                          memory: 100Mi
                      requests:
                          cpu: 25m
                          memory: 10Mi
                  ports:
                      - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
    name: example
spec:
    type: NodePort
    selector:
        app: example
    ports:
        - port: 8080
          targetPort: 8080
```
