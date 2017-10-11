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


# This script creates several Github statuses using the Github API: https://developer.github.com/v3/repos/statuses/
# This is intended to run before the tests start, so the icons show up on the Github PR and block submit until
# the tests finish.

# The script expects the following env variables:
# ghprbPullId: The pull request ID, injected from the ghpbr plugin.
# ghprbActualCommit: The commit hash, injected from the ghpbr plugin.
# access_token: The Github API access token. Injected by the Jenkins credential provider.

set -e
set +x

for job in "Linux-KVM-Kubeadm" "OSX-Virtualbox-Kubeadm" "OSX-Virtualbox" "OSX-XHyve" "OSX-Hyperkit" "Linux-Virtualbox" "Linux-KVM" "Linux-KVM-Alt" "Linux-None" "Windows-Virtualbox" "Windows-Kubeadm-CRI-O" "Linux-Container"; do
  target_url="https://storage.googleapis.com/minikube-builds/logs/${ghprbPullId}/${job}.txt"
  curl "https://api.github.com/repos/kubernetes/minikube/statuses/${ghprbActualCommit}?access_token=$access_token" \
    -H "Content-Type: application/json" \
    -X POST \
    -d "{\"state\": \"pending\", \"description\": \"Jenkins\", \"target_url\": \"${target_url}\", \"context\": \"${job}\"}"
done

