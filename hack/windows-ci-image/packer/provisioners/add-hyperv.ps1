$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# https://minikube.sigs.k8s.io/docs/tutorials/docker_desktop_replacement/#Windows
$packages = ('Containers', 'Microsoft-Hyper-V')
$maxRetries = 5 # TODO(mloskot): Do exponential backoff, instead of linear?

foreach ($packageName in $packages) {
    Write-Host ">>> Installing $packageName..."

    for ($retries = 0; ; $retries++) {
        if ($retries -gt $maxRetries) {
            Write-Error ">>>> Could not install package $packageName" -ErrorAction Stop
        }

        choco install $packageName --source windowsfeatures --yes --no-prompt --no-progress --ignore-detected-reboot
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
}
