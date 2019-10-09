function Jenkins {
	Get-Process e2e-windows-amd64 2>$NULL
	if ($?) {
		return $TRUE
	}
	return $FALSE
}

if (Jenkins) {
	exit 0
}
echo "waiting to see if any jobs are coming in..."
timeout 30
if (Jenkins) {
	exit 0
}
echo "doing it"
taskkill /IM putty.exe
taskkill /F /IM java.exe
Get-VM | Where-Object {$_.Name -ne "DockerDesktopVM"} | Foreach {
	C:\var\jenkins\workspace\Hyper-V_Windows_integration\out\minikube-windows-amd64.exe delete -p $_.Name
	Suspend-VM $_.Name
	Stop-VM $_.Name -Force
	Remove-VM $_.Name -Force
}
Remove-Item -path C:\Users\admin\.minikube -recurse -force
shutdown /r