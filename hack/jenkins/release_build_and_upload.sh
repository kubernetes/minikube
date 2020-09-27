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

# This script builds all the minikube binary for all 3 platforms as well as Windows-installer and .deb
# This is intended to be run on a new release tag in order to build/upload the required files for a release

# The script expects the following env variables:
# VERSION_MAJOR: The major version of the tag to be released.
# VERSION_MINOR: The minor version of the tag to be released.
# VERSION_BUILD: The build version of the tag to be released.
# BUCKET: The GCP bucket the build files should be uploaded to.
# GITHUB_TOKEN: The Github API access token. Injected by the Jenkins credential provider.

set -eux -o pipefail
readonly VERSION="${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}"
readonly DEB_VERSION="${VERSION/-/\~}"
readonly RPM_VERSION="${DEB_VERSION}"
readonly TAGNAME="v${VERSION}"

# Make sure the tag matches the Makefile
grep -E "^VERSION_MAJOR \\?=" Makefile | grep "${VERSION_MAJOR}"
grep -E "^VERSION_MINOR \\?=" Makefile | grep "${VERSION_MINOR}"
grep -E "^VERSION_BUILD \\?=" Makefile | grep "${VERSION_BUILD}"

# Force go packages to the Jekins home directory
export GOPATH=$HOME/go

# Verify ISO exists
echo "Verifying ISO exists ..."
make verify-iso

# Build and upload
env BUILD_IN_DOCKER=y \
  make -j 16 \
  all \
  out/minikube-installer.exe \
  "out/minikube_${DEB_VERSION}-0_amd64.deb" \
  "out/minikube-${RPM_VERSION}-0.x86_64.rpm" \
  "out/docker-machine-driver-kvm2_${DEB_VERSION}-0_amd64.deb" \
  "out/docker-machine-driver-kvm2-${RPM_VERSION}-0.x86_64.rpm"

make checksum

# unversioned names to avoid updating upstream Kubernetes documentation each release
cp "out/minikube_${DEB_VERSION}-0_amd64.deb" out/minikube_latest_amd64.deb
cp "out/minikube-${RPM_VERSION}-0.x86_64.rpm" out/minikube-latest.x86_64.rpm

gsutil -m cp out/* "gs://$BUCKET/releases/$TAGNAME/"

# Update "latest" release for non-beta/non-alpha builds
if ! [[ ${VERSION_BUILD} =~ ^[0-9]+$ ]]; then
  echo "NOTE: ${VERSION} appears to be a non-standard release, not updating /releases/latest"
  exit 0
fi

#echo "Updating Docker images ..."
#make push-gvisor-addon-image push-storage-provisioner-manifest

echo "Updating latest bucket for ${VERSION} release ..."
gsutil cp -r "gs://${BUCKET}/releases/${TAGNAME}/*" "gs://${BUCKET}/releases/latest/"
