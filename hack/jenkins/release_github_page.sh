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

# This script generates the Github Release page and uploads all the binaries/etc to that page
# This is intended to be run on a new release tag in order to generate the github release page for that release

# The script expects the following env variables:
# VERSION_MAJOR: The major version of the tag to be released.
# VERSION_MINOR: The minor version of the tag to be released.
# VERSION_BUILD: The build version of the tag to be released.
# ISO_SHA256: The sha 256 of the minikube-iso for the current release.
# GITHUB_TOKEN: The Github API access token. Injected by the Jenkins credential provider.

set -eux -o pipefail
readonly VERSION="${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}"
readonly DEB_VERSION="${VERSION/-/\~}"
readonly RPM_VERSION="${DEB_VERSION}"

readonly TAGNAME="v${VERSION}"

readonly GITHUB_ORGANIZATION="kubernetes"
readonly GITHUB_REPO="minikube"
readonly PROJECT_NAME="${GITHUB_REPO}"

# Pre-release
if ! [[ ${VERSION_BUILD} =~ ^[0-9]+$ ]]; then
  RELEASE_FLAGS="-p"
fi

RELEASE_NOTES=$(perl -e "\$p=0; while(<>) { if(/^## Version ${VERSION}/) { \$p=1 } elsif (/^##/) { \$p=0 }; if (\$p) { print }}" < CHANGELOG.md)
if [[ "${RELEASE_NOTES}" = "" ]]; then
  RELEASE_NOTES="(missing for ${VERSION})"
fi

readonly DESCRIPTION="## Release Notes

${RELEASE_NOTES}

## Installation

See [https://minikube.sigs.k8s.io/docs/start/](Getting Started)

## ISO Checksum

\`${ISO_SHA256}\`"

# ================================================================================
# Deleting release from github before creating new one
github-release delete \
  --user "${GITHUB_ORGANIZATION}" \
  --repo "${GITHUB_REPO}" \
  --tag "${TAGNAME}" \
  || true

# Creating a new release in github
github-release release ${RELEASE_FLAGS} \
    --user "${GITHUB_ORGANIZATION}" \
    --repo "${GITHUB_REPO}" \
    --tag "${TAGNAME}" \
    --name "${TAGNAME}" \
    --description "${DESCRIPTION}"

# Uploading the files into github
FILES_TO_UPLOAD=(
    'minikube-linux-amd64'
    'minikube-linux-amd64.sha256'
    'minikube-darwin-amd64'
    'minikube-darwin-amd64.sha256'
    'minikube-windows-amd64.exe'
    'minikube-windows-amd64.exe.sha256'
    'minikube-installer.exe'
    "minikube_${DEB_VERSION}.deb"
    "minikube-${RPM_VERSION}.rpm"
    'docker-machine-driver-kvm2'
    'docker-machine-driver-kvm2.sha256'
    'docker-machine-driver-hyperkit'
    'docker-machine-driver-hyperkit.sha256'
)

for UPLOAD in "${FILES_TO_UPLOAD[@]}"
do
    n=0
    until [ $n -ge 5 ]
    do
        github-release upload \
          --user "${GITHUB_ORGANIZATION}" \
          --repo "${GITHUB_REPO}" \
          --tag "${TAGNAME}" \
          --name "$UPLOAD" \
          --file "out/$UPLOAD" && break
        n=$[$n+1]
        sleep 15
    done
done
