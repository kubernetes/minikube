Schtasks /create /tn cleanup_reboot /sc MINUTE /mo 30 /tr "Powershell C:\jenkins\windows_cleanup_and_reboot.ps1" /f
Disable-ScheduledTask -TaskName cleanup_reboot
