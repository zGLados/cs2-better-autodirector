# Build CS2 Better Auto Director Installer
# This script builds the GUI application and then creates an installer using Inno Setup

Write-Host "====================================================" -ForegroundColor Cyan
Write-Host "  CS2 Better Auto Director - Installer Builder" -ForegroundColor Cyan
Write-Host "====================================================" -ForegroundColor Cyan
Write-Host ""

# Refresh PATH environment variable to include Node.js
$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

# Check if Inno Setup is installed
$InnoSetupPaths = @(
    "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
    "C:\Program Files\Inno Setup 6\ISCC.exe",
    "C:\Program Files (x86)\Inno Setup 5\ISCC.exe",
    "C:\Program Files\Inno Setup 5\ISCC.exe"
)

$InnoSetupPath = $null
foreach ($path in $InnoSetupPaths) {
    if (Test-Path $path) {
        $InnoSetupPath = $path
        break
    }
}

if (-not $InnoSetupPath) {
    Write-Host "ERROR: Inno Setup not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please download and install Inno Setup from:" -ForegroundColor Yellow
    Write-Host "https://jrsoftware.org/isdl.php" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "After installation, run this script again." -ForegroundColor Yellow
    exit 1
}

Write-Host "[1/3] Found Inno Setup at: $InnoSetupPath" -ForegroundColor Green
Write-Host ""

# Check if Wails is available
$wailsPath = Get-Command wails -ErrorAction SilentlyContinue
if (-not $wailsPath) {
    Write-Host "ERROR: Wails not found!" -ForegroundColor Red
    Write-Host "Please install Wails first: https://wails.io/docs/gettingstarted/installation" -ForegroundColor Yellow
    exit 1
}

Write-Host "[2/3] Building GUI application..." -ForegroundColor Yellow
Write-Host ""

# Build the GUI
Set-Location "$PSScriptRoot\..\gui"
$buildResult = & wails build -skipbindings 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: GUI build failed!" -ForegroundColor Red
    Write-Host $buildResult
    exit 1
}

Write-Host "GUI build completed successfully!" -ForegroundColor Green
Write-Host ""

# Return to scripts folder
Set-Location $PSScriptRoot

# Build the installer
Write-Host "[3/3] Building installer with Inno Setup..." -ForegroundColor Yellow
Write-Host ""

$installerScript = "$PSScriptRoot\installer.iss"
$buildOutput = & $InnoSetupPath $installerScript 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Installer build failed!" -ForegroundColor Red
    Write-Host $buildOutput
    exit 1
}

Write-Host "====================================================" -ForegroundColor Green
Write-Host "  Installer built successfully!" -ForegroundColor Green
Write-Host "====================================================" -ForegroundColor Green
Write-Host ""

# Find the output file
$outputFile = Get-Item "$PSScriptRoot\..\build\CS2BetterAutoDirector-Setup.exe" -ErrorAction SilentlyContinue
if ($outputFile) {
    Write-Host "Installer location:" -ForegroundColor Cyan
    Write-Host $outputFile.FullName -ForegroundColor White
    Write-Host ""
    Write-Host "Size: $([math]::Round($outputFile.Length / 1MB, 2)) MB" -ForegroundColor Gray
    Write-Host ""
} else {
    Write-Host "Installer should be in: build\CS2BetterAutoDirector-Setup.exe" -ForegroundColor Cyan
}

Write-Host "You can now distribute this installer to install the application" -ForegroundColor Yellow
Write-Host "on any Windows PC. It will:" -ForegroundColor Yellow
Write-Host "  - Install to C:\Program Files\CS2BetterAutoDirector" -ForegroundColor Gray
Write-Host "  - Create Start Menu shortcuts" -ForegroundColor Gray
Write-Host "  - Optionally create Desktop icon" -ForegroundColor Gray
Write-Host "  - Automatically copy GSI config to CS2 (if found)" -ForegroundColor Gray
Write-Host ""
