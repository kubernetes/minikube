# Copyright 2019 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

docker ps -aq | ForEach -Process {docker rm -fv $_}

mkdir -p out
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/minikube-windows-amd64.exe out/
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/e2e-windows-amd64.exe out/
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/testdata .
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/setup_docker_desktop_windows.ps1 out/

./out/setup_docker_desktop_windows.ps1

./out/minikube-windows-amd64.exe delete --all

$started=Get-Date -UFormat %s

out/e2e-windows-amd64.exe --minikube-start-args="--driver=docker" --binary=out/minikube-windows-amd64.exe --test.v --test.timeout=180m | tee testout.txt
$env:result=$lastexitcode
# If the last exit code was 0->success, x>0->error
If($env:result -eq 0){
	$env:status="success"
	echo "minikube: SUCCESS"
} Else {
	$env:status="failure"
	echo "minikube: FAIL"
}

$ended=Get-Date -UFormat %s
$elapsed=$ended-$started
$elapsed=$elapsed/60
$elapsed=[math]::Round($elapsed)

Get-Content testout.txt | go tool test2json -t > testout.json

$gopogh_status=gopogh --in testout.json --out testout.html --name "Docker_Windows" -pr $env:MINIKUBE_LOCATION --repo github.com/kubernetes/minikube/ --details $env:COMMIT

$failures=echo $gopogh_status | jq '.NumberOfFail'
$tests=echo $gopogh_status | jq '.NumberOfTests'
$bad_status="$failures / $tests failures"

$description="$status in $elapsed minute(s)."
If($env:status -eq "failure") {
	$description="completed with $bad_status in $elapsed minutes"
}
echo $description

$env:SHORT_COMMIT=$env:COMMIT.substring(0, 7)
$gcs_bucket="minikube-builds/logs/$env:MINIKUBE_LOCATION/$env:SHORT_COMMIT"
$env:target_url="https://storage.googleapis.com/$gcs_bucket/Docker_Windows.html"

#Upload logs to gcs
gsutil -qm cp testout.txt gs://$gcs_bucket/Docker_Windowsout.txt
gsutil -qm cp testout.json gs://$gcs_bucket/Docker_Windows.json
gsutil -qm cp testout.html gs://$gcs_bucket/Docker_Windows.html

# Update the PR with the new info
$json = "{`"state`": `"$env:status`", `"description`": `"Jenkins: $description`", `"target_url`": `"$env:target_url`", `"context`": `"Docker_Windows`"}"
Invoke-WebRequest -Uri "https://api.github.com/repos/kubernetes/minikube/statuses/$env:COMMIT`?access_token=$env:access_token" -Body $json -ContentType "application/json" -Method Post -usebasicparsing

# Just shutdown Docker, it's safer than anything else
Get-Process "*Docker Desktop*" | Stop-Process

# Uncomment once tunnel is fixed on Windows: https://github.com/kubernetes/minikube/issues/8304
#./out/minikube-windows-amd64.exe tunnel --cleanup
#./out/minikube-windows-amd64.exe delete --all --purge

Exit $env:result
