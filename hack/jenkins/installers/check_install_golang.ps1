# Copyright 2023 The Kubernetes Authors All rights reserved.
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

function AddToPathIfMissing {
    param (
        [string]$pathToAdd,
        [string]$scope
    )
    $Path = [Environment]::GetEnvironmentVariable("Path", $scope)
    if ($Path -NotLike "*$pathToAdd*" ) {
        $Path = $Path + ";" + $pathToAdd
	[Environment]::SetEnvironmentVariable("Path", $Path, $scope)
	# refresh the terminals Path
	$env:Path = [Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [Environment]::GetEnvironmentVariable("Path", "User")
    }
}

# ensure Go is in Path
AddToPathIfMissing -pathToAdd "C:\Program Files\Go\bin" -scope "Machine"
AddToPathIfMissing -pathToAdd "$HOME\go\bin" -scope "User"

# Download Go
$GoVersion = "1.23.2"
$CurrentGo = go version
if ((!$?) -or ($CurrentGo -NotLike "*$GoVersion*")) {
    (New-Object Net.WebClient).DownloadFile("https://go.dev/dl/go$GoVersion.windows-amd64.zip", "$env:TEMP\golang.zip")
    Remove-Item "c:\Program Files\Go\*" -Recurse
    Add-Type -Assembly "System.IO.Compression.Filesystem"
    [System.IO.Compression.ZipFile]::ExtractToDirectory("$env:TEMP\golang.zip", "$env:TEMP\golang")
    Copy-Item -Path "$env:TEMP\golang\go\*" -Destination "c:\Program Files\Go\" -Recurse
    Remove-Item "$env:TEMP\golang" -Recurse
    Remove-Item "$env:TEMP\golang.zip"
}
