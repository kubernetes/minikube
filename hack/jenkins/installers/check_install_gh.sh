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

set -eux -o pipefail

GH_VERSION="2.60.1"

echo "Installing latest version of gh"
curl -qLO "https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_amd64.tar.gz"
tar -xf "gh_${GH_VERSION}_linux_amd64.tar.gz" &&
sudo mv "gh_${GH_VERSION}_linux_amd64/bin/gh" /usr/local/bin/gh
rm "gh_${GH_VERSION}_linux_amd64.tar.gz"
rm -rf "gh_${GH_VERSION}_linux_amd64"

echo "Authorizing bot with gh"
echo "${access_token}" | gh auth login --with-token
gh config set prompt disabled
