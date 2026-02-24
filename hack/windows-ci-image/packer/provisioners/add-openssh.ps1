$ErrorActionPreference = 'Stop'

Write-Host '>>> Installing SSH public key...'

# https://learn.microsoft.com/en-us/windows-server/administration/openssh/openssh_keymanagement
if (-not (Test-Path -Path C:\ProgramData\ssh\admin.pub.txt)) {
    Write-Error -Message "Public key C:\ProgramData\ssh\admin.pub.txt not found" -ErrorAction Stop
}

$publicKey = Get-Content -Path C:\ProgramData\ssh\admin.pub.txt
try {
    $publicKey = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($publicKey))
} catch {
    $publicKey = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($publicKey))
}
Add-Content -Path C:\ProgramData\ssh\administrators_authorized_keys -Value "$publicKey" -Force

icacls.exe "$env:ProgramData\ssh\administrators_authorized_keys" /inheritance:r /grant "Administrators:F" /grant "SYSTEM:F"

Write-Host '>>> Installing SSH public key done'

# WARNING: This is super slow, may take even 15 minutes
# Write-Host '>>> Installing OpenSSH Server (Windows Capability)...'
# Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
# Write-Host '>>> Installing OpenSSH Server (Windows Capability) completed'

# NOTE: This is quick, but Chocolatey may timeout frequently
# Write-Host '>>> Installing OpenSSH Server (choco)...'
# Set-ExecutionPolicy Bypass -Scope Process -Force;
# choco install openssh --yes --no-prompt --no-progress --ignore-detected-reboot
# & 'C:\Program Files\OpenSSH-Win64\install-sshd.ps1'
# Write-Host '>>> Installing OpenSSH Server (choco) done'

# NOTE: This is even quicker and reliable too
Write-Host '>>> Installing OpenSSH Server (zip)...'
Set-ExecutionPolicy Bypass -Scope Process -Force;
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
Invoke-WebRequest -Uri 'https://github.com/PowerShell/Win32-OpenSSH/releases/download/10.0.0.0p2-Preview/OpenSSH-Win64.zip' -OutFile C:\openssh.zip
Expand-Archive -Path C:\openssh.zip -DestinationPath 'C:\Program Files\'
Remove-Item -Path C:\openssh.zip -Force
& 'C:\Program Files\OpenSSH-Win64\install-sshd.ps1'
Write-Host '>>> Installing OpenSSH Server (zip) done'

Write-Host '>>> Starting OpenSSH Server...'
Start-Service -Name sshd
Set-Service -Name sshd -StartupType 'Automatic'
Write-Host '>>> Starting OpenSSH Server completed'

Write-Host '>>> Setting Windows Firewall for SSH...'
if (-not (Get-NetFirewallRule -Name 'OpenSSH-Server-In-TCP' -ErrorAction SilentlyContinue)) {
    New-NetFirewallRule -Name 'OpenSSH-Server-In-TCP' -DisplayName 'OpenSSH SSH Server (sshd)' -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22
}
Set-NetFirewallRule -DisplayName 'OpenSSH SSH Server (sshd)' -Profile Any -Enabled True
Write-Host '>>> Setting Windows Firewall for SSH completed'

Write-Host '>>> Setting PowerShell 5.1 as SSH default shell...'
New-ItemProperty -Path 'HKLM:\SOFTWARE\OpenSSH' -Name DefaultShell -Value 'C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe' -PropertyType String -Force
New-ItemProperty -Path 'HKLM:\SOFTWARE\OpenSSH' -Name DefaultShellCommandOption -Value '/c' -PropertyType String -Force
Write-Host '>>> Setting PowerShell 5.1 as SSH default shell completed'

Restart-Service -Name sshd
