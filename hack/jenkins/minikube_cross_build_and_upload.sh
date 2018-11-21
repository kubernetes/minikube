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

# This script builds the minikube binary for all 3 platforms and uploads them.
# This is to done as part of the CI tests for Github PRs

# The script expects the following env variables:
# ghprbPullId: The pull request ID, injected from the ghpbr plugin.
# ghprbActualCommit: The commit hash, injected from the ghpbr plugin.

set -eux -o pipefail

readonly bucket="minikube-builds"

declare -rx BUILD_IN_DOCKER=y
declare -rx GOPATH=/var/lib/jenkins/go
declare -rx ISO_BUCKET="${bucket}/${ghprbPullId}"
declare -rx ISO_VERSION="testing"
declare -rx TAG="${ghprbActualCommit}"


docker kill $(docker ps -q) || true
docker rm $(docker ps -aq) || true
make -j 16 all && failed=$? || failed=$?

gsutil cp "gs://${bucket}/logs/index.html" \
  "gs://${bucket}/logs/${ghprbPullId}/index.html"

if [[ "${failed}" -ne 0 ]]; then
  echo "build failed"
  exit "${failed}"
fi

git diff ${ghprbActualCommit} --name-only \
  $(git merge-base origin/master ${ghprbActualCommit}) \
  | grep -q deploy/iso/minikube && rebuild=1 || rebuild=0

if [[ "${rebuild}" -eq 1 ]]; then
	echo "ISO changes detected ... rebuilding ISO"
	make release-iso
fi

cp -r test/integration/testdata out/

# Don't upload the buildroot artifacts if they exist
rm -r out/buildroot || true

gsutil -m cp -r out/* "gs://${bucket}/${ghprbPullId}/"
