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

mkdir -p out
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/minikube-windows-amd64.exe out/
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/e2e-windows-amd64.exe out/
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/testdata .

./out/minikube-windows-amd64.exe delete --all --purge
out/e2e-windows-amd64.exe -minikube-start-args="--vm-driver=virtualbox" -expected-default-driver=hyperv -binary=out/minikube-windows-amd64.exe -test.v -test.timeout=60m > ./out/test.out 2>&1
$env:result=$lastexitcode
# If the last exit code was 0->success, x>0->error
If($env:result -eq 0){$env:status="success"}
Else {$env:status="failure"}

# generate json output using go tool test2json
cmd /c 'go tool test2json -t < ./out/test.out > ./out/test.json'
$env:GO111MODULE='on'
go get -u "github.com/medyagh/gopogh@v0.0.16"
# Generate html report 
gopogh -in ./out/test.json  -out ./out/test.html -name $env:JOB_NAME -pr $env:MINIKUBE_LOCATION -repo github.com/kubernetes/minikube/  -details $env:COMMIT
gsutil -qm cp ./out/test.json "gs://minikube-builds/logs/$env:MINIKUBE_LOCATION/$env:JOB_NAME.json"
gsutil -qm cp ./out/test.html "gs://minikube-builds/logs/$env:MINIKUBE_LOCATION/$env:JOB_NAME.html"


$env:target_url="https://storage.googleapis.com/minikube-builds/logs/$env:MINIKUBE_LOCATION/$env:JOB_NAME.html"
$json = "{`"state`": `"$env:status`", `"description`": `"Jenkins`", `"target_url`": `"$env:target_url`", `"context`": `"VirtualBox_Windows`"}"
Invoke-WebRequest -Uri "https://api.github.com/repos/kubernetes/minikube/statuses/$env:COMMIT`?access_token=$env:access_token" -Body $json -ContentType "application/json" -Method Post -usebasicparsing

Exit $env:result
