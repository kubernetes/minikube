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

set -o errexit
set -o nounset
set -o pipefail

MINIKUBE_ROOT=${GOPATH}/src/k8s.io/minikube
KUBE_ROOT=${GOPATH}/src/k8s.io/kubernetes

source ${MINIKUBE_ROOT}/hack/godeps/utils.sh

godep::ensure_godep_version v79
godep::sync_staging

rm -rf ${MINIKUBE_ROOT}/vendor ${MINIKUBE_ROOT}/Godeps
godep save ./...

cp -r ${KUBE_ROOT}/pkg/generated/openapi ${MINIKUBE_ROOT}/vendor/k8s.io/kubernetes/pkg/generated/

godep::remove_staging_from_json
git checkout -- ${MINIKUBE_ROOT}/vendor/golang.org/x/sys/windows

pushd ${MINIKUBE_ROOT} >/dev/null
    git apply ${MINIKUBE_ROOT}/hack/tpr-patch.diff
    git apply ${MINIKUBE_ROOT}/hack/kube-proxy-patch.diff
    git apply ${MINIKUBE_ROOT}/hack/openapi.diff
popd >/dev/null

