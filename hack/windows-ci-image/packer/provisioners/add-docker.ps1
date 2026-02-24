$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# https://minikube.sigs.k8s.io/docs/tutorials/docker_desktop_replacement/#Windows
$packageName = 'docker-engine'
$maxRetries = 5 # TODO(mloskot): Do exponential backoff, instead of linear?

Write-Host ">>> Installing $packageName..."

if (-not (Get-Service -Name vmic*)) {
    Write-Error ">>>> add-docker.ps1 must run after add-hyperv.ps1" -ErrorAction Stop
}

for ($retries = 0; ; $retries++) {
    if ($retries -gt $maxRetries) {
        Write-Error ">>>> Could not install package $packageName" -ErrorAction Stop
    }

    choco install $packageName --yes --no-prompt --no-progress --ignore-detected-reboot
    Write-Host ">>>> choco install exit code $LASTEXITCODE"

    # Ignore exit code 3010 (ERROR_SUCCESS_REBOOT_REQUIRED)
    if ($LASTEXITCODE -eq 0 -or $LASTEXITCODE -eq 3010) {
        $global:LASTEXITCODE = 0 # Packer will fail on exit code 3010
        Write-Host ">>>> Package $packageName successfully installed"
        break
    }
    else {
        Write-Host ">>>> Error installing $packageName, retrying in $maxRetries seconds..."
        Start-Sleep -Seconds $maxRetries
    }
}

Write-Host ">>> Installing $packageName done"
