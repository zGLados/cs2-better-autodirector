# Better Auto Observer - Setup Script
# This script checks requirements and builds the application

$ErrorActionPreference = "Stop"

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host " Better Auto Observer - Setup Script" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "This script will:" -ForegroundColor White
Write-Host " - Check if Go and GCC are installed" -ForegroundColor White
Write-Host " - Build better-autoobserver.exe" -ForegroundColor White
Write-Host " - Copy GSI config to CS folder (if found)" -ForegroundColor White
Write-Host ""
Write-Host "Press any key to continue or Ctrl+C to cancel..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
Write-Host ""

# Check if Go is installed
Write-Host "[1/4] Checking Go installation..." -ForegroundColor Green
$goInstalled = Get-Command go -ErrorAction SilentlyContinue

if (-not $goInstalled) {
    Write-Host ""
    Write-Host "ERROR: Go is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install the following requirements:" -ForegroundColor Yellow
    Write-Host "  1. Go 1.21 or higher: https://go.dev/dl/" -ForegroundColor White
    Write-Host "  2. GCC Compiler (TDM-GCC): https://jmeubank.github.io/tdm-gcc/download/" -ForegroundColor White
    Write-Host ""
    Write-Host "After installation, run this script again." -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

$goVer = go version
Write-Host "Go is already installed: $goVer" -ForegroundColor Green
Write-Host ""

# Check if GCC is installed
Write-Host "[2/4] Checking GCC installation..." -ForegroundColor Green
$gccInstalled = Get-Command gcc -ErrorAction SilentlyContinue

if (-not $gccInstalled) {
    Write-Host ""
    Write-Host "ERROR: GCC not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "robotgo requires a C compiler to build." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please install one of the following:" -ForegroundColor White
    Write-Host "  1. TDM-GCC (recommended): https://jmeubank.github.io/tdm-gcc/download/" -ForegroundColor White
    Write-Host "  2. MinGW-w64: https://www.mingw-w64.org/" -ForegroundColor White
    Write-Host ""
    Write-Host "After installation, run this script again." -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

$gccVer = gcc --version | Select-Object -First 1
Write-Host "GCC is already installed: $gccVer" -ForegroundColor Green
Write-Host ""

# Build application
Write-Host "[3/4] Building better-autoobserver.exe..." -ForegroundColor Green
Write-Host ""

Write-Host "Downloading Go dependencies..."
try {
    # Go to root directory
    $rootDir = Split-Path -Parent $PSScriptRoot
    Set-Location $rootDir
    
    go mod tidy
    if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed" }
    
    Write-Host "Compiling executable..."
    go build -ldflags="-s -w" -o better-autoobserver.exe
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }
    
    Write-Host "Build successful!" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    Write-Host ""
    Write-Host "Possible solutions:" -ForegroundColor Yellow
    Write-Host " 1. Make sure Go and GCC are properly installed" -ForegroundColor White
    Write-Host " 2. Restart your computer and try again" -ForegroundColor White
    Write-Host " 3. Check the error message above for details" -ForegroundColor White
    Read-Host "Press Enter to exit"
    exit 1
}
Write-Host ""

# Copy GSI config
Write-Host "[4/4] Setting up Counter-Strike integration..." -ForegroundColor Green
Write-Host ""

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
        Copy-Item "config\gamestate_integration_autoobserver.cfg" -Destination $csConfigPath -Force
        Write-Host "✓ gamestate_integration_autoobserver.cfg copied" -ForegroundColor Green
        
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
        Write-Host "  - config\gamestate_integration_autoobserver.cfg"
        Write-Host "  - config\spectator_bindings.cfg"
    }
} else {
    Write-Host "Could not find Counter-Strike installation automatically." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please copy these files manually to:" -ForegroundColor Yellow
    Write-Host "Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\"
    Write-Host "  - config\gamestate_integration_autoobserver.cfg"
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

if (Test-Path "better-autoobserver.exe") {
    $fileSize = (Get-Item "better-autoobserver.exe").Length
    $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
    Write-Host "Created: better-autoobserver.exe ($fileSizeMB MB)" -ForegroundColor Green
}

Write-Host ""
Write-Host "To start the Auto Observer:" -ForegroundColor White
Write-Host "  1. Start Counter-Strike"
Write-Host "  2. Go into spectator mode"
Write-Host "  3. Run: better-autoobserver.exe"
Write-Host ""

if (-not $csConfigPath) {
    Write-Host "IMPORTANT:" -ForegroundColor Yellow
    Write-Host "- Copy config\gamestate_integration_autoobserver.cfg to your CS cfg folder"
}

Write-Host ""
Read-Host "Press Enter to exit"
