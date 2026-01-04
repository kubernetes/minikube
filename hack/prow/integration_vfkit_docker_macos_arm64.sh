#!/bin/bash

# Copyright 2025 The Kubernetes Authors All rights reserved.
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

set -e
set -x

OS="darwin"
ARCH="arm64"
DRIVER="vfkit"
CONTAINER_RUNTIME="docker"
# in prow, if you want libvirtd to be run, you have to start a privileged container as root
EXTRA_START_ARGS="--network=vmnet-shared" 
EXTRA_TEST_ARGS="-test.run TestFunctional -binary=out/minikube"
JOB_NAME="KVM_Containerd_Linux_x86"
#  marking all directories ('*') as trusted, since .git belongs to root, not minikube user
git config --global --add safe.directory '*'

# aws macos instance doesn't have kubectl pre-installed
./hack/prow/installer/check_install_kubectl.sh ${OS} ${ARCH}

# vmnet-helper breaks when the logfile name is too long(>125 chars), so we use short commit hash here
COMMIT=$(git rev-parse HEAD | cut -c1-8)
MINIKUBE_LOCATION=$COMMIT
echo "running test in $(pwd)"

set -e
source ./hack/prow/common.sh 
