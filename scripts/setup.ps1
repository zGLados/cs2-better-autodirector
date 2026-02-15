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
    Write-Host "Copying gamestate_integration_autoobserver.cfg..."
    
    try {
        Copy-Item "config\gamestate_integration_autoobserver.cfg" -Destination $csConfigPath -Force
        Write-Host "Config copied successfully!" -ForegroundColor Green
    } catch {
        Write-Host "WARNING: Could not copy config automatically." -ForegroundColor Yellow
        Write-Host "Please copy config\gamestate_integration_autoobserver.cfg manually to:" -ForegroundColor Yellow
        Write-Host $csConfigPath
    }
} else {
    Write-Host "Could not find Counter-Strike installation automatically." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please copy config\gamestate_integration_autoobserver.cfg manually to:" -ForegroundColor Yellow
    Write-Host "Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\"
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
    
    $goVersion = "1.22.0"
    $goInstaller = "go$goVersion.windows-amd64.msi"
    $goUrl = "https://go.dev/dl/$goInstaller"
    $goPath = "$env:TEMP\$goInstaller"
    
    Write-Host "Downloading Go $goVersion..."
    try {
        $ProgressPreference = 'SilentlyContinue'
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        
        # Try download with retry logic
        $maxRetries = 3
        $retryCount = 0
        $downloaded = $false
        
        while (-not $downloaded -and $retryCount -lt $maxRetries) {
            try {
                Invoke-WebRequest -Uri $goUrl -OutFile $goPath -UseBasicParsing -TimeoutSec 300
                $downloaded = $true
            } catch {
                $retryCount++
                if ($retryCount -lt $maxRetries) {
                    Write-Host "Download failed. Retrying ($retryCount/$maxRetries)..." -ForegroundColor Yellow
                    Start-Sleep -Seconds 2
                } else {
                    throw
                }
            }
        }
        
        Write-Host "Installing Go (this may take a moment)..."
        Start-Process msiexec.exe -ArgumentList "/i `"$goPath`" /quiet /norestart" -Wait
        
        Write-Host "Go installed successfully!" -ForegroundColor Green
        
        # Update PATH for current session by reading from registry
        $machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        $env:Path = "$machinePath;$userPath"
        
        # Verify Go is now available
        Start-Sleep -Seconds 2
        $goInstalled = Get-Command go -ErrorAction SilentlyContinue
        
        if (-not $goInstalled) {
            Write-Host "WARNING: Go was installed but is not yet available." -ForegroundColor Yellow
            Write-Host "Please run the script again in a new terminal." -ForegroundColor Yellow
            Read-Host "Press Enter to exit"
            exit 0
        }
    } catch {
        Write-Host "ERROR: Failed to download/install Go" -ForegroundColor Red
        Write-Host $_.Exception.Message -ForegroundColor Red
        Read-Host "Press Enter to exit"
        exit 1
    }
} else {
    $goVer = go version
    Write-Host "Go is already installed: $goVer" -ForegroundColor Green
}
Write-Host ""

# ============================================================
# 2. CHECK & INSTALL GCC
# ============================================================
Write-Host "[2/4] Checking GCC installation..." -ForegroundColor Green
$gccInstalled = Get-Command gcc -ErrorAction SilentlyContinue

if (-not $gccInstalled) {
    Write-Host "GCC is NOT installed. Installing TDM-GCC..." -ForegroundColor Yellow
    
    $gccVersion = "10.3.0-2"
    $gccInstaller = "tdm64-gcc-$gccVersion.exe"
    $gccUrl = "https://github.com/jmeubank/tdm-gcc/releases/download/v$gccVersion/$gccInstaller"
    $gccPath = "$env:TEMP\$gccInstaller"
    
    Write-Host "Downloading TDM-GCC $gccVersion..."
    try {
        $ProgressPreference = 'SilentlyContinue'
        
        # Try download with retry logic
        $maxRetries = 3
        $retryCount = 0
        $downloaded = $false
        
        while (-not $downloaded -and $retryCount -lt $maxRetries) {
            try {
                Invoke-WebRequest -Uri $gccUrl -OutFile $gccPath -UseBasicParsing -TimeoutSec 300
                $downloaded = $true
            } catch {
                $retryCount++
                if ($retryCount -lt $maxRetries) {
                    Write-Host "Download failed. Retrying ($retryCount/$maxRetries)..." -ForegroundColor Yellow
                    Start-Sleep -Seconds 2
                } else {
                    throw
                }
            }
        }
        
        Write-Host "Installing TDM-GCC..."
        Start-Process $gccPath -ArgumentList "/S" -Wait
        
        # Update PATH for current session
        $env:Path += ";C:\TDM-GCC-64\bin"
        
        Write-Host "GCC installed successfully!" -ForegroundColor Green
    } catch {
        Write-Host "ERROR: Failed to download/install GCC" -ForegroundColor Red
        Write-Host "Please install manually from: https://jmeubank.github.io/tdm-gcc/download/" -ForegroundColor Yellow
        Read-Host "Press Enter to exit"
        exit 1
    }
} else {
    $gccVer = gcc --version | Select-Object -First 1
    Write-Host "GCC is already installed: $gccVer" -ForegroundColor Green
}
Write-Host ""

# ============================================================
# 3. BUILD APPLICATION
# ============================================================
Write-Host "[3/4] Building better-autoobserver.exe..." -ForegroundColor Green
Write-Host ""

Write-Host "Downloading Go dependencies..."
try {
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
    Write-Host " 1. Close this window and run setup.exe again"
    Write-Host " 2. Restart your computer and run setup.exe again"
    Read-Host "Press Enter to exit"
    exit 1
}
Write-Host ""

# ============================================================
# 4. COPY GSI CONFIG
# ============================================================
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
    Write-Host "Copying gamestate_integration_autoobserver.cfg..."
    
    try {
        Copy-Item "gamestate_integration_autoobserver.cfg" -Destination $csConfigPath -Force
        Write-Host "Config copied successfully!" -ForegroundColor Green
    } catch {
        Write-Host "WARNING: Could not copy config automatically." -ForegroundColor Yellow
        Write-Host "Please copy gamestate_integration_autoobserver.cfg manually to:" -ForegroundColor Yellow
        Write-Host $csConfigPath
    }
} else {
    Write-Host "Could not find Counter-Strike installation automatically." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please copy gamestate_integration_autoobserver.cfg manually to:" -ForegroundColor Yellow
    Write-Host "Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\"
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
    Write-Host "- Copy gamestate_integration_autoobserver.cfg to your CS cfg folder"
}

Write-Host ""
Read-Host "Press Enter to exit"
