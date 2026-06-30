<#
.SYNOPSIS
    Deploys the self-hosted runner VM manually using your personal Azure credentials.
    Use this for dev/test instead of the GitHub Actions workflow when a service principal
    is not available.

.EXAMPLE
    .\deploy-local.ps1
    .\deploy-local.ps1 -VmName "vm-my-runner" -Location "westus2"
#>
param(
    [string]$SubscriptionId = '87d8bcb6-0a7a-4a53-a915-80ebc318693c',
    [string]$ResourceGroup  = 'rg-minikube-ci-runner',
    [string]$Location       = 'eastus',
    [string]$VmName         = 'vm-minikube-runner'
)

$ErrorActionPreference = 'Stop'
$ScriptDir = $PSScriptRoot

# ---------------------------------------------------------------------------
# 1. Login and set subscription
# ---------------------------------------------------------------------------
Write-Host "Logging in to Azure..."
az login
az account set --subscription $SubscriptionId

# ---------------------------------------------------------------------------
# 2. Prompt for VM password (not echoed to console)
# ---------------------------------------------------------------------------
$vmPassword = Read-Host "Enter VM admin password (min 12 chars, upper+lower+digit+symbol)" -AsSecureString
$env:RUNNER_AZ_VM_PASSWORD = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
    [Runtime.InteropServices.Marshal]::SecureStringToBSTR($vmPassword)
)

# ---------------------------------------------------------------------------
# 3. Create resource group
# ---------------------------------------------------------------------------
Write-Host "Creating resource group '$ResourceGroup' in '$Location'..."
az group create `
    --name $ResourceGroup `
    --location $Location `
    --tags project=minikube purpose=ci-runner

# ---------------------------------------------------------------------------
# 4. Deploy VM via Bicep
# ---------------------------------------------------------------------------
Write-Host "Deploying runner VM '$VmName'..."
az deployment group create `
    --resource-group $ResourceGroup `
    --template-file "$ScriptDir\runner.bicep" `
    --parameters "$ScriptDir\runner.bicepparam" `
    --parameters "vmName=$VmName" `
    --name "runner-local-$(Get-Date -Format 'yyyyMMdd-HHmmss')"

if ($LASTEXITCODE -ne 0) { throw "Deployment failed." }

# ---------------------------------------------------------------------------
# 5. Print next steps
# ---------------------------------------------------------------------------
$hostname = "$VmName.$Location.cloudapp.azure.com"
Write-Host ""
Write-Host "VM deployed. Next steps:"
Write-Host ""
Write-Host "  1. Wait ~2 min for OpenSSH extension to finish, then SSH in:"
Write-Host "     ssh minikubeadmin@$hostname"
Write-Host ""
Write-Host "  2. Enable Hyper-V (VM will reboot):"
Write-Host "     powershell -Command `"Install-WindowsFeature -Name Hyper-V -IncludeManagementTools; shutdown /r /t 10`""
Write-Host ""
Write-Host "  3. After reboot, SSH back in and run the runner setup:"
Write-Host "     scp hack\windows-ci-runner\setup-runner.ps1 minikubeadmin@${hostname}:C:/Users/minikubeadmin/"
Write-Host "     ssh minikubeadmin@$hostname"
Write-Host "     powershell -ExecutionPolicy Bypass -File C:\Users\minikubeadmin\setup-runner.ps1 -GitHubPAT `"<your-pat>`""
