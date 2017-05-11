#!/bin/bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script expects 
# K8S_VERSION = the version of kubernetes to be upgraded
# REMOTE = the name of the git remote repository

KUBE_ORG_ROOT=$GOPATH/src/k8s.io
KUBE_ROOT=${KUBE_ORG_ROOT}/kubernetes
MINIKUBE_ROOT=${KUBE_ORG_ROOT}/minikube

UPSTREAM_REMOTE=${REMOTE:-origin}

echo "Restoring old dependencies..."
${MINIKUBE_ROOT}/hack/godeps/godep-restore.sh

# Upgrade kubernetes
pushd ${KUBE_ROOT} >/dev/null
  echo "Upgrading to version ${K8S_VERSION}..."
  git fetch ${UPSTREAM_REMOTE}
  git checkout ${K8S_VERSION}
  ./hack/godep-restore.sh
popd >/dev/null

echo "Saving new minikube dependencies..."
${MINIKUBE_ROOT}/hack/godeps/godep-save.sh

LOCALKUBE_BUCKET="minikube-builds/localkube-builds/${K8S_VERSION}" make release-localkube

