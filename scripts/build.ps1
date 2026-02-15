# CS2 Better Auto Director - Build Script (GUI Version)
# This script builds the Wails-based GUI application

$ErrorActionPreference = "Stop"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "CS2 Better Auto Director - Build Script" -ForegroundColor Cyan
Write-Host "Version 3.0 - GUI Edition (Wails)" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "Checking Go installation..." -ForegroundColor Green
$goInstalled = Get-Command go -ErrorAction SilentlyContinue

if (-not $goInstalled) {
    Write-Host ""
    Write-Host "ERROR: Go is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go 1.21 or higher from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host "After installation, RESTART your terminal and run this script again." -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

$goVer = go version
Write-Host "Go found: $goVer" -ForegroundColor Green
Write-Host ""

# Check if Node.js is installed
Write-Host "Checking Node.js installation..." -ForegroundColor Green
$nodeInstalled = Get-Command node -ErrorAction SilentlyContinue

if (-not $nodeInstalled) {
    Write-Host ""
    Write-Host "ERROR: Node.js is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Wails requires Node.js for frontend build." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Install with:" -ForegroundColor White
    Write-Host "  winget install OpenJS.NodeJS.LTS" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Or download from: https://nodejs.org/" -ForegroundColor White
    Write-Host ""
    Write-Host "After installation, RESTART your terminal and run this script again." -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

$nodeVer = node --version
$npmVer = npm --version
Write-Host "Node.js found: $nodeVer" -ForegroundColor Green
Write-Host "npm found: $npmVer" -ForegroundColor Green
Write-Host ""

# Check if Wails CLI is installed
Write-Host "Checking Wails CLI installation..." -ForegroundColor Green
$wailsInstalled = Get-Command wails -ErrorAction SilentlyContinue

if (-not $wailsInstalled) {
    Write-Host ""
    Write-Host "ERROR: Wails CLI is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Installing Wails CLI..." -ForegroundColor Yellow
    try {
        go install github.com/wailsapp/wails/v2/cmd/wails@latest
        Write-Host "Wails CLI installed successfully!" -ForegroundColor Green
        Write-Host ""
        Write-Host "IMPORTANT: Please RESTART your terminal and run this script again." -ForegroundColor Cyan
        Write-Host "The Wails CLI was installed to your Go bin directory." -ForegroundColor White
        Write-Host ""
        Read-Host "Press Enter to exit"
        exit 0
    } catch {
        Write-Host "ERROR: Failed to install Wails CLI" -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red
        Read-Host "Press Enter to exit"
        exit 1
    }
}

$wailsVer = wails version
Write-Host "Wails found:" -ForegroundColor Green
Write-Host $wailsVer -ForegroundColor White
Write-Host ""

# Verify environment
Write-Host "Verifying Wails environment..." -ForegroundColor Green
wails doctor
Write-Host ""

# Navigate to Wails project directory
Write-Host "Navigating to Wails project..." -ForegroundColor Green
$rootDir = Split-Path -Parent $PSScriptRoot
$wailsDir = Join-Path $rootDir "cs2-autodirector-gui"

if (-not (Test-Path $wailsDir)) {
    Write-Host ""
    Write-Host "ERROR: Wails project directory not found!" -ForegroundColor Red
    Write-Host "Expected: $wailsDir" -ForegroundColor White
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

Set-Location $wailsDir
Write-Host "Working directory: $wailsDir" -ForegroundColor White
Write-Host ""

# Build with Wails
Write-Host "Building with Wails..." -ForegroundColor Green
Write-Host "This may take a few minutes..." -ForegroundColor Yellow
Write-Host ""

try {
    wails build
    if ($LASTEXITCODE -ne 0) { throw "wails build failed" }
    
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "SUCCESS! Build complete" -ForegroundColor Green
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host ""
    
    # Copy executable to root directory
    $builtExe = Join-Path $wailsDir "build\bin\cs2-better-autodirector.exe"
    $targetExe = Join-Path $rootDir "cs2-better-autodirector.exe"
    
    if (Test-Path $builtExe) {
        Copy-Item $builtExe $targetExe -Force
        $fileSize = (Get-Item $targetExe).Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        
        Write-Host "Created: cs2-better-autodirector.exe" -ForegroundColor Green
        Write-Host "Location: $targetExe" -ForegroundColor White
        Write-Host "File size: $fileSizeMB MB" -ForegroundColor White
        Write-Host ""
        Write-Host "You can now run the program with:" -ForegroundColor White
        Write-Host "    cs2-better-autodirector.exe          (GUI mode - default)" -ForegroundColor Cyan
        Write-Host "    cs2-better-autodirector.exe -nogui   (CLI mode)" -ForegroundColor Cyan
        Write-Host ""
    } else {
        Write-Host "WARNING: Executable not found at expected location" -ForegroundColor Yellow
        Write-Host "Expected: $builtExe" -ForegroundColor White
    }
    
} catch {
    Write-Host ""
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

Read-Host "Press Enter to exit"
