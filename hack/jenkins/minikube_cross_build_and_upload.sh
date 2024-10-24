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
# This is to done as part of the CI tests for GitHub PRs

# The script expects the following env variables:
# ghprbPullId: The pull request ID, injected from the ghpbr plugin.
# ghprbActualCommit: The commit hash, injected from the ghpbr plugin.

set -eux -o pipefail

readonly bucket="minikube-builds"

# Make sure the right golang version is installed based on Makefile
./hack/jenkins/installers/check_install_golang.sh /usr/local

sudo apt-get -y install fakeroot

declare -rx BUILD_IN_DOCKER=y
declare -rx GOPATH="$HOME/go"
declare -rx ISO_BUCKET="${bucket}/${ghprbPullId}"
declare -rx ISO_VERSION="testing"
declare -rx TAG="${ghprbActualCommit}"

declare -rx DEB_VER="$(make deb_version)"

docker kill $(docker ps -q) || true
docker rm $(docker ps -aq) || true
docker system prune -a --volumes -f
make -j 16 \
  all \
  minikube-darwin-arm64 \
  out/mkcmp \
  out/minikube_${DEB_VER}_amd64.deb \
  out/minikube_${DEB_VER}_arm64.deb \
  out/docker-machine-driver-kvm2_$(make deb_version_base).deb \
  out/docker-machine-driver-kvm2_${DEB_VER}_amd64.deb \
  out/docker-machine-driver-kvm2_${DEB_VER}_arm64.deb \
&& failed=$? || failed=$?

BUILT_VERSION=$("out/minikube-$(go env GOOS)-$(go env GOARCH)" version)
echo ${BUILT_VERSION}

COMMIT=$(echo ${BUILT_VERSION} | grep 'commit:' | awk '{print $2}')
if (echo ${COMMIT} | grep -q dirty); then
  echo "'minikube version' reports dirty commit: ${COMMIT}"
  exit 1
fi


gsutil cp "gs://${bucket}/logs/index.html" \
  "gs://${bucket}/logs/${ghprbPullId}/index.html"

if [[ "${failed}" -ne 0 ]]; then
  echo "build failed"
  exit "${failed}"
fi

cp -r test/integration/testdata out/

# Don't upload the buildroot artifacts if they exist
rm -rf out/buildroot

# At this point, the out directory contains the jenkins scripts (populated by jenkins),
# testdata, and our build output. Push the changes to GCS so that worker nodes can re-use them.

# -d: delete remote files that don't exist (removed test files, for instance)
# -J: gzip compression
# -R: recursive. strangely, this is not the default for sync.
gsutil -m rsync -dJR out "gs://${bucket}/${ghprbPullId}"

readonly bucket_mirror="minikube/latest"
readonly HEAD="master"
if [[ "${ghprbPullId}" == "${HEAD}" ]]; then
  # Copy artifacts to known mirror location
  gsutil cp -R "gs://${bucket}/${ghprbPullId}/minikube-*" "gs://${bucket_mirror}"
fi
