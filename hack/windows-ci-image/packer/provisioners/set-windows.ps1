Write-Host '>>> Applying Windows optimization...'

# ==========================================
# 1. Basics
# ==========================================
Set-TimeZone -Id "Pacific Standard Time"; Stop-Service -Name tzautoupdate -Force; Set-Service -Name tzautoupdate -StartupType Disabled

# ==========================================
# 2. Service Management
# ==========================================
# Early service disabling block removed (Consolidated into main list below)

# Disable Server Manager at Logon (Server specific)
Get-ScheduledTask -TaskName ServerManager -ErrorAction SilentlyContinue | Disable-ScheduledTask -ErrorAction SilentlyContinue

# Helper function for Reg ADD
function Add-Reg {
    param($path, $key, $type, $val)
    reg add $path /v $key /t $type /d $val /f | Out-Null
}

# Disable Windows Error Reporting UI
Add-Reg "HKLM\SOFTWARE\Microsoft\Windows\Windows Error Reporting" "DontShowUI" "REG_DWORD" "1"
Add-Reg "HKLM\SOFTWARE\Microsoft\Windows\Windows Error Reporting" "Disabled" "REG_DWORD" "1"

# ==========================================
# 3. Registry & UI Tweaks
# ==========================================
Write-Host "Applying Registry Tweaks..."

## 3.1 Visual Effects & Performance
$explorerVisuals = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Explorer\VisualEffects"
if (!(Test-Path $explorerVisuals)) { New-Item -Path $explorerVisuals -Force | Out-Null }
New-ItemProperty -Path $explorerVisuals -Name "VisualFXSetting" -Value 2 -PropertyType DWORD -Force | Out-Null

$visualEffectsList = "AnimateMinMax","ComboBoxAnimation","CursorShadow","DropShadow",
    "FadeToGrey","ListviewAlphaSelect","ListviewShadow","MenuAnimation",
    "SelectionFade","TaskbarAnimations","TooltipAnimation","UIEffects"
foreach ($e in $visualEffectsList) {
    Set-ItemProperty "HKCU:\Control Panel\Desktop" -Name $e -Value 0 -ErrorAction SilentlyContinue
}

## 3.2 Desktop & Colors
Set-ItemProperty -Path "HKCU:\Control Panel\Desktop" -Name "Wallpaper" -Value ""
Set-ItemProperty -Path "HKCU:\Control Panel\Colors" -Name "Background" -Value "0 0 0"

# Force Solid Color Background
$wallpapersReg = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Explorer\Wallpapers"
if (!(Test-Path $wallpapersReg)) { New-Item -Path $wallpapersReg -Force | Out-Null }
New-ItemProperty -Path $wallpapersReg -Name "BackgroundType" -Value 1 -PropertyType DWORD -Force | Out-Null

# Disable Window Dragging Contents
Set-ItemProperty -Path "HKCU:\Control Panel\Desktop" -Name "DragFullWindows" -Value "0" -ErrorAction SilentlyContinue

# Disable Lock Screen
Add-Reg "HKLM\SOFTWARE\Policies\Microsoft\Windows\Personalization" "NoLockScreen" "REG_DWORD" "1"

## 3.3 Transparency
$themesReg = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize"
if (!(Test-Path $themesReg)) { New-Item -Path $themesReg -Force | Out-Null }
New-ItemProperty -Path $themesReg -Name "EnableTransparency" -Value 0 -PropertyType DWORD -Force | Out-Null

## 3.4 Search / Cortana / Edge / Ads
# Disable Edge Startup Boost
$edgeReg = "HKLM:\SOFTWARE\Policies\Microsoft\Edge"
if (!(Test-Path $edgeReg)) { New-Item -Path $edgeReg -Force | Out-Null }
New-ItemProperty -Path $edgeReg -Name "StartupBoostEnabled" -Value 0 -PropertyType DWORD -Force | Out-Null
New-ItemProperty -Path $edgeReg -Name "SmartScreenEnabled" -Value 0 -PropertyType DWORD -Force | Out-Null

# Disable Cortana & Search Indexing
Add-Reg "HKLM\SOFTWARE\Policies\Microsoft\Windows\Windows Search" "AllowCortana" "REG_DWORD" "0"
Add-Reg "HKLM\SOFTWARE\Policies\Microsoft\Windows\Windows Search" "DisableIndexing" "REG_DWORD" "1"

# Disable News and Interests
Add-Reg "HKLM\SOFTWARE\Policies\Microsoft\Dsh" "AllowNewsAndInterests" "REG_DWORD" "0"

# Disable Lock Screen Ads & Spotlight
Add-Reg "HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\ContentDeliveryManager" "SubscribedContent-338388Enabled" "REG_DWORD" "0"
Add-Reg "HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\SpotlightState" "Roaming" "REG_DWORD" "0"

# Disable Preview Desktop & Taskbar Animations
Add-Reg "HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Advanced" "TaskbarMn" "REG_DWORD" "0"
Add-Reg "HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Advanced" "DisablePreviewDesktop" "REG_DWORD" "1"
Add-Reg "HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Advanced" "IconsOnly" "REG_DWORD" "1"

# Performance: Font Smoothing & Color Depth
Add-Reg "HKCU\Control Panel\Desktop" "LogPixels" "REG_DWORD" "96"
Add-Reg "HKCU\Control Panel\Desktop" "FontSmoothing" "REG_SZ" "0"
Add-Reg "HKCU\Control Panel\Desktop" "UserPreferencesMask" "REG_BINARY" "9012038010000000"

# ==========================================
# 3.5 Apply to Default User (For Sysprep Persistence)
# ==========================================
Write-Host "Applying settings to Default User Profile (for Sysprep persistence)..."

# Load Default User Hive
reg load "HKLM\TempDefault" "C:\Users\Default\NTUSER.DAT" | Out-Null

if (Test-Path "HKLM:\TempDefault") {
    # 1. Visual Effects
    $defExplorerVisuals = "HKLM:\TempDefault\Software\Microsoft\Windows\CurrentVersion\Explorer\VisualEffects"
    if (!(Test-Path $defExplorerVisuals)) { New-Item -Path $defExplorerVisuals -Force | Out-Null }
    New-ItemProperty -Path $defExplorerVisuals -Name "VisualFXSetting" -Value 2 -PropertyType DWORD -Force | Out-Null

    $visualEffectsList = "AnimateMinMax","ComboBoxAnimation","CursorShadow","DropShadow",
        "FadeToGrey","ListviewAlphaSelect","ListviewShadow","MenuAnimation",
        "SelectionFade","TaskbarAnimations","TooltipAnimation","UIEffects"
    foreach ($e in $visualEffectsList) {
        Set-ItemProperty "HKLM:\TempDefault\Control Panel\Desktop" -Name $e -Value 0 -ErrorAction SilentlyContinue
    }

    # 2. Desktop & Colors
    Set-ItemProperty -Path "HKLM:\TempDefault\Control Panel\Desktop" -Name "Wallpaper" -Value ""
    Set-ItemProperty -Path "HKLM:\TempDefault\Control Panel\Colors" -Name "Background" -Value "0 0 0"

    $defWallpapersReg = "HKLM:\TempDefault\Software\Microsoft\Windows\CurrentVersion\Explorer\Wallpapers"
    if (!(Test-Path $defWallpapersReg)) { New-Item -Path $defWallpapersReg -Force | Out-Null }
    New-ItemProperty -Path $defWallpapersReg -Name "BackgroundType" -Value 1 -PropertyType DWORD -Force | Out-Null
    
    # 3. Disable Transparency
    $defThemesReg = "HKLM:\TempDefault\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize"
    if (!(Test-Path $defThemesReg)) { New-Item -Path $defThemesReg -Force | Out-Null }
    New-ItemProperty -Path $defThemesReg -Name "EnableTransparency" -Value 0 -PropertyType DWORD -Force | Out-Null

    # 4. Lock Screen & Ads (CurrentUser specific keys in Default hive)
    Add-Reg "HKLM\TempDefault\Software\Microsoft\Windows\CurrentVersion\ContentDeliveryManager" "SubscribedContent-338388Enabled" "REG_DWORD" "0"
    Add-Reg "HKLM\TempDefault\Software\Microsoft\Windows\CurrentVersion\SpotlightState" "Roaming" "REG_DWORD" "0"
    Add-Reg "HKLM\TempDefault\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced" "TaskbarMn" "REG_DWORD" "0"
    Add-Reg "HKLM\TempDefault\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced" "DisablePreviewDesktop" "REG_DWORD" "1"
    Add-Reg "HKLM\TempDefault\Software\Microsoft\Windows\CurrentVersion\Explorer\Advanced" "IconsOnly" "REG_DWORD" "1"

    # Unload Hive
    [gc]::Collect() # Force garbage collection to release handles
    reg unload "HKLM\TempDefault" | Out-Null
    Write-Host "Default User settings updated."
} else {
    Write-Error "Failed to load Default User hive."
}

# ==========================================
# 4. Power Settings
# ==========================================
Write-Host "Configuring Power Settings..."
powercfg /change monitor-timeout-ac 0
powercfg /change disk-timeout-ac 0
powercfg /change standby-timeout-ac 0
powercfg /change hibernate-timeout-ac 0
powercfg -X -monitor-timeout-ac 0
powercfg -X -standby-timeout-ac 0
powercfg -X -hibernate-timeout-ac 0
powercfg -H OFF

Write-Host "Restarting Explorer to apply changes..."
Get-Process explorer -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

Write-Host '>>> Applying Windows optimization completed'
