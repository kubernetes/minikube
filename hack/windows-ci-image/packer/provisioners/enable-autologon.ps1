$ErrorActionPreference = 'Stop'

# Our testing has shown that Windows 10 does not allow packer to run a Windows scheduled task until the admin user (packer) has logged into the system.
# So we enable AutoAdminLogon and use packer's windows-restart provisioner to get the system into a good state to allow scheduled tasks to run.
Write-Output '>>> Enabling AutoLogon...'

if ([string]::IsNullOrEmpty($Env:AUTOLOGON_USER_PASSWORD)) { throw 'env:AUTOLOGON_USER_PASSWORD must be set' }

Write-Output ">>> Enabling AutoAdminLogon to allow scheduled task created by elevated_user=$env:UserName to run..."
Set-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon' -Name AutoAdminLogon -Value 1
Set-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon' -Name DefaultUsername -Value "$env:UserName"
Set-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon' -Name DefaultPassword -Value "$env:AUTOLOGON_USER_PASSWORD"

Write-Output '>>> Enabling AutoLogon completed'
Write-Output '>>> IMPORTANT: Run Packer provisioner windows-restart as next step'
