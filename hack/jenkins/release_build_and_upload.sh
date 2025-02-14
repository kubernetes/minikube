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
# GITHUB_TOKEN: The GitHub API access token. Injected by the Jenkins credential provider.

set -eux -o pipefail
readonly VERSION="${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}"
readonly DEB_VERSION="${VERSION/-/\~}"
readonly DEB_REVISION="0"
readonly RPM_VERSION="${DEB_VERSION}"
readonly RPM_REVISION="0"
readonly TAGNAME="v${VERSION}"

# Make sure the tag matches the Makefile
grep -E "^VERSION_MAJOR \\?=" Makefile | grep "${VERSION_MAJOR}"
grep -E "^VERSION_MINOR \\?=" Makefile | grep "${VERSION_MINOR}"
grep -E "^VERSION_BUILD \\?=" Makefile | grep "${VERSION_BUILD}"

# Force go packages to the Jekins home directory
# export GOPATH=$HOME/go
./hack/jenkins/installers/check_install_golang.sh "/usr/local"

# Make sure docker is installed and configured
./hack/jenkins/installers/check_install_docker.sh
# Verify ISO exists
echo "Verifying ISO exists ..."
make verify-iso

# Generate licenses
make generate-licenses

# Build and upload
env BUILD_IN_DOCKER=y \
  make -j 16 \
  all \
  out/minikube-linux-arm64 \
  out/minikube-linux-arm64.tar.gz \
  out/minikube-darwin-arm64 \
  out/minikube-darwin-arm64.tar.gz \
  out/minikube-installer.exe \
  "out/minikube_${DEB_VERSION}-${DEB_REVISION}_amd64.deb" \
  "out/minikube_${DEB_VERSION}-${DEB_REVISION}_arm64.deb" \
  "out/minikube_${DEB_VERSION}-${DEB_REVISION}_armhf.deb" \
  "out/minikube_${DEB_VERSION}-${DEB_REVISION}_ppc64el.deb" \
  "out/minikube_${DEB_VERSION}-${DEB_REVISION}_s390x.deb" \
  "out/docker-machine-driver-kvm2_${DEB_VERSION}-${DEB_REVISION}_amd64.deb"
  # "out/docker-machine-driver-kvm2_${DEB_VERSION}-${DEB_REVISION}_arm64.deb"

env BUILD_IN_DOCKER=y \
  make \
  "out/minikube-${RPM_VERSION}-${RPM_REVISION}.x86_64.rpm" \
  "out/minikube-${RPM_VERSION}-${RPM_REVISION}.aarch64.rpm" \
  "out/minikube-${RPM_VERSION}-${RPM_REVISION}.armv7hl.rpm" \
  "out/minikube-${RPM_VERSION}-${RPM_REVISION}.ppc64le.rpm" \
  "out/minikube-${RPM_VERSION}-${RPM_REVISION}.s390x.rpm" \
  "out/docker-machine-driver-kvm2-${RPM_VERSION}-${RPM_REVISION}.x86_64.rpm"

# check if 'commit: <commit-id>' line contains '-dirty' commit suffix
BUILT_VERSION=$("out/minikube-$(go env GOOS)-$(go env GOARCH)" version)
echo ${BUILT_VERSION}

COMMIT=$(echo ${BUILT_VERSION} | grep 'commit:' | awk '{print $2}')
if (echo ${COMMIT} | grep -q dirty); then
  echo "'minikube version' reports dirty commit: ${COMMIT}"
  exit 1
fi

# Don't upload temporary copies, avoid unused duplicate files in the release storage
rm -f out/minikube-linux-x86_64
rm -f out/minikube-linux-i686
rm -f out/minikube-linux-aarch64
rm -f out/minikube-linux-armhf
rm -f out/minikube-linux-armv7hl
rm -f out/minikube-linux-ppc64el
rm -f out/minikube-windows-amd64

make checksum

# unversioned names to avoid updating upstream Kubernetes documentation each release
cp "out/minikube_${DEB_VERSION}-0_amd64.deb" out/minikube_latest_amd64.deb
cp "out/minikube_${DEB_VERSION}-0_arm64.deb" out/minikube_latest_arm64.deb
cp "out/minikube_${DEB_VERSION}-0_armhf.deb" out/minikube_latest_armhf.deb
cp "out/minikube_${DEB_VERSION}-0_ppc64el.deb" out/minikube_latest_ppc64el.deb
cp "out/minikube_${DEB_VERSION}-0_s390x.deb" out/minikube_latest_s390x.deb

cp "out/minikube-${RPM_VERSION}-0.x86_64.rpm" out/minikube-latest.x86_64.rpm
cp "out/minikube-${RPM_VERSION}-0.aarch64.rpm" out/minikube-latest.aarch64.rpm
cp "out/minikube-${RPM_VERSION}-0.armv7hl.rpm" out/minikube-latest.armv7hl.rpm
cp "out/minikube-${RPM_VERSION}-0.ppc64le.rpm" out/minikube-latest.ppc64le.rpm
cp "out/minikube-${RPM_VERSION}-0.s390x.rpm" out/minikube-latest.s390x.rpm



echo "Generating tarballs for kicbase images"
# first get the correct tag of the kic base image
KIC_VERSION=$(grep -E "Version =" pkg/drivers/kic/types.go | cut -d \" -f 2 | cut -d "-" -f 1)
# then generate tarballs for all achitectures
for ARCH in "amd64" "arm64" "arm/v7" "ppc64le" "s390x" 
do
  SUFFIX=$(echo $ARCH | sed 's/\///g')
  IMAGE_NAME=kicbase/stable:${KIC_VERSION}
  TARBALL_NAME=out/kicbase-${KIC_VERSION}-${SUFFIX}.tar
  docker pull ${IMAGE_NAME} --platform linux/${ARCH}
  docker image save ${IMAGE_NAME} -o ${TARBALL_NAME}
  openssl sha256 "${TARBALL_NAME}" | awk '{print $2}' > "${TARBALL_NAME}.sha256"
  docker rmi -f ${IMAGE_NAME}
done


# upload to google bucket
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
