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

# 3. Instructions
Write-Host ""
Write-Host "🎉 Build Complete!"
Write-Host "To use 'gg', you need to add the build directory to your PATH."
Write-Host ""
Write-Host "Run this command to add it to your user PATH:"
Write-Host "  \$env:Path += \";$BuildDir\""
Write-Host "  [Environment]::SetEnvironmentVariable(\"Path\", \$env:Path + \";$BuildDir\", [EnvironmentVariableTarget]::User)"
Write-Host ""

