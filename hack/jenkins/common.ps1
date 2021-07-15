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

(New-Object Net.WebClient).DownloadFile("https://github.com/medyagh/gopogh/releases/download/v0.8.0/gopogh.exe", "C:\Go\bin\gopogh.exe")
(New-Object Net.WebClient).DownloadFile("https://github.com/gotestyourself/gotestsum/releases/download/v1.6.4/gotestsum_1.6.4_windows_amd64.tar.gz", "$env:TEMP\gotestsum.tar.gz")
tar --directory "C:\Go\bin\" -xzvf "$env:TEMP\gotestsum.tar.gz" "gotestsum.exe"

gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/minikube-windows-amd64.exe out/
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/e2e-windows-amd64.exe out/
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/testdata .
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/windows_integration_setup.ps1 out/
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/windows_integration_teardown.ps1 out/

./out/minikube-windows-amd64.exe delete --all

./out/windows_integration_setup.ps1

$started=Get-Date -UFormat %s

gotestsum --jsonfile testout.json -f standard-verbose --raw-command -- `
  go tool test2json -t `
  out/e2e-windows-amd64.exe --minikube-start-args="--driver=$driver" --binary=out/minikube-windows-amd64.exe --test.v --test.timeout=$timeout |
  Tee-Object -FilePath testout.txt

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
$elapsed=[math]::Round($elapsed, 2)

$gopogh_status=gopogh --in testout.json --out_html testout.html --out_summary testout_summary.json --name "$env:JOB_NAME" -pr $env:MINIKUBE_LOCATION --repo github.com/kubernetes/minikube/ --details $env:COMMIT

$failures=echo $gopogh_status | jq '.NumberOfFail'
$tests=echo $gopogh_status | jq '.NumberOfTests'
$bad_status="$failures / $tests failures"

$description="$status in $elapsed minute(s)."
If($env:status -eq "failure") {
	$description="completed with $bad_status in $elapsed minute(s)."
}
echo $description

$env:SHORT_COMMIT=$env:COMMIT.substring(0, 7)
$gcs_bucket="minikube-builds/logs/$env:MINIKUBE_LOCATION/$env:SHORT_COMMIT"

#Upload logs to gcs
gsutil -qm cp testout.txt gs://$gcs_bucket/${env:JOB_NAME}out.txt
gsutil -qm cp testout.json gs://$gcs_bucket/${env:JOB_NAME}.json
gsutil -qm cp testout.html gs://$gcs_bucket/${env:JOB_NAME}.html
gsutil -qm cp testout_summary.json gs://$gcs_bucket/${env:JOB_NAME}_summary.json

$env:target_url="https://storage.googleapis.com/$gcs_bucket/$env:JOB_NAME.html"
# Update the PR with the new info
$json = "{`"state`": `"$env:status`", `"description`": `"Jenkins: $description`", `"target_url`": `"$env:target_url`", `"context`": `"${env:JOB_NAME}`"}"
Invoke-WebRequest -Uri "https://api.github.com/repos/kubernetes/minikube/statuses/$env:COMMIT`?access_token=$env:access_token" -Body $json -ContentType "application/json" -Method Post -usebasicparsing

./out/windows_integration_teardown.ps1

Exit $env:result
