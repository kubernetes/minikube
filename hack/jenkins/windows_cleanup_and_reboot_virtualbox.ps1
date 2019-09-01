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

VBoxManage list vms | Foreach {
	$m = $_.Substring(1, $_.LastIndexOf('"')-1)
	VBoxManage controlvm $m poweroff
	VBoxManage unregistervm $m --delete
}

Remove-Item -path C:\Users\jenkins\.minikube -recurse -force
shutdown /r