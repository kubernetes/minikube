---
title: "Setup minikube as CI step in GitHub Actions"
linkTitle: "Minikube in GitHub Actions"
weight: 1
date: 2020-06-02
description: >
  How to use minikube in GitHub Actions for testing your app
---

To install and start a minikube cluster, add the following step to your [GitHub Actions workflow](https://docs.github.com/en/actions/writing-workflows/quickstart).

```yaml
steps:
- name: start minikube
  id: minikube
  uses: medyagh/setup-minikube@latest
```

for more information see GitHub Actions marketplace [setup-minikube](https://github.com/marketplace/actions/setup-minikube).

## Example: build image & deploy to minikube on each PR

Requirements:

- a valid Dockerfile
- a valid `deployment.yaml` file with `imagePullPolicy: Never` see below for an example

Create workflow:

- copy this yaml to your workflow file for example in `.github/workflows/pr.yml`:

  ```yaml
  name: CI
  on:
    - pull_request
  jobs:
    job1:
      runs-on: ubuntu-latest
      name: build example and deploy to minikube
      steps:
      - uses: actions/checkout@v4
        with:
          repository: medyagh/local-dev-example-with-minikube
      - name: Start minikube
        uses: medyagh/setup-minikube@latest
      - name: Try the cluster!
        run: kubectl get pods -A
      - name: Build image
        run: |
          minikube image build -t local/devex:v1 .
      - name: Deploy to minikube
        run:
          kubectl apply -f deploy/k8s.yaml
          kubectl wait --for=condition=ready pod -l app=local-devex
      - name: Test service URLs
        run: |
          minikube service list
          minikube service local-devex-svc --url
          echo "------------------opening the service------------------"
          curl $(minikube service local-devex-svc --url)
  ```

The above example workflow yaml, will do the following steps on each coming PR:

1. Checks out the the source code
2. Installs & starts minikube
3. Tries out the cluster just by running `kubectl` command
4. Build the docker image using minikube's [docker-env]({{< ref "/docs/commands/docker-env.md" >}}) feature
5. Apply the deployment yaml file minikube
6. Check the service been created in minikube

### Example minikube deployment yaml with a service

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
