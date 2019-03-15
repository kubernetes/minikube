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

set -e

export TAGNAME=v${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}
export DEB_VERSION=${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}
export RPM_VERSION=${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}
export GOPATH=~/go

# Make sure the tag matches the Makefile
cat Makefile | grep "VERSION_MAJOR ?=" | grep $VERSION_MAJOR
cat Makefile | grep "VERSION_MINOR ?=" | grep $VERSION_MINOR
cat Makefile | grep "VERSION_BUILD ?=" | grep $VERSION_BUILD

# Build and upload
BUILD_IN_DOCKER=y make -j 16 all out/minikube-installer.exe out/minikube_${DEB_VERSION}.deb out/minikube-${RPM_VERSION}.rpm
make checksum

gsutil -m cp out/* gs://$BUCKET/releases/$TAGNAME/

# Bump latest
gsutil cp -r gs://$BUCKET/releases/$TAGNAME/* gs://$BUCKET/releases/latest/
