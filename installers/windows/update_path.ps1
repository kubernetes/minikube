
param(
    [cmdletbinding()]
    
    # This parameter dictates if the path needs to be added 
    [Parameter(Mandatory=$false,ParameterSetName="EnvironmentVariableAddOperation")]
    [switch]
    $Add,

    # This parameter dictates if the path needs to be removed 
    [Parameter(Mandatory=$false,ParameterSetName="EnvironmentVariableRemoveOperation")]
    [switch]
    $Remove,

    # This parameter tells us the path inside the $PATH Environment Variable for which the operation needs to be performed
    [Parameter(Mandatory=$true)]
    [string]
    $Path
)

$currentSystemPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine)

try {
    if ($Add) {
        Write-Output "[Info]:: Path needs to be added."
        Write-Output "[Info]:: Checking if the given path already exists or not"

        if ($currentSystemPath -match [Regex]::Escape($Path)) {
            Write-Output "[Info]:: The provided path already exists in the system. Exiting now."
        } else {
            Write-Output "[Info]:: The given path was not found. Adding it now."
            if ($currentSystemPath.EndsWith(";")) {
                $newSystemPath = $currentSystemPath + $Path.Trim() + ";"
            } else {
                $newSystemPath = $currentSystemPath + ";" + $Path.Trim() + ";"
            }
            [Environment]::SetEnvironmentVariable("Path", $newSystemPath, [EnvironmentVariableTarget]::Machine)
            Write-Output "[Info]:: Path has been added successfully."
        }
    } else {
        Write-Output "[Info]:: Path needs to be added."
        Write-Output "[Info]:: Checking if the given path already exists or not"

        if ($currentSystemPath -match [Regex]::Escape($Path)) {
            Write-Output "[Info]:: The provided path exists in the system. Removing now."
            $newSystemPath = $currentSystemPath.Replace(($Path.Trim() + ";"), "")
            [Environment]::SetEnvironmentVariable("Path", $newSystemPath, [EnvironmentVariableTarget]::Machine)
        } else {
            Write-Output "[Info]:: The given path was not found. Exiting now."
        }
    }
}
catch {
    Write-Output "[Error]:: There was an error while execution. Please see the details below. Ensure that the script is running with administrator privileges."
    Write-Output $_
}

Write-Output "[Exit]:: Program is now exiting"