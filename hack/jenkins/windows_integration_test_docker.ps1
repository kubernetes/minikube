# Copyright 2019 The Kubernetes Authors All rights reserved.
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

mkdir -p out
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/minikube-windows-amd64.exe out/
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/e2e-windows-amd64.exe out/
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/testdata .

./out/minikube-windows-amd64.exe delete --all

out/e2e-windows-amd64.exe -minikube-start-args="--driver=docker" -binary=out/minikube-windows-amd64.exe -test.v -test.timeout=65m
$env:result=$lastexitcode
# If the last exit code was 0->success, x>0->error
If($env:result -eq 0){$env:status="success"}
Else {$env:status="failure"}

# $env:SHORT_COMMIT=$env:COMMIT.substring(0, 7)
# to be used later to implement https://github.com/kubernetes/minikube/issues/6593
$env:target_url="https://storage.googleapis.com/minikube-builds/logs/$env:MINIKUBE_LOCATION/Docker_Windows.txt"
$json = "{`"state`": `"$env:status`", `"description`": `"Jenkins`", `"target_url`": `"$env:target_url`", `"context`": `"Docker_Windows`"}"
Invoke-WebRequest -Uri "https://api.github.com/repos/kubernetes/minikube/statuses/$env:COMMIT`?access_token=$env:access_token" -Body $json -ContentType "application/json" -Method Post -usebasicparsing

Exit $env:result
