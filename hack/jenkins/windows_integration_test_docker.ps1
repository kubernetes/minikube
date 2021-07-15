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

# Docker's kubectl breaks things, and comes earlier in the path than the regular kubectl. So download the expected kubectl and replace Docker's version.
(New-Object Net.WebClient).DownloadFile("https://dl.k8s.io/release/v1.20.0/bin/windows/amd64/kubectl.exe", "C:\Program Files\Docker\Docker\resources\bin\kubectl.exe")

gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/setup_docker_desktop_windows.ps1 out/
gsutil.cmd -m cp -r gs://minikube-builds/$env:MINIKUBE_LOCATION/common.ps1 out/

$env:SHORT_COMMIT=$env:COMMIT.substring(0, 7)
$gcs_bucket="minikube-builds/logs/$env:MINIKUBE_LOCATION/$env:SHORT_COMMIT"

./out/setup_docker_desktop_windows.ps1
If ($lastexitcode -gt 0) {
	echo "Docker failed to start, exiting."

	$json = "{`"state`": `"failure`", `"description`": `"Jenkins: docker failed to start`", `"target_url`": `"https://storage.googleapis.com/$gcs_bucket/Docker_Windows.txt`", `"context`": `"Docker_Windows`"}"

	Invoke-WebRequest -Uri "https://api.github.com/repos/kubernetes/minikube/statuses/$env:COMMIT`?access_token=$env:access_token" -Body $json -ContentType "application/json" -Method Post -usebasicparsing

	docker system prune --all --force
	Exit $lastexitcode
}

$driver="docker"
$timeout="180m"
$env:JOB_NAME="Docker_Windows"

. ./out/common.ps1
