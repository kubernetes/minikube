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

K8S_ORG_ROOT=${GOPATH}/src/k8s.io
MINIKUBE_ROOT=${K8S_ORG_ROOT}/minikube
KUBE_ROOT=${K8S_ORG_ROOT}/kubernetes
KUBE_VERSION=${KUBE_VERSION:=$(python ${MINIKUBE_ROOT}/hack/get_k8s_version.py --k8s-version-only 2>&1)}

source ${MINIKUBE_ROOT}/hack/godeps/utils.sh

godep::restore_kubernetes
${MINIKUBE_ROOT}/hack/godeps/godep-save.sh
