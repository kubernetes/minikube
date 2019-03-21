#!/bin/bash

# Copyright 2018 The Kubernetes Authors All rights reserved.
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

set -eux -o pipefail

if [ "$#" -ne 1 ]; then
  echo "Usage: tag_release.sh <major>.<minor>.<build>" >&2
  exit 1
fi

readonly version=$1
readonly tag="v${version}"
if [[ ! "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "supplied version does not match expectations: ${version}"
  exit 2
fi

readonly clean_repo=$(mktemp -d)
git clone --depth 1 git@github.com:kubernetes/minikube.git "${clean_repo}"
cd "${clean_repo}"
git fetch
git checkout master
git pull
git tag -a "${tag}" -m "$version Release"
git push origin "${tag}"

