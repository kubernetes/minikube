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

set -ex

MINIKUBE_LOCATION=$KOKORO_GITHUB_PULL_REQUEST_NUMBER
COMMIT=$KOKORO_GITHUB_PULL_REQUEST_COMMIT
OS_ARCH="darwin-amd64"
VM_DRIVER="docker"
JOB_NAME="Docker_macOS"
EXTRA_START_ARGS=""
EXPECTED_DEFAULT_DRIVER="hyperkit"

cd github/minikube/hack/jenkins

docker-machine create --driver virtualbox default
docker-machine env default
eval "$(docker-machine env default)"
docker info

# Force python3.7
export CLOUDSDK_PYTHON=/usr/bin/python3

#source common.sh
mkdir -p out/ testdata/
# Add the out/ directory to the PATH, for using new drivers.
PATH="$(pwd)/out/":$PATH
export PATH
gsutil -qm cp \
  "gs://minikube-builds/${MINIKUBE_LOCATION}/minikube-${OS_ARCH}" \
  "gs://minikube-builds/${MINIKUBE_LOCATION}/docker-machine-driver"-* \
  "gs://minikube-builds/${MINIKUBE_LOCATION}/e2e-${OS_ARCH}" out

gsutil -qm cp -r "gs://minikube-builds/${MINIKUBE_LOCATION}/testdata"/* testdata/

gsutil -qm cp "gs://minikube-builds/${MINIKUBE_LOCATION}/gvisor-addon" testdata/
export MINIKUBE_BIN="out/minikube-${OS_ARCH}"
export E2E_BIN="out/e2e-${OS_ARCH}"
chmod +x "${MINIKUBE_BIN}" "${E2E_BIN}" out/docker-machine-driver-*
"${MINIKUBE_BIN}" version

"${MINIKUBE_BIN}" start --driver=docker
