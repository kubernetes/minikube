#!/bin/bash

# Copyright 2019 The Kubernetes Authors All rights reserved.
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

# This script runs functional and addon tests on Cloud Shell

# The script expects the following env variables:
# MINIKUBE_LOCATION: GIT_COMMIT from upstream build.
# COMMIT: Actual commit ID from upstream build
# EXTRA_BUILD_ARGS (optional): Extra args to be passed into the minikube integrations tests
# access_token: The Github API access token. Injected by the Jenkins credential provider.

set -ex

gcloud cloud-shell ssh --authorize-session << EOF
 OS="linux"
 ARCH="amd64"
 DRIVER="docker"
 JOB_NAME="Docker_Cloud_Shell"
 CONTAINER_RUNTIME="docker"
 EXTRA_TEST_ARGS="-test.run (TestFunctional|TestAddons)"

 # Need to set these in cloud-shell or will not be present in common.sh
 MINIKUBE_LOCATION=$MINIKUBE_LOCATION
 COMMIT=$COMMIT
 EXTRA_BUILD_ARGS=$EXTRA_BUILD_ARGS
 access_token=$access_token

 # Prevent cloud-shell is ephemeral warnings on apt-get
 touch ~/.cloudshell/no-apt-get-warning

 gsutil -m cp -r gs://minikube-builds/${MINIKUBE_LOCATION}/installers .
 chmod +x ./installers/*.sh
 gsutil -m cp -r gs://minikube-builds/${MINIKUBE_LOCATION}/common.sh .
 chmod +x ./common.sh
 source ./common.sh
EOF
