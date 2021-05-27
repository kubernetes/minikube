#!/bin/bash

# Copyright 2021 The Kubernetes Authors All rights reserved.
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

set -x -o pipefail

# Make sure docker is installed and configured
./hack/jenkins/installers/check_install_docker.sh || true
yes|gcloud auth configure-docker

# Make sure gh is installed and configured
./hack/jenkins/installers/check_install_gh.sh

if [[ $SP_VERSION != v* ]]; then
	SP_VERSION=v$SP_VERSION
fi

SED="sed -i"
if [ "$(uname)" = "Darwin" ]; then
       SED="sed -i ''"
fi

# Write the new version back into the Makefile
${SED} "s/STORAGE_PROVISIONER_TAG ?= .*/STORAGE_PROVISIONER_TAG ?= ${SP_VERSION}/" Makefile

# Build the new image
CIBUILD=yes make push-storage-provisioner-manifest

ec=$?
if [ $ec -gt 0 ]; then
	exit $ec
fi

# Bump the preload version
PLV=$(egrep "PreloadVersion =" pkg/minikube/download/preload.go | cut -d \" -f 2)
RAW=${PLV:1}
RAW=$((RAW+1))
PLV=v${RAW}

${SED} "s/PreloadVersion = .*/PreloadVersion = \"${PLV}\"/" pkg/minikube/download/preload.go

# Open a PR with the changes
git config user.name "minikube-bot"
git config user.email "minikube-bot@google.com"

branch=storage-provisioner-${SP_VERSION}
git checkout -b ${branch}

git add Makefile pkg/minikube/download/preload.go
git commit -m "Update storage provisioner to ${SP_VERSION}"
git remote add minikube-bot git@github.com:minikube-bot/minikube.git
git push -f minikube-bot ${branch}

gh pr create --fill --base master --head minikube-bot:${branch}
