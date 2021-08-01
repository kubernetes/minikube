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

# This script should be run from the minikube repo root, and requires
# the following env variables to be set:
#   VERSION_MAJOR
#   VERSION_MINOR
#   VERSION_BUILD
#   access_token

set -eux -o pipefail

readonly REPO_DIR=$PWD
readonly NEW_VERSION=${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}
readonly SRC_DIR=$(mktemp -d)
readonly TAG="v${NEW_VERSION}"

if ! [[ "${VERSION_BUILD}" =~ ^[0-9]+$ ]]; then
  echo "NOTE: ${NEW_VERSION} appears to be a non-standard release, not updating brew"
  exit 0
fi

cd "${SRC_DIR}"
git clone https://github.com/kubernetes/minikube
cd minikube
readonly revision=$(git rev-list -n1 "${TAG}")

# Required for the brew command
export HOMEBREW_GITHUB_API_TOKEN="${access_token}"

# brew installed as the Jenkins user using:
# sh -c "$(curl -fsSL https://raw.githubusercontent.com/Linuxbrew/install/master/install.sh)"
export PATH=/home/linuxbrew/.linuxbrew/bin:$PATH

# avoid "error: you need to resolve your current index first" message
cd "${SRC_DIR}"

brew bump-formula-pr \
  --strict minikube \
  --revision="${revision}" \
  --message="This PR was automatically created by minikube release scripts. Contact @medyagh with any questions." \
  --no-browse \
  --tag="${TAG}" \
  && status=0 || status=$?

rm -Rf "${SRC_DIR}"
exit $status
