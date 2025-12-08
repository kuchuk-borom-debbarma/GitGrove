# GitGrove Installation Helper for Windows
# Adds 'gg' to your User PATH

Write-Host "Welcome to GitGrove Installer!" -ForegroundColor Cyan
Write-Host "This script will help you add the 'gg' binary to your PATH."
Write-Host ""

# 1. Ask for location
$binPath = Read-Host "Please enter the absolute path to the directory containing your 'gg.exe' binary"

# Validate path
if (-not (Test-Path "$binPath")) {
    Write-Error "Error: Directory '$binPath' does not exist."
    exit 1
}

# Check for executable (warning only)
if (-not (Test-Path "$binPath\gg.exe")) {
    Write-Warning "Warning: 'gg.exe' binary not found in '$binPath'. Proceeding anyway, but ensure it exists later."
}

# 2. Add to User PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

if ($currentPath -like "*$binPath*") {
    Write-Host "GitGrove is already in your PATH!" -ForegroundColor Green
} else {
    $newPath = "$currentPath;$binPath"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "Successfully added '$binPath' to User PATH." -ForegroundColor Green
    Write-Host "You may need to restart your terminal (PowerShell/CMD) for changes to take effect." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Cyan
Read-Host "Press Enter to exit..."
