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
set -eux -o pipefail
GO_VERSION=$1
OS=$2
ARCH=$3
readonly OS_ARCH="${OS}-${ARCH}"

echo "running build in $(pwd), current pr number: ${PULL_NUMBER:-none}"

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

if [ "${OS}" != "darwin" ]; then

	BUILT_VERSION=$("out/minikube-${OS_ARCH}" version)
	echo ${BUILT_VERSION}

	COMMIT=$(echo ${BUILT_VERSION} | grep 'commit:' | awk '{print $2}')
	if (echo ${COMMIT} | grep -q dirty); then
		echo "'minikube version' reports dirty commit: ${COMMIT}"
		exit 1
	fi
fi
if [[ "${failed}" -ne 0 ]]; then
  echo "build failed"
  exit "${failed}"
fi
