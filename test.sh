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
set -e

REPO_PATH="k8s.io/minikube"

# Run all test files.
for TEST in $(find -name "*.go" | grep -v vendor | grep -v integration | grep _test.go | cut -d/ -f2- | sed 's|/\w*.go||g' | uniq); do
  echo $TEST
  go test -v ${REPO_PATH}/${TEST}
done

# Check gofmt
diff -u <(echo -n) <(gofmt -l -s . | grep -v "vendor\|\.gopath\|localkubecontents.go")
