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

cd "$( dirname "${BASH_SOURCE[0]}" )"

REPO_PATH="k8s.io/minikube"

export GOPATH=${PWD}/.gopath
export GO15VENDOREXPERIMENT=1
export OS=${OS:-$(go env GOOS)}
export ARCH=${ARCH:-$(go env GOARCH)}

rm -f ${GOPATH}/src/${REPO_PATH}
mkdir -p $(dirname ${GOPATH}/src/${REPO_PATH})
ln -s ${PWD} $GOPATH/src/${REPO_PATH}

CGO_ENABLED=0 GOARCH=${ARCH} GOOS=${OS} go build --installsuffix cgo -a -o minikube cli/main.go
