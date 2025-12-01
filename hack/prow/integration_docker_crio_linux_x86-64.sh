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

OS="linux"
ARCH="amd64"
DRIVER="docker"
CONTAINER_RUNTIME="crio"
EXTRA_START_ARGS="" 
EXTRA_TEST_ARGS=""
JOB_NAME="Docker_Crio_Linux_x86-64"

git config --global --add safe.directory '*'
COMMIT=$(git rev-parse HEAD)
MINIKUBE_LOCATION=$COMMIT


# when docker is the driver, we run integration tests directly in prow cluster
# by default, prow jobs run in root, so we must switch to a non-root user to run docker driver


source ./hack/prow/common.sh 
