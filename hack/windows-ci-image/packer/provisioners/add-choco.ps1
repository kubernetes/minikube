$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# NOTE: Prefer Chocolatey from GitHub, instead of the official location which often keeps being unavailable
# iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

Write-Host '>>> Installing Chocolatey...'

Set-ExecutionPolicy Bypass -Scope Process -Force;
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;

$chocoVersion = '2.6.0'
$url = "https://github.com/chocolatey/choco/releases/download/${chocoVersion}/chocolatey.${chocoVersion}.nupkg"
$tempDir = Join-Path $env:TEMP "chocolatey_install"
$nupkgPath = Join-Path $tempDir "chocolatey.zip"

New-Item -ItemType Directory -Force -Path $tempDir | Out-Null

Write-Host ">>>> Downloading Chocolatey v$chocoVersion from $url..."
(New-Object System.Net.WebClient).DownloadFile($url, $nupkgPath)

Write-Host ">>>> Extracting..."
Expand-Archive -Path $nupkgPath -DestinationPath $tempDir -Force

Write-Host ">>>> Running installation script..."
$installScript = Join-Path $tempDir "tools\chocolateyInstall.ps1"
& $installScript

# Cleanup
try {
    Remove-Item -Recurse -Force $tempDir -ErrorAction Stop
} catch {
    Write-Warning ">>>> Failed to cleanup temp dir (likely file in use): $_"
}

# Bypass confirmation prompts
choco feature enable --name=allowGlobalConfirmation

# Avoid exit code 3010 (ERROR_SUCCESS_REBOOT_REQUIRED) which fails Packer pipeline
# https://github.com/chocolatey/choco/issues/3087#issuecomment-1552742454
# FIXME(mloskot): This leads to false success on "Failed to fetch...docker-engine not installed. The package was not found with the source(s) listed."
#choco feature disable --name=usePackageExitCodes

# Ignore any detected reboots
choco feature disable --name=exitOnRebootDetected

Write-Host '>>> Installing Chocolatey done'
