# Setup script to add this directory to User PATH
# Intended to be run after unzipping the release.

$BinDir = Get-Location
Write-Host "GitGrove Windows Setup"
Write-Host "Adding $BinDir to User PATH..."

$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -like "*$BinDir*") {
    Write-Host "Already in User PATH."
} else {
    $NewPath = "$CurrentPath;$BinDir"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host "✅ Added to User PATH."
}

Write-Host "👉 Please restart PowerShell."
