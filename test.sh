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

# Check for python on host, and use it if possible, otherwise fall back on python dockerized
if [[ -f $(which python 2>&1) ]]; then
    PYTHON="python"
else
    PYTHON="docker run --rm -it -v $(pwd):/minikube -w /minikube python python"
fi


COV_FILE=coverage.txt
COV_TMP_FILE=coverage_tmp.txt

# Run "go test" on packages that have test files.  Also create coverage profile
echo "Running go tests..."
cd ${GOPATH}/src/${REPO_PATH}
rm -f out/$COV_FILE || true
echo "mode: count" > out/$COV_FILE
for pkg in $(go list -f '{{ if .TestGoFiles }} {{.ImportPath}} {{end}}' ./cmd/... ./pkg/...); do
    go test -tags "container_image_ostree_stub containers_image_openpgp" -v $pkg -covermode=count -coverprofile=out/$COV_TMP_FILE
    # tail -n +2 skips the first line of the file
    # for coverprofile the first line is the `mode: count` line which we only want once in our file
    tail -n +2 out/$COV_TMP_FILE >> out/$COV_FILE || (echo "Unable to append coverage for $pkg" && exit 1)
done
rm out/$COV_TMP_FILE

# Ignore these paths in the following tests.
ignore="vendor\|\_gopath\|assets.go\|out\/"

# Check gofmt
echo "Checking gofmt..."
set +e
files=$(gofmt -l -s . | grep -v ${ignore})
set -e
if [[ $files ]]; then
    gofmt -d ${files}
    echo "Gofmt errors in files: $files"
    exit 1
fi

# Check boilerplate
echo "Checking boilerplate..."
BOILERPLATEDIR=./hack/boilerplate
# Grep returns a non-zero exit code if we don't match anything, which is good in this case.
set +e
files=$(${PYTHON} ${BOILERPLATEDIR}/boilerplate.py --rootdir . --boilerplate-dir ${BOILERPLATEDIR} | grep -v $ignore)
set -e
if [[ ! -z ${files} ]]; then
    echo "Boilerplate missing in: ${files}."
    exit 1
fi

echo "Checking releases.json schema"
go run deploy/minikube/schema_check.go
