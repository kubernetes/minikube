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
docker system prune --all --force --volumes
Get-Process "*Docker Desktop*" | Stop-Process
Get-VM | Where-Object {$_.Name -ne "DockerDesktopVM"} | Foreach {
	Suspend-VM $_.Name
	Stop-VM $_.Name -Force
	Remove-VM $_.Name -Force
}
VBoxManage list vms | Foreach {
	$m = $_.Substring(1, $_.LastIndexOf('"')-1)
	VBoxManage controlvm $m poweroff
	VBoxManage unregistervm $m --delete
}
shutdown /r
