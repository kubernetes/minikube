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

godep::ensure_godep_version() {
  GODEP_VERSION=${1:-"v79"}
  if [[ "$(godep version)" == *"godep ${GODEP_VERSION}"* ]]; then
    return
  fi
  go get -d -u github.com/tools/godep 2>/dev/null
  pushd "${GOPATH}/src/github.com/tools/godep" >/dev/null
    git checkout "${GODEP_VERSION}"
    go install .
  popd >/dev/null
  godep version
}

godep::sync_staging() {

pushd ${KUBE_ROOT} >/dev/null
  KUBE_VERSION=$(git describe)
popd >/dev/null

for repo in $(ls ${KUBE_ROOT}/staging/src/k8s.io); do
  rm -rf ${GOPATH}/src/k8s.io/${repo}
  cp -a "${KUBE_ROOT}/staging/src/k8s.io/${repo}" "${GOPATH}/src/k8s.io/"

  pushd "${GOPATH}/src/k8s.io/${repo}" >/dev/null
    git init >/dev/null
    git config --local user.email "nobody@k8s.io"
    git config --local user.name "$0"
    git add . >/dev/null
    git commit -q -m "Kubernetes ${KUBE_VERSION}" >/dev/null
    git tag ${KUBE_VERSION}
  popd >/dev/null
done
}

godep::restore_kubernetes() {
  pushd ${KUBE_ROOT} >/dev/null
    git checkout ${KUBE_VERSION}
    ./hack/godep-restore.sh
    bazel build //pkg/generated/openapi:zz_generated.openapi
  popd >/dev/null
  godep::sync_staging
}

godep::remove_staging_from_json() {
  go run ${MINIKUBE_ROOT}/hack/godeps/godeps-json-updater.go --godeps-file ${MINIKUBE_ROOT}/Godeps/Godeps.json --kubernetes-dir ${KUBE_ROOT}
}
