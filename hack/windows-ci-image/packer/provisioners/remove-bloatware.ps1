Write-Host '>>> Uninstalling applications...'

$bloatware = @(
    "*bing*",
    "*feedback*",
    "*getstarted*",
    "*msteams*"
    "*news*",
    "*onenote*",
    "*outlook*",
    "*people*",
    "*quickassist*",
    "*skype*"
    "*solitaire*",
    "*stickynotes*",
    "*tips*",
    "*weather*",
    "*windowscommunicationsapps*", # Mail & Calendar
    "*xbox*",
    "*yourphone*",
    "*zune*",
    "Clipchamp.Clipchamp*",
    "Microsoft.BingWeather*"
    "Microsoft.GamingApp*",
    "Microsoft.GetHelp*",
    "Microsoft.MicrosoftOfficeHub*",
    "Microsoft.OneDriveSync*",
    "Microsoft.Paint*",
    "Microsoft.PowerAutomateDesktop*",
    "Microsoft.ScreenSketch*",
    "Microsoft.Teams*",
    "Microsoft.Todos*",
    "Microsoft.Windows.DevHome*",
    "Microsoft.Windows.Photos*",
    "Microsoft.WindowsCalculator*",
    "Microsoft.WindowsCamera*",
    "Microsoft.WindowsSoundRecorder*"
)

if (Get-Command Get-AppxPackage -ErrorAction SilentlyContinue) {
    foreach ($app in $bloatware) {
        Write-Host "Removing $app..."
        Get-AppxPackage -AllUsers -Name $app -ErrorAction SilentlyContinue | Where-Object { $_.Name -notlike "*XboxGameCallableUI*" -and $_.Name -notlike "*PeopleExperienceHost*" } | Remove-AppxPackage -AllUsers -ErrorAction SilentlyContinue
        Get-AppxProvisionedPackage -Online | Where-Object { $_.DisplayName -Like $app -and $_.DisplayName -notlike "*XboxGameCallableUI*" -and $_.DisplayName -notlike "*PeopleExperienceHost*" } | Remove-AppxProvisionedPackage -Online -ErrorAction SilentlyContinue
    }
} else {
    Write-Host "Appx removal tools not found. Skipping (Expected on Server Core)."
}

Write-Host "Removing OneDrive..."
Stop-Process -Name "OneDrive" -ErrorAction SilentlyContinue
$onedrive32 = "C:\Windows\SysWOW64\OneDriveSetup.exe"
$onedrive64 = "C:\Windows\System32\OneDriveSetup.exe"
if (Test-Path $onedrive32) { Start-Process $onedrive32 -ArgumentList "/uninstall" -Wait -NoNewWindow -ErrorAction SilentlyContinue }
if (Test-Path $onedrive64) { Start-Process $onedrive64 -ArgumentList "/uninstall" -Wait -NoNewWindow -ErrorAction SilentlyContinue }
Remove-Item -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run\OneDrive" -ErrorAction SilentlyContinue
Remove-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run" -Name "OneDrive" -ErrorAction SilentlyContinue

Write-Host '>>> Uninstalling applications done'