# Download and Install Inno Setup
# This script automatically downloads and installs Inno Setup 6

Write-Host "====================================================" -ForegroundColor Cyan
Write-Host "  Inno Setup 6 - Automatic Installer" -ForegroundColor Cyan
Write-Host "====================================================" -ForegroundColor Cyan
Write-Host ""

# Check if already installed
$InnoSetupPaths = @(
    "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
    "C:\Program Files\Inno Setup 6\ISCC.exe"
)

foreach ($path in $InnoSetupPaths) {
    if (Test-Path $path) {
        Write-Host "Inno Setup 6 is already installed at:" -ForegroundColor Green
        Write-Host $path -ForegroundColor White
        Write-Host ""
        Write-Host "You can now run: .\scripts\build-installer.bat" -ForegroundColor Yellow
        exit 0
    }
}

Write-Host "Inno Setup 6 not found. Starting download..." -ForegroundColor Yellow
Write-Host ""

# Download URL for Inno Setup 6
$downloadUrl = "https://jrsoftware.org/download.php/is.exe"
$installerPath = "$env:TEMP\innosetup-6-installer.exe"

try {
    Write-Host "[1/3] Downloading Inno Setup 6..." -ForegroundColor Cyan
    Write-Host "From: $downloadUrl" -ForegroundColor Gray
    Write-Host "To: $installerPath" -ForegroundColor Gray
    Write-Host ""
    
    # Download with progress
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $downloadUrl -OutFile $installerPath -UseBasicParsing
    $ProgressPreference = 'Continue'
    
    if (-not (Test-Path $installerPath)) {
        throw "Download failed - file not found"
    }
    
    $fileSize = (Get-Item $installerPath).Length / 1MB
    Write-Host "Download complete! Size: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Green
    Write-Host ""
    
    Write-Host "[2/3] Installing Inno Setup 6..." -ForegroundColor Cyan
    Write-Host "This will open the installer window." -ForegroundColor Yellow
    Write-Host "Please follow the installation wizard." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Recommended settings:" -ForegroundColor Gray
    Write-Host "  - Install location: C:\Program Files (x86)\Inno Setup 6" -ForegroundColor Gray
    Write-Host "  - Accept all default options" -ForegroundColor Gray
    Write-Host ""
    
    # Run installer
    Start-Process -FilePath $installerPath -Wait
    
    Write-Host "[3/3] Verifying installation..." -ForegroundColor Cyan
    Write-Host ""
    
    # Check if installed
    $installed = $false
    foreach ($path in $InnoSetupPaths) {
        if (Test-Path $path) {
            Write-Host "====================================================" -ForegroundColor Green
            Write-Host "  Inno Setup 6 installed successfully!" -ForegroundColor Green
            Write-Host "====================================================" -ForegroundColor Green
            Write-Host ""
            Write-Host "Location: $path" -ForegroundColor White
            Write-Host ""
            Write-Host "You can now build the installer with:" -ForegroundColor Yellow
            Write-Host "  .\scripts\build-installer.bat" -ForegroundColor Cyan
            Write-Host ""
            $installed = $true
            break
        }
    }
    
    if (-not $installed) {
        Write-Host "Installation appears incomplete." -ForegroundColor Yellow
        Write-Host "Please check if Inno Setup was installed correctly." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "If you cancelled the installation, run this script again." -ForegroundColor Gray
    }
    
    # Clean up
    if (Test-Path $installerPath) {
        Remove-Item $installerPath -Force
        Write-Host "Cleaned up temporary files." -ForegroundColor Gray
    }
    
} catch {
    Write-Host "ERROR: Failed to download or install Inno Setup" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    Write-Host "Please download manually from:" -ForegroundColor Yellow
    Write-Host "https://jrsoftware.org/isdl.php" -ForegroundColor Cyan
    exit 1
}
