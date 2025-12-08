$ErrorActionPreference = "Stop"

Write-Host "🚀 Installing GitGrove for Windows..."

$ProjectRoot = git rev-parse --show-toplevel
$BuildDir = Join-Path $ProjectRoot "build"
$OldReleaseDir = Join-Path $ProjectRoot "build\release"
$SrcDir = Join-Path $ProjectRoot "src"

# 1. Clean old artifacts
if (Test-Path $OldReleaseDir) {
    Remove-Item -Recurse -Force $OldReleaseDir
}

# 2. Build
Write-Host "📦 Building binary..."
if (-not (Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null
}

Set-Location $SrcDir
$Env:GOOS = "windows"
$Env:GOARCH = "amd64"
go build -o "$BuildDir\gg.exe" .\cmd\gitgrove\main.go

Write-Host "✅ Build complete: $BuildDir\gg.exe"

# 3. Add to PATH
Write-Host "🔗 Configuring PATH..."
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -like "*$BuildDir*") {
    Write-Host "GitGrove is already in your User PATH."
} else {
    $NewPath = "$CurrentPath;$BuildDir"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host "✅ Added to User PATH."
}

Write-Host ""
Write-Host "🎉 Installation Complete! Version v1.0"
Write-Host "👉 Please restart PowerShell to use 'gg'."
