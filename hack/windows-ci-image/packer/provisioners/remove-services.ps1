Write-Output '>>> Disabling unused services...'

# --- 1. The "Safe for Everyone" List ---
# Disables Telemetry, Maps, Fax, Insider Service, etc.
$safeServices = @(
    "DiagTrack",           # Connected User Experiences and Telemetry
    "MapsBroker",          # Downloaded Maps Manager
    "Fax",                 # Fax
    "RemoteRegistry",      # Remote Registry
    "wisvc",               # Windows Insider Service
    "TabletInputService",  # Touch Keyboard (Safe for non-touchscreens)
    "Themes",              # Visual themes (Unnecessary for headless)
    "UxSms",               # Desktop Window Manager Session Manager (Unnecessary for headless)
    "FontCache",           # Font Cache
    "PcaSvc",              # Program Compatibility Assistant
    "WerSvc",              # Windows Error Reporting Service
    "WbioSrvc"             # Windows Biometric Service
)

# --- 2. The "Gamer/Xbox" List ---
# Only run this if you DO NOT use the Xbox App or Game Pass
$xboxServices = @(
    "XboxGipSvc",          # Xbox Game Input Protocol
    "XblAuthManager",      # Xbox Live Auth Manager
    "XblGameSave",         # Xbox Live Game Save
    "XboxNetApiSvc"        # Xbox Live Networking Service
)

# --- 3. The "Search" List ---
# Stops file indexing. Search will break.
$searchServices = @(
    "WSearch",             # Windows Search Service
    "SysMain"              # Superfetch (Pre-loading apps, high disk usage)
)

# --- 4. The "Headless/Dev" List ---
# Disables Audio, Bluetooth, and Printing
$headlessServices = @(
    "Audiosrv",                 # Windows Audio
    "AudioEndpointBuilder",     # Windows Audio Endpoint Builder
    "BthAvctpSvc",              # Bluetooth Audio Gateway Service
    "BthServ",                  # Bluetooth Support Service
    "Spooler",                  # Print Spooler (DISABLES PRINTING)
    "upnphost",                 # UPnP Device Host (Network plug-and-play)
    "SSDPSRV",                  # SSDP Discovery (Network device discovery)
    "DeviceAssociationService", # Device Association Service (Pairing devices)
    "ScDeviceEnum",             # Smart Card Device Enumeration (Rarely used)
    "SCardSvr",                 # Smart Card (Rarely used)
    "DoSvc",                    # Delivery Optimization (Peer-to-peer updates)
    "lfsvc",                    # Geolocation Service
    "RmSvc",                    # Radio Management (Airplane mode etc)
    "WpnService",               # Windows Push Notifications System Service
    "CDPSvc",                   # Connected Devices Platform Service
    "TrkWks",                   # Distributed Link Tracking Client
    "InventorySvc"              # Inventory and Compatibility Appraiser (Telemetry)
)

# --- THE EXECUTION LOOP ---
# This combines the lists and disables them one by one.

$servicesToDisable = $safeServices + $xboxServices + $searchServices + $headlessServices

foreach ($service in $servicesToDisable) {
    try {
        # Check if service exists first
        if (Get-Service -Name $service -ErrorAction SilentlyContinue) {
            Write-Host "Disabling: $service" -ForegroundColor Cyan
            Stop-Service -Name $service -Force -ErrorAction Stop
            Set-Service -Name $service -StartupType Disabled -ErrorAction Stop
        } else {
            Write-Host "Service not found (already removed?): $service" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "Error modifying $service. $_" -ForegroundColor Red
    }
}

Write-Output '>>> Disabling unused services completed'

# Check running serives 
# Get-Service | Where-Object {$_.Status -eq 'Running'}

# TODO:
# install also & ([scriptblock]::Create((irm "https://debloat.raphi.re/"))) -RunDefaults -Silent
