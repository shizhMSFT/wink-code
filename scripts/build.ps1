#!/usr/bin/env pwsh
# Cross-platform build script for wink CLI (PowerShell version)
# Builds binaries for Linux, macOS, and Windows

$ErrorActionPreference = "Stop"

# Version info
$VERSION = if ($env:VERSION) { $env:VERSION } else { "dev" }
$BUILD_TIME = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$LDFLAGS = "-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'"

# Output directory
$OUT_DIR = "dist"
if (-not (Test-Path $OUT_DIR)) {
    New-Item -ItemType Directory -Path $OUT_DIR | Out-Null
}

Write-Host "Building wink CLI v$VERSION..." -ForegroundColor Cyan
Write-Host "Build time: $BUILD_TIME" -ForegroundColor Cyan
Write-Host ""

function Build-Platform {
    param(
        [string]$OS,
        [string]$Arch,
        [string]$OutputName
    )
    
    Write-Host "Building for $OS ($Arch)..." -ForegroundColor Yellow
    $env:GOOS = $OS
    $env:GOARCH = $Arch
    
    $output = Join-Path $OUT_DIR $OutputName
    go build -ldflags $LDFLAGS -o $output ./cmd/wink
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ $output" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed to build $output" -ForegroundColor Red
        exit 1
    }
}

# Build for all platforms
Build-Platform -OS "linux" -Arch "amd64" -OutputName "wink-linux-amd64"
Build-Platform -OS "linux" -Arch "arm64" -OutputName "wink-linux-arm64"
Build-Platform -OS "darwin" -Arch "amd64" -OutputName "wink-darwin-amd64"
Build-Platform -OS "darwin" -Arch "arm64" -OutputName "wink-darwin-arm64"
Build-Platform -OS "windows" -Arch "amd64" -OutputName "wink-windows-amd64.exe"

Write-Host ""
Write-Host "Build complete! Binaries in $OUT_DIR/" -ForegroundColor Green
Get-ChildItem $OUT_DIR | Format-Table Name, Length, LastWriteTime
