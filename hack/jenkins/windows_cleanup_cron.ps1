Schtasks /create /tn cleanup_reboot /sc HOURLY /tr "Powershell C:\jenkins\windows_cleanup_and_reboot.ps1" /f
