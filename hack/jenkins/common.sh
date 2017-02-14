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


# This script downloads the test files from the build bucket and makes some executable.

# The script expects the following env variables:
# OS_ARCH: The operating system and the architecture separated by a hyphen '-' (e.g. darwin-amd64, linux-amd64, windows-amd64)
# VM_DRIVER: the vm-driver to use for the test
# EXTRA_BUILD_ARGS: additional flags to pass into minikube start
# JOB_NAME: the name of the logfile and check name to update on github


# Copy only the files we need to this workspace
mkdir -p out/ testdata/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/minikube-${OS_ARCH} out/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/e2e-${OS_ARCH} out/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/testdata/busybox.yaml testdata/
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/testdata/pvc.yaml testdata/

# Set the executable bit on the e2e binary and out binary
chmod +x out/e2e-${OS_ARCH}
chmod +x out/minikube-${OS_ARCH}

MINIKUBE_WANTREPORTERRORPROMPT=False \
	./out/minikube-${OS_ARCH} delete || true

rm -rf $HOME/.minikube || true

# See the default image
./out/minikube-${OS_ARCH} start -h | grep iso

# Force use-vendored-driver for local driver job
if [ "$JOB_NAME" = "Linux-Local" ]; then
    ./out/minikube-${OS_ARCH} config set use-vendored-driver true
fi

# Allow this to fail, we'll switch on the return code below.
set +e
out/e2e-${OS_ARCH} -minikube-args="--vm-driver=${VM_DRIVER} --cpus=4 --v=100 ${EXTRA_BUILD_ARGS}" -test.v -test.timeout=30m -binary=out/minikube-${OS_ARCH}
result=$?
set -e

if [[ $result -eq 0 ]]; then
  status="success"
else
  status="failure"
fi

set +x
target_url="https://storage.googleapis.com/minikube-builds/logs/${MINIKUBE_LOCATION}/${JOB_NAME}.txt"
curl "https://api.github.com/repos/kubernetes/minikube/statuses/${COMMIT}?access_token=$access_token" \
  -H "Content-Type: application/json" \
  -X POST \
  -d "{\"state\": \"$status\", \"description\": \"Jenkins\", \"target_url\": \"$target_url\", \"context\": \"${JOB_NAME}\"}"
set -x

exit $result