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


# This script creates the localkube binary for a specified kubernetes Github tag (ex: v1.4.0)

# The script expects the following env variables:
# K8S_VERSION: The version of kubernetes to build localkube with
# COMMIT: The commit to build minikube with


set -e

export GOPATH=$PWD

cd $GOPATH/src/k8s.io/minikube
git checkout origin/$COMMIT
echo "======= Restoring Minikube Deps ======="
godep restore ./...

cd $GOPATH/src/k8s.io/kubernetes
git fetch --tags

echo "======= Checking out Kubernetes ${K8S_VERSION} ======="
git checkout ${K8S_VERSION}
godep restore ./...

echo "======= Saving Kubernetes ${K8S_VERSION} Dependency======="
cd $GOPATH/src/k8s.io/minikube
rm -rf Godeps/ vendor/
godep save ./...

# Test and make for all platforms
make test cross

# Build the e2e test target for Darwin and Linux. We don't run tests on Windows yet.
# We build these on Linux, but run the tests on different platforms.
# This makes it easier to provision slaves, since they don't need to have a go toolchain.'
GOPATH=$(pwd)/_gopath GOOS=darwin GOARCH=amd64 go test -c k8s.io/minikube/test/integration --tags=integration -o out/e2e-darwin-amd64
GOPATH=$(pwd)/_gopath GOOS=linux GOARCH=amd64 go test -c k8s.io/minikube/test/integration --tags=integration -o out/e2e-linux-amd64
cp -r test/integration/testdata out/

# Upload to localkube builds
gsutil cp out/localkube  gs://minikube/k8sReleases/${K8S_VERSION}+testing/localkube-linux-amd64

# Upload the SHA
openssl sha256 out/localkube | awk '{print $2}' > out/localkube.sha256
gsutil cp out/localkube.sha256  gs://minikube/k8sReleases/${K8S_VERSION}+testing/localkube-linux-amd64.sha256


# Upload the version of minikube that we used
gsutil cp -r out/* gs://minikubetest/localkubetest/${COMMIT}

make gendocs

git config user.name "minikube-bot"
git config user.email "minikube-bot@google.com"

git checkout -b "jenkins-${K8S_VERSION}"

git status
git add -A
git commit -m "Upgrade to k8s version ${K8S_VERSION}"

