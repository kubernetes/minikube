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

# A utility to write commit statuses to github, since authenticating is complicated now
function Write-GithubStatus {
	param (
		$JsonBody
	)
	# Manually build basic authorization headers, ugh
	$creds = "minikube-bot:$($env:access_token)"
	$encoded = [System.Convert]::ToBase64String([System.Text.Encoding]::ASCII.GetBytes($creds))
	$auth = "Basic $encoded"
	$headers = @{
		Authorization = $auth
	}
	Invoke-WebRequest -Uri "https://api.github.com/repos/kubernetes/minikube/statuses/$env:COMMIT" -Headers $headers -Body $JsonBody -ContentType "application/json" -Method Post -usebasicparsing
}

# cleanup possible leftovers from previous run
rm -r -Force test_reports
rm -Force testout*

$env:SHORT_COMMIT=$env:COMMIT.substring(0, 7)
$gcs_bucket="minikube-builds/logs/$env:MINIKUBE_LOCATION/$env:ROOT_JOB_ID"
$env:MINIKUBE_SUPPRESS_DOCKER_PERFORMANCE="true"

# Docker's kubectl breaks things, and comes earlier in the path than the regular kubectl. So download the expected kubectl and replace Docker's version.
$KubeVersion = (Invoke-WebRequest -Uri 'https://dl.k8s.io/release/stable.txt' -UseBasicParsing).Content
(New-Object Net.WebClient).DownloadFile("https://dl.k8s.io/release/$KubeVersion/bin/windows/amd64/kubectl.exe", "C:\Program Files\Docker\Docker\resources\bin\kubectl.exe")

# Setup the cleanup and reboot cron
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/windows_cleanup_and_reboot.ps1 C:\jenkins
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/windows_cleanup_cron.ps1 out/
./out/windows_cleanup_cron.ps1

# Grab all the scripts we'll need for integration tests
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/minikube-windows-amd64.exe out/
gsutil.cmd -m cp gs://minikube-builds/$env:MINIKUBE_LOCATION/e2e-windows-amd64.exe out/
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/testdata .
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/windows_integration_setup.ps1 out/
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/windows_integration_teardown.ps1 out/

# Make sure Docker is up and running
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/setup_docker_desktop_windows.ps1 out/
./out/setup_docker_desktop_windows.ps1
If ($lastexitcode -gt 0) {
	echo "Docker failed to start, exiting."
	$json = "{`"state`": `"failure`", `"description`": `"Jenkins: docker failed to start`", `"target_url`": `"https://storage.googleapis.com/$gcs_bucket/$env:JOB_NAME.txt`", `"context`": `"$env:JOB_NAME`"}"
	Write-GithubStatus -JsonBody $json
	./out/windows_integration_teardown.ps1
	Exit $lastexitcode
}
docker system prune -a --volumes -f

# install/update Go if required
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/installers/check_install_golang.ps1 out/
./out/check_install_golang.ps1

# Download gopogh and gotestsum
go install github.com/medyagh/gopogh/cmd/gopogh@v0.29.0
go install gotest.tools/gotestsum@v1.12.0
# temporary: remove the old install of gopogh & gotestsum as it's taking priority over our current install, preventing updating
if (Test-Path "C:\Go") {
    Remove-Item "C:\Go" -Recurse -Force
}

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

$gopogh_status=gopogh --in testout.json --out_html testout.html --out_summary testout_summary.json --name "$env:JOB_NAME" -pr $env:MINIKUBE_LOCATION --repo github.com/kubernetes/minikube/ --details "${env:COMMIT}:$(Get-Date -Format "yyyy-MM-dd"):$env:ROOT_JOB_ID"

$failures=echo $gopogh_status | jq '.NumberOfFail'
$tests=echo $gopogh_status | jq '.NumberOfTests'
$bad_status="$failures / $tests failures"

$description="$status in $elapsed minutes."
If($env:status -eq "failure") {
	$description="completed with $bad_status in $elapsed minutes."
}
echo $description


#Upload logs to gcs
If($env:EXTERNAL -eq "yes"){
	# If we're not already in GCP, we won't have credentials to upload to GCS
	# Instead, move logs to a predictable spot Jenkins can find and upload itself
	mkdir -p test_reports
	cp testout.txt test_reports/out.txt
	cp testout.json test_reports/out.json
	cp testout.html test_reports/out.html
	cp testout_summary.json test_reports/summary.json
} Else {
	gsutil -qm cp testout.txt gs://$gcs_bucket/${env:JOB_NAME}out.txt
	gsutil -qm cp testout.json gs://$gcs_bucket/${env:JOB_NAME}.json
	gsutil -qm cp testout.html gs://$gcs_bucket/${env:JOB_NAME}.html
	gsutil -qm cp testout_summary.json gs://$gcs_bucket/${env:JOB_NAME}_summary.json
}

$env:target_url="https://storage.googleapis.com/$gcs_bucket/$env:JOB_NAME.html"

# Update the PR with the new info
$json = "{`"state`": `"$env:status`", `"description`": `"Jenkins: $description`", `"target_url`": `"$env:target_url`", `"context`": `"$env:JOB_NAME`"}"
Write-GithubStatus -JsonBody $json

./out/windows_integration_teardown.ps1

Exit $env:result
