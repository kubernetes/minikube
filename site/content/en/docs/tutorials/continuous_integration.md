---
title: "Continuous Integration"
weight: 1
description: >
  How to run minikube in CI (Continuous Integration)
---

## Overview


Most continuous integration environments are already running inside a VM, and may not support nested virtualization.
You could use either `none` or `docker` driver in CI.

To see a working example of running minikube in CI checkout [minikube-ci/examples](https://github.com/minikube-ci/examples) that contains working examples.


## Supported / Tested CI Platforms


For any platform not yet listed we are looking for your help! Please file Pull Requests and / or Issues for missing CI platforms 😄

| Platform | Known to Work? | Status |
|---|---|--|
| [Prow](https://github.com/kubernetes/test-infra/tree/master/prow) | [Yes](https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes/minikube) ✔️ | [![Prow](https://prow.k8s.io/badge.svg?jobs=pull-minikube-build)](https://prow.k8s.io/?job=pull-minikube-build) |
| [Google Cloud Build](https://cloud.google.com/cloud-build/) | [Yes](https://github.com/minikube-ci/examples/blob/master/gcb.md) ✔️ | |
| [GitHub](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/about-continuous-integration) | [Yes](https://github.com/minikube-ci/examples/blob/master/.github/workflows/minikube.yml) ✔️ | [![GitHub](https://github.com/minikube-ci/examples/workflows/Minikube/badge.svg)](https://github.com/minikube-ci/examples/actions) |
| [Azure Pipelines](https://azure.microsoft.com/en-us/services/devops/pipelines/) | [Yes](https://github.com/minikube-ci/examples/blob/master/azure-pipelines.yml) ✔️ | [![Azure Pipelines](https://dev.azure.com/medyagh0825/minikube-ci/_apis/build/status/examples?api-version=5.1-preview.1)](https://dev.azure.com/medyagh0825/minikube-ci/_build) 
| [Travis CI](https://travis-ci.com/) | [Yes](https://github.com/minikube-ci/examples/blob/master/.travis.yml) ✔️ | [![Travis CI](https://travis-ci.com/minikube-ci/examples.svg?branch=master)](https://travis-ci.com/minikube-ci/examples/) |
| [CircleCI](https://circleci.com/) | [Yes](https://github.com/minikube-ci/examples/blob/master/.circleci) ✔️ | [![CircleCI](https://circleci.com/gh/minikube-ci/examples.svg?style=svg)](https://circleci.com/gh/minikube-ci/examples) |
| [Gitlab](https://about.gitlab.com/product/continuous-integration/) | [Yes](https://github.com/minikube-ci/examples/blob/master/.gitlab-ci.yml) ✔️ | Gitlab |




## Example

 Here is an example, that runs minikube from a non-root user, and ensures that the latest stable kubectl is installed:

```shell
curl -LO \
  https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
  && install minikube-linux-amd64 /tmp/
  
kv=$(curl -sSL https://dl.k8s.io/release/stable.txt)
curl -LO \
  https://dl.k8s.io/$kv/bin/linux/amd64/kubectl \
  && install kubectl /tmp/

/tmp/minikube-linux-amd64 config set WantUpdateNotification false
/tmp/minikube-linux-amd64 start --driver=docker
```
