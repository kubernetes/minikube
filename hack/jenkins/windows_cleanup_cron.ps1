Schtasks /create /tn cleanup_reboot /sc HOURLY /tr "Powershell gsutil -m cp gs://minikube-builds/master/windows_cleanup_and_reboot.ps1 C:\jenkins; C:\jenkins\windows_cleanup_and_reboot.ps1"
