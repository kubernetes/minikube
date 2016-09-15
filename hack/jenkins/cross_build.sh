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

# The script expects the following env variabls:
# ghprbPullId: The pull request ID, injected from the ghpbr plugin.
# ghprbActualCommit: The commit hash, injected from the ghpbr plugin.

gsutil cp gs://minikube-builds/logs/index.html gs://minikube-builds/logs/${ghprbPullId}/index.html

# Build all platforms (Windows, Linux, OSX)
make cross

# Build the e2e test target for Darwin and Linux. We don't run tests on Windows yet.
# We build these on Linux, but run the tests on different platforms.
# This makes it easier to provision slaves, since they don't need to have a go toolchain.'
GOPATH=$(pwd)/_gopath GOOS=darwin GOARCH=amd64 go test -c k8s.io/minikube/test/integration --tags=integration -o out/e2e-darwin-amd64
GOPATH=$(pwd)/_gopath GOOS=linux GOARCH=amd64 go test -c k8s.io/minikube/test/integration --tags=integration -o out/e2e-linux-amd64
cp -r test/integration/testdata out/

# Upload everything we built to Cloud Storage.
gsutil -m cp -r out/* gs://minikube-builds/${ghprbPullId}/
