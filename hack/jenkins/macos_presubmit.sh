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
EXPECTED_DEFAULT_DRIVER="docker"

cd github/minikube

#docker-machine create --driver virtualbox  --virtualbox-cpu-count 2 --virtualbox-memory 8000 default
#docker-machine env default
#eval "$(docker-machine env default)"
docker info
docker version --format {{.Server.Os}}-{{.Server.Version}}

# Force python3.7
export CLOUDSDK_PYTHON=/usr/bin/python3

curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/darwin/amd64/kubectl"
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl

make integration-functional-only
