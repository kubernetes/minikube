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
OS_ARCH="darwin-amd64"
VM_DRIVER="hyperkit"
JOB_NAME="Docker_macOS"
EXTRA_START_ARGS=""
EXPECTED_DEFAULT_DRIVER="hyperkit"

cd github/minikube/hack/jenkins

#osascript -e 'quit app "Docker"';
#/Applications/Docker.app/Contents/MacOS/Docker --quit-after-install --unattended || true
#osascript -e 'quit app "Docker"'; /Applications/Docker.app/Contents/MacOS/Docker --unattended &; while [ -z "$(docker info 2> /dev/null )" ]; do printf "."; sleep 1; done; echo "" || true

docker info

id -Gn $(whoami)

brew install hyperkit

source common.sh
