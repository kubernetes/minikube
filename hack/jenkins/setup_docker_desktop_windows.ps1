$attempt = 10
while($attempt -ne 0) {
  Write-Host "Attempt ", $attempt
  Write-Host "Wait for 3 minutes"
  & "C:\Program Files\Docker\Docker\Docker Desktop.exe"
  Start-Sleep 180
  $dockerInfo = docker info
  Write-Host "Docker Info ", $dockerInfo
  $serverVersion = $dockerInfo | Where-Object {$_ -Match "Server Version"}
  Write-Host "Server Version ", $serverVersion
  if (![System.String]::IsNullOrEmpty($serverVersion)) {
    Write-Host "Docker successfully started!"
    exit 0
  }
  Write-Host "Restarting Docker Desktop"
  Get-Process "*Docker Desktop*" | Stop-Process
  $attempt -= 1
}
