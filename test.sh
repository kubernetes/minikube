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

# Run "go test" on packages that have test files.
cd ${GOPATH}/src/${REPO_PATH}
TESTS=$(go list -f '{{ if .TestGoFiles }} {{.ImportPath}} {{end}}' ./...)
go test -v ${TESTS}

# Check gofmt
diff -u <(echo -n) <(gofmt -l -s . | grep -v "vendor\|\_gopath\|localkubecontents.go")
