#!/bin/bash

# Copyright 2020 The Kubernetes Authors All rights reserved.
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

# This script builds the minikube binary for all 3 platforms and uploads them.
# The binaries are built on master and uploaded to a latest bucket.

set -eux -o pipefail

readonly bucket="minikube/latest"

# Make sure the right golang version is installed based on Makefile
WANT_GOLANG_VERSION=$(grep '^GO_VERSION' Makefile | awk '{ print $3 }')
./hack/jenkins/installers/check_install_golang.sh $WANT_GOLANG_VERSION /usr/local


declare -rx GOPATH=/var/lib/jenkins/go

GIT_COMMIT_AT_HEAD=$(git rev-parse HEAD | xargs)
MINIKUBE_LATEST_COMMIT=$(gsutil stat gs://minikube/latest/minikube-linux-amd64 | grep commit | awk '{ print $2 }' | xargs)

if [ "$GIT_COMMIT_AT_HEAD" = "$MINIKUBE_LATEST_COMMIT" ]; then
  echo "The current uploaded binary is already latest, skipping build"
  exit 0
fi

make cross && failed=$? || failed=$?
if [[ "${failed}" -ne 0 ]]; then
  echo "build failed"
  exit "${failed}"
fi
gsutil cp out/minikube-* "gs://${bucket}"
gsutil setmeta -r -h "x-goog-meta-commit:$GIT_COMMIT_AT_HEAD" "gs://${bucket}"
