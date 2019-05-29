<#
	.DESCRIPTION
    This script is used to add/remove the installation path of Minikube in the PATH Environment variable as part of installation/uninstallation of Minikube.
    The script assumes that the PATH exists before running.
    
    .PARAMETER Add
    This is a Switch parameter which tells the script to ADD the path supplied to the System's PATH Environment variable.
    
    .PARAMETER Remove
    This is a Switch parameter which tells the script to REMOVE the path supplied from the System's PATH Environment variable.

    .PARAMETER Path
    This parameter accepts a string which needs to be added/removed from the System's PATH Environment Variable.
#>

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
        Write-Output "Path needs to be added."
        Write-Output "Checking if the given path already exists or not"

        if ($currentSystemPath -match [Regex]::Escape($Path)) {
            Write-Output "The provided path already exists in the system. Exiting now."
        } else {
            Write-Output "The given path was not found. Adding it now."
            if ($currentSystemPath.EndsWith(";")) {
                $newSystemPath = $currentSystemPath + $Path.Trim() + ";"
            } else {
                $newSystemPath = $currentSystemPath + ";" + $Path.Trim() + ";"
            }
            [Environment]::SetEnvironmentVariable("Path", $newSystemPath, [EnvironmentVariableTarget]::Machine)
            Write-Output "Path has been added successfully."
        }
    } else {
        Write-Output "Path needs to be added."
        Write-Output "Checking if the given path already exists or not"

        if ($currentSystemPath -match [Regex]::Escape($Path)) {
            Write-Output "The provided path exists in the system. Removing now."
            $newSystemPath = $currentSystemPath.Replace(($Path.Trim() + ";"), "")
            [Environment]::SetEnvironmentVariable("Path", $newSystemPath, [EnvironmentVariableTarget]::Machine)
        } else {
            Write-Output "The given path was not found. Exiting now."
        }
    }
}
catch {
    Write-Output "[Error]:: There was an error while execution. Please see the details below. Ensure that the script is running with administrator privileges."
    Write-Output $_
}