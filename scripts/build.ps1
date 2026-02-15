# Better Auto Observer - Build Script

$ErrorActionPreference = "Stop"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Better Auto Observer - Build Script" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "Checking Go installation..." -ForegroundColor Green
$goInstalled = Get-Command go -ErrorAction SilentlyContinue

if (-not $goInstalled) {
    # Try common Go installation paths
    $commonGoPaths = @(
        "C:\Program Files\Go\bin\go.exe",
        "C:\Go\bin\go.exe",
        "$env:USERPROFILE\go\bin\go.exe"
    )
    
    foreach ($path in $commonGoPaths) {
        if (Test-Path $path) {
            Write-Host "Found Go at: $path" -ForegroundColor Yellow
            Write-Host "But it's not in your PATH environment variable." -ForegroundColor Yellow
            Write-Host ""
            Write-Host "SOLUTION: Please restart your terminal/PowerShell window!" -ForegroundColor Cyan
            Write-Host "The Go installer added it to PATH, but you need to reload the environment." -ForegroundColor White
            Write-Host ""
            Read-Host "Press Enter to exit"
            exit 1
        }
    }
    
    Write-Host ""
    Write-Host "ERROR: Go is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install the following requirements:" -ForegroundColor Yellow
    Write-Host "  1. Go 1.21 or higher: https://go.dev/dl/" -ForegroundColor White
    Write-Host "  2. GCC Compiler (TDM-GCC): https://jmeubank.github.io/tdm-gcc/download/" -ForegroundColor White
    Write-Host ""
    Write-Host "After installation, RESTART your terminal and run this script again." -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

$goVer = go version
Write-Host "Go found: $goVer" -ForegroundColor Green
Write-Host ""

# Check if GCC is installed
Write-Host "Checking GCC installation..." -ForegroundColor Green
$gccInstalled = Get-Command gcc -ErrorAction SilentlyContinue

if (-not $gccInstalled) {
    # Try common GCC installation paths
    $commonGccPaths = @(
        "C:\TDM-GCC-64\bin\gcc.exe",
        "C:\MinGW\bin\gcc.exe",
        "C:\msys64\mingw64\bin\gcc.exe",
        "C:\Program Files\TDM-GCC-64\bin\gcc.exe"
    )
    
    foreach ($path in $commonGccPaths) {
        if (Test-Path $path) {
            Write-Host "Found GCC at: $path" -ForegroundColor Yellow
            Write-Host "But it's not in your PATH environment variable." -ForegroundColor Yellow
            Write-Host ""
            Write-Host "SOLUTION: Please restart your terminal/PowerShell window!" -ForegroundColor Cyan
            Write-Host "The GCC installer should have added it to PATH." -ForegroundColor White
            Write-Host ""
            Write-Host "If that doesn't work, add this to your PATH manually:" -ForegroundColor White
            Write-Host (Split-Path $path) -ForegroundColor Cyan
            Write-Host ""
            Read-Host "Press Enter to exit"
            exit 1
        }
    }
    
    Write-Host ""
    Write-Host "ERROR: GCC not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "robotgo requires a C compiler to build." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please install one of the following:" -ForegroundColor White
    Write-Host "  1. TDM-GCC (recommended): https://jmeubank.github.io/tdm-gcc/download/" -ForegroundColor White
    Write-Host "  2. MinGW-w64: https://www.mingw-w64.org/" -ForegroundColor White
    Write-Host ""
    Write-Host "After installation, RESTART your terminal and run this script again." -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

$gccVer = gcc --version | Select-Object -First 1
Write-Host "GCC found: $gccVer" -ForegroundColor Green
Write-Host ""

# Download Go dependencies
Write-Host "Downloading Go dependencies..." -ForegroundColor Green
try {
    go mod tidy
    if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed" }
} catch {
    Write-Host "ERROR: Failed to download dependencies" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""
Write-Host "Building executable..." -ForegroundColor Green
try {
    # Go to root directory (one folder up)
    $rootDir = Split-Path -Parent $PSScriptRoot
    Set-Location $rootDir
    
    go build -ldflags="-s -w" -o better-autoobserver.exe
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }
    
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "SUCCESS! Created: better-autoobserver.exe" -ForegroundColor Green
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host ""
    
    if (Test-Path "cs2-better-autodirector.exe") {
        $fileSize = (Get-Item "cs2-better-autodirector.exe").Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        Write-Host "File size: $fileSizeMB MB" -ForegroundColor White
    }
    
    Write-Host ""
    Write-Host "You can now run the program with:" -ForegroundColor White
    Write-Host "    cs2-better-autodirector.exe" -ForegroundColor Cyan
    Write-Host ""
} catch {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}

Read-Host "Press Enter to exit"
