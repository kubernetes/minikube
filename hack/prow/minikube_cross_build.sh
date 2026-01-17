#!/bin/bash

# Copyright 2025 The Kubernetes Authors All rights reserved.
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

# This script requires 3 parameter:
# $1. GO_VERSION: The version of golang to install
# $2. OS: The operating system
# $3. ARCH: The architecture
# $4. ISDOCKERDRIVER: Whether we use docker driver
set -eux -o pipefail
GO_VERSION=$1
OS=$2
ARCH=$3
ISDOCKERDRIVER=${4:-false}
readonly OS_ARCH="${OS}-${ARCH}"

echo "running build in $(pwd), current pr number: ${PULL_NUMBER:-none}"
# if we have changes in the kicbase dockerfile, we need to modify the source code
# to point to the new local image. The image will be built in the common.sh script.
if [ "$ISDOCKERDRIVER" = "true" ]; then
	git remote add origin https://github.com/kubernetes/minikube.git
	git fetch origin master
	if git diff origin/master -- deploy/kicbase/Dockerfile; then
		echo "kicbase Dockerfile unchanged, skipping image build"
	else
		echo "kicbase Dockerfile changed, change to use local debug image"
		make local-kicbase-hardcode
	fi
fi
# hack/prow/installer/check_install_golang.sh /usr/local $GO_VERSION
# declare -rx GOPATH="$HOME/go"

declare -rx BUILD_IN_DOCKER=y
make -j 16 \
  out/minikube-${OS_ARCH} \
  out/e2e-${OS_ARCH} \
  out/gvisor-addon \
&& failed=$? || failed=$?

export MINIKUBE_BIN="out/minikube-${OS_ARCH}"
export E2E_BIN="out/e2e-${OS_ARCH}"
chmod +x "${MINIKUBE_BIN}" "${E2E_BIN}"

BUILT_VERSION=$("out/minikube-$(go env GOOS)-$(go env GOARCH)" version)
echo ${BUILT_VERSION}

COMMIT=$(echo ${BUILT_VERSION} | grep 'commit:' | awk '{print $2}')
if (echo ${COMMIT} | grep -q dirty); then
  echo "'minikube version' reports dirty commit: ${COMMIT}"
  exit 1
fi

if [[ "${failed}" -ne 0 ]]; then
  echo "build failed"
  exit "${failed}"
fi
