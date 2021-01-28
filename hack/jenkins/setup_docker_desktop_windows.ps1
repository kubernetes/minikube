# Copyright 2021 The Kubernetes Authors All rights reserved.
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

Get-Process "*Docker Desktop*" | Stop-Process

$attempt = 10
while($attempt -ne 0) {
  Write-Host "Attempt ", $attempt
  Write-Host "Wait for 2 minutes"
  & "C:\Program Files\Docker\Docker\Docker Desktop.exe"
  Start-Sleep 120
  $dockerInfo = docker info
  Write-Host "Docker Info ", $dockerInfo
  $serverVersion = $dockerInfo | Where-Object {$_ -Match "Server Version"}
  Write-Host "Server Version ", $serverVersion
  if (![System.String]::IsNullOrEmpty($serverVersion)) {
    Write-Host "Docker successfully started!"
    exit 0
  }
  Write-Host "Restarting Docker Desktop"
  Get-Process "*Docker Desktop*" | Stop-Process
  $attempt -= 1
}
exit 1
