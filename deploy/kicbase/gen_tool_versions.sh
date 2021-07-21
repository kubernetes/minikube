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

# If changed, consider changing `cmd/minikube/cmd/version.go`
echo "Tool,Version"
docker --version \
  | sed -r 's|^\S+ \S+ ([0-9a-zA-Z.-]+), .*$|docker,\1|'
containerd --version \
  | sed -r 's|^\S+ \S+ (\S+) \S+$|containerd,\1|'
crio --version \
  | sed -rn 's|^Version:\s+(\S+)$|crio,\1|p'
podman --version \
  | sed -r 's|^\S+ \S+ (\S+)$|podman,\1|'
crictl --version \
  | sed -r 's|^\S+ \S+ (\S+)$|crictl,\1|'
buildctl --version \
  | sed -r 's|^\S+ \S+ (\S+) \S+$|buildctl,\1|'
ctr --version \
  | sed -r 's|^\S+ \S+ (\S+)$|ctr,\1|'
runc --version \
  | sed -rn 's|^runc version (\S+)$|runc,\1|p'
