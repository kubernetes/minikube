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
docker system prune --all --force
Get-Process "*Docker Desktop*" | Stop-Process
shutdown /r
