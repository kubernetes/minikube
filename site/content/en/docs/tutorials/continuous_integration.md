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


For any platform not yet listed or listed as "Unsure :question:" we are looking for your help!
Please file Pull Requests and / or Issues for missing CI platforms :smile:

| Platform | Known to Work? | Status |
|---|---|--|
| [Prow](https://github.com/kubernetes/test-infra/tree/master/prow) | [Yes](https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes/minikube) :heavy_check_mark: | [![Prow](https://prow.k8s.io/badge.svg?jobs=pull-minikube-build)](https://prow.k8s.io/?job=pull-minikube-build) |
| [Google Cloud Build](https://cloud.google.com/cloud-build/) | [Yes](./gcb.md) :heavy_check_mark: | [![cloud build status](https://storage.googleapis.com/minikube-ci-example/build/working.svg)](https://pantheon.corp.google.com/cloud-build/dashboard?project=k8s-minikube) |
| [Github](https://help.github.com/en/actions/automating-your-workflow-with-github-actions/about-continuous-integration) | [Yes](.github/workflows/minikube.yml) :heavy_check_mark: | [![Github](https://github.com/minikube-ci/examples/workflows/Minikube/badge.svg)](https://github.com/minikube-ci/examples/actions) |
| [Azure Pipelines](https://azure.microsoft.com/en-us/services/devops/pipelines/) | [Yes](azure-pipelines.yml) :heavy_check_mark: | [![Azure Pipelines](https://dev.azure.com/medyagh0825/minikube-ci/_apis/build/status/examples?api-version=5.1-preview.1)](https://dev.azure.com/medyagh0825/minikube-ci/_build) 
| [Travis CI](https://travis-ci.com/) | [Yes](.travis.yml) :heavy_check_mark: | [![Travis CI](https://travis-ci.com/minikube-ci/examples.svg?branch=master)](https://travis-ci.com/minikube-ci/examples/) |
| [CircleCI](https://circleci.com/) | [Yes](.circleci) :heavy_check_mark: | [![CircleCI](https://circleci.com/gh/minikube-ci/examples.svg?style=svg)](https://circleci.com/gh/minikube-ci/examples) |
| [Gitlab](https://about.gitlab.com/product/continuous-integration/) | [Yes](.gitlab-ci.yml) :heavy_check_mark: | ![Gitlab](https://gitlab.com/minikube-ci/examples/badges/master/pipeline.svg) |




## Example

 Here is an example, that runs minikube from a non-root user, and ensures that the latest stable kubectl is installed:

```shell
curl -LO \
  https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
  && install minikube-linux-amd64 /tmp/
  
kv=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
curl -LO \
  https://storage.googleapis.com/kubernetes-release/release/$kv/bin/linux/amd64/kubectl \
  && install kubectl /tmp/

export MINIKUBE_WANTUPDATENOTIFICATION=false
/tmp/minikube-linux-amd64 start --driver=docker
```
