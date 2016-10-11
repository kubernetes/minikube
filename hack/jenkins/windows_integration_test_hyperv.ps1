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

# $ErrorActionPreference = "Stop"
mkdir -p out
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/* out/
# chmod +x out/e2e-windows-amd64.exe
# chmod +x out/minikube-windows-amd64.exe
cp -r out/testdata ./


./out/minikube-windows-amd64.exe delete # || true

# Allow this to fail, we'll switch on the return code below.
# set +e
out/e2e-windows-amd64.exe --% -minikube-args="--vm-driver=hyperv --cpus=4 $env:EXTRA_BUILD_ARGS" -test.v -test.timeout=30m -binary=out/minikube-windows-amd64.exe
result=$?
# set -e

# If the last exit code was non-zero, return w/ error
If (-Not $?) {exit $?}

# if [[ $result -eq 0 ]]; then
#   status="success"
# else
#  status="failure"
# fi

# set +x
# target_url="https://storage.googleapis.com/minikube-builds/logs/$env:MINIKUBE_LOCATION/OSX-hyperv.txt"
# curl "https://api.github.com/repos/kubernetes/minikube/statuses/$env:COMMIT?access_token=$access_token" \
#  -H "Content-Type: application/json" \
#  -X POST \
#  -d "{\"state\": \"$status\", \"description\": \"Jenkins\", \"target_url\": \"$target_url\", \"context\": \"OSX-hyperv\"}"
# set -x

# exit $result