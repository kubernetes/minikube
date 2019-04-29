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
#   BOT_PASSWORD

set -eux -o pipefail

readonly REPO_DIR=$PWD
readonly NEW_VERSION=${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_BUILD}
readonly NEW_SHA256=$(awk '{ print $1 }' "${REPO_DIR}/out/minikube-darwin-amd64.sha256")
readonly BUILD_DIR=$(mktemp -d)
readonly GITHUB_USER="minikube-bot"

if [ -z "${NEW_SHA256}" ]; then
  echo "SHA256 is empty :("
  exit 1
fi

git config --global user.name "${GITHUB_USER}"
git config --global user.email "${GITHUB_USER}@google.com"

cd "${BUILD_DIR}"
git clone --depth 1 "git@github.com:${GITHUB_USER}/homebrew-cask.git"
cd homebrew-cask
git remote add upstream https://github.com/Homebrew/homebrew-cask.git
git fetch upstream

git checkout upstream/master
git checkout -b "${NEW_VERSION}"
sed -e "s/\$PKG_VERSION/${NEW_VERSION}/g" \
    -e "s/\$MINIKUBE_DARWIN_SHA256/${NEW_SHA256}/g" \
    "${REPO_DIR}/installers/darwin/brew-cask/minikube.rb.tmpl" > Casks/minikube.rb
git add Casks/minikube.rb
git commit -F- <<EOF
Update minikube to ${NEW_VERSION}

- [x] brew cask audit --download {{cask_file}} is error-free.
- [x] brew cask style --fix {{cask_file}} reports no offenses.
- [x] The commit message includes the cask’s name and version.

EOF

git push origin "${NEW_VERSION}"
curl -v -k -u "${GITHUB_USER}:${BOT_PASSWORD}" -X POST https://api.github.com/repos/Homebrew/homebrew-cask/pulls \
-d @- <<EOF

{
    "title": "Update minikube to ${NEW_VERSION}",
    "head": "${GITHUB_USER}:${NEW_VERSION}",
    "base": "master",
    "body": "cc @balopat"
}
EOF

rm -Rf "${BUILD_DIR}"
