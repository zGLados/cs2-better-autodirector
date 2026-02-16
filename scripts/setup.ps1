# CS2 Better Auto Director - Setup Script (GUI Version)
# This script checks requirements and builds the Wails GUI application

$ErrorActionPreference = "Stop"

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host " CS2 Better Auto Director - Setup Script" -ForegroundColor Cyan
Write-Host " Version 3.0 - GUI Edition" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "This script will:" -ForegroundColor White
Write-Host " - Check if Go, Node.js, and Wails are installed" -ForegroundColor White
Write-Host " - Build cs2-better-autodirector.exe (with GUI)" -ForegroundColor White
Write-Host " - Copy GSI config to CS folder (if found)" -ForegroundColor White
Write-Host ""
Write-Host "Press any key to continue or Ctrl+C to cancel..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
Write-Host ""

# Check if Go is installed
Write-Host "[1/5] Checking Go installation..." -ForegroundColor Green
$goInstalled = Get-Command go -ErrorAction SilentlyContinue

if (-not $goInstalled) {
    Write-Host ""
    Write-Host "ERROR: Go is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go 1.21 or higher from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host "After installation, run this script again." -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

$goVer = go version
Write-Host "Go is already installed: $goVer" -ForegroundColor Green
Write-Host ""

# Check if Node.js is installed
Write-Host "[2/5] Checking Node.js installation..." -ForegroundColor Green
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
    Write-Host "After installation, run this script again." -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

$nodeVer = node --version
$npmVer = npm --version
Write-Host "Node.js is already installed: $nodeVer (npm $npmVer)" -ForegroundColor Green
Write-Host ""

# Check if Wails CLI is installed
Write-Host "[3/5] Checking Wails CLI installation..." -ForegroundColor Green
$wailsInstalled = Get-Command wails -ErrorAction SilentlyContinue

if (-not $wailsInstalled) {
    Write-Host ""
    Write-Host "Wails CLI not found. Installing..." -ForegroundColor Yellow
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
Write-Host "Wails is already installed:" -ForegroundColor Green
Write-Host $wailsVer -ForegroundColor White
Write-Host ""

# Build application
Write-Host "[4/5] Building cs2-better-autodirector.exe..." -ForegroundColor Green
Write-Host ""

try {
    # Navigate to Wails project directory
    $rootDir = Split-Path -Parent $PSScriptRoot
    $wailsDir = Join-Path $rootDir "cs2-autodirector-gui"
    
    if (-not (Test-Path $wailsDir)) {
        throw "Wails project directory not found: $wailsDir"
    }
    
    Set-Location $wailsDir
    
    Write-Host "Building with Wails..." -ForegroundColor Yellow
    Write-Host "This may take a few minutes..." -ForegroundColor Yellow
    Write-Host ""
    
    wails build
    if ($LASTEXITCODE -ne 0) { throw "wails build failed" }
    
    # Copy executable to root directory
    $builtExe = Join-Path $wailsDir "build\bin\cs2-better-autodirector.exe"
    $targetExe = Join-Path $rootDir "cs2-better-autodirector.exe"
    
    if (Test-Path $builtExe) {
        Copy-Item $builtExe $targetExe -Force
        Write-Host "Build successful!" -ForegroundColor Green
        Write-Host "Created: cs2-better-autodirector.exe" -ForegroundColor Green
    } else {
        throw "Executable not found at expected location: $builtExe"
    }
    
} catch {
    Write-Host ""
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    Write-Host "Possible solutions:" -ForegroundColor Yellow
    Write-Host " 1. Make sure Go, Node.js, and Wails are properly installed" -ForegroundColor White
    Write-Host " 2. Run 'wails doctor' to check environment" -ForegroundColor White
    Write-Host " 3. Check the error message above for details" -ForegroundColor White
    Read-Host "Press Enter to exit"
    exit 1
}
Write-Host ""

# Copy GSI config
Write-Host "[5/5] Setting up Counter-Strike integration..." -ForegroundColor Green
Write-Host ""

# Go back to root directory for config copy
Set-Location $rootDir

# Find CS config folder
$csPaths = @(
    "C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg",
    "D:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg",
    "E:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg",
    "F:\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg"
)

$csConfigPath = $null
foreach ($path in $csPaths) {
    if (Test-Path $path) {
        $csConfigPath = $path
        break
    }
}

if ($csConfigPath) {
    Write-Host "Found Counter-Strike config folder:" -ForegroundColor Green
    Write-Host $csConfigPath
    Write-Host ""
    Write-Host "Copying configuration files..."
    
    try {
        Copy-Item "config\gamestate_integration_autodirector.cfg" -Destination $csConfigPath -Force
        Write-Host "✓ gamestate_integration_autodirector.cfg copied" -ForegroundColor Green
        
        Copy-Item "config\spectator_bindings.cfg" -Destination $csConfigPath -Force
        Write-Host "✓ spectator_bindings.cfg copied" -ForegroundColor Green
        
        Write-Host ""
        Write-Host "IMPORTANT: Run these commands in CS2 console:" -ForegroundColor Yellow
        Write-Host "  exec spectator_bindings" -ForegroundColor Cyan
        Write-Host "  bind F9 spec_mode_toggle" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Then press F9 to toggle between weapon bindings and spectator bindings!" -ForegroundColor Green
        Write-Host ""
    } catch {
        Write-Host "WARNING: Could not copy configs automatically." -ForegroundColor Yellow
        Write-Host "Please copy these files manually to:" -ForegroundColor Yellow
        Write-Host $csConfigPath
        Write-Host "  - config\gamestate_integration_autodirector.cfg"
        Write-Host "  - config\spectator_bindings.cfg"
    }
} else {
    Write-Host "Could not find Counter-Strike installation automatically." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please copy these files manually to:" -ForegroundColor Yellow
    Write-Host "Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\"
    Write-Host "  - config\gamestate_integration_autodirector.cfg"
    Write-Host "  - config\spectator_bindings.cfg"
    Write-Host ""
    Write-Host "Then run in CS2 console:" -ForegroundColor Cyan
    Write-Host "  exec spectator_bindings"
    Write-Host "  bind F9 spec_mode_toggle"
}

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host " INSTALLATION COMPLETE!" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

if (Test-Path "cs2-better-autodirector.exe") {
    $fileSize = (Get-Item "cs2-better-autodirector.exe").Length
    $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
    Write-Host "Created: cs2-better-autodirector.exe ($fileSizeMB MB)" -ForegroundColor Green
}

Write-Host ""
Write-Host "To start the Auto Director:" -ForegroundColor White
Write-Host ""
Write-Host "GUI Mode (with dashboard):" -ForegroundColor Cyan
Write-Host "  cs2-better-autodirector.exe" -ForegroundColor White
Write-Host ""
Write-Host "CLI Mode (terminal only):" -ForegroundColor Cyan
Write-Host "  cs2-better-autodirector.exe -nogui" -ForegroundColor White
Write-Host ""
Write-Host "Steps:" -ForegroundColor Yellow
Write-Host "  1. Start Counter-Strike"
Write-Host "  2. Go into spectator mode (GOTV/Demo)"
Write-Host "  3. Run the executable (GUI or CLI mode)"
Write-Host ""

if (-not $csConfigPath) {
    Write-Host "IMPORTANT:" -ForegroundColor Yellow
    Write-Host "- Copy config\gamestate_integration_autodirector.cfg to your CS cfg folder"
}

Write-Host ""
Read-Host "Press Enter to exit"
