# Download and Install Node.js LTS
# This script automatically downloads and installs Node.js

Write-Host "====================================================" -ForegroundColor Cyan
Write-Host "  Node.js LTS - Automatic Installer" -ForegroundColor Cyan
Write-Host "====================================================" -ForegroundColor Cyan
Write-Host ""

# Check if already installed
$nodeCheck = Get-Command node -ErrorAction SilentlyContinue
$npmCheck = Get-Command npm -ErrorAction SilentlyContinue

if ($nodeCheck -and $npmCheck) {
    $nodeVersion = & node --version
    $npmVersion = & npm --version
    Write-Host "Node.js is already installed!" -ForegroundColor Green
    Write-Host "  Node.js: $nodeVersion" -ForegroundColor White
    Write-Host "  npm: $npmVersion" -ForegroundColor White
    Write-Host ""
    Write-Host "You can now build the GUI with:" -ForegroundColor Yellow
    Write-Host "  cd gui" -ForegroundColor Cyan
    Write-Host "  wails build -skipbindings" -ForegroundColor Cyan
    exit 0
}

Write-Host "Node.js not found. Starting download..." -ForegroundColor Yellow
Write-Host ""

# Node.js LTS download URL (v20.x LTS)
$downloadUrl = "https://nodejs.org/dist/v20.11.0/node-v20.11.0-x64.msi"
$installerPath = "$env:TEMP\nodejs-installer.msi"

try {
    Write-Host "[1/3] Downloading Node.js LTS (v20.11.0)..." -ForegroundColor Cyan
    Write-Host "From: $downloadUrl" -ForegroundColor Gray
    Write-Host "To: $installerPath" -ForegroundColor Gray
    Write-Host ""
    Write-Host "This may take a few minutes..." -ForegroundColor Yellow
    
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
    
    Write-Host "[2/3] Installing Node.js..." -ForegroundColor Cyan
    Write-Host "This will install Node.js and npm globally." -ForegroundColor Yellow
    Write-Host ""
    
    # Run MSI installer silently
    Write-Host "Running installer (this may take 1-2 minutes)..." -ForegroundColor Gray
    $installArgs = @(
        "/i"
        $installerPath
        "/quiet"
        "/norestart"
        "ADDLOCAL=ALL"
    )
    
    $process = Start-Process -FilePath "msiexec.exe" -ArgumentList $installArgs -Wait -PassThru
    
    if ($process.ExitCode -ne 0) {
        throw "Installation failed with exit code: $($process.ExitCode)"
    }
    
    Write-Host "Installation completed!" -ForegroundColor Green
    Write-Host ""
    
    Write-Host "[3/3] Verifying installation..." -ForegroundColor Cyan
    Write-Host ""
    
    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
    
    # Verify installation
    $nodeCheck = Get-Command node -ErrorAction SilentlyContinue
    $npmCheck = Get-Command npm -ErrorAction SilentlyContinue
    
    if ($nodeCheck -and $npmCheck) {
        $nodeVersion = & node --version
        $npmVersion = & npm --version
        
        Write-Host "====================================================" -ForegroundColor Green
        Write-Host "  Node.js installed successfully!" -ForegroundColor Green
        Write-Host "====================================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "Installed versions:" -ForegroundColor White
        Write-Host "  Node.js: $nodeVersion" -ForegroundColor Cyan
        Write-Host "  npm: $npmVersion" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "IMPORTANT: Please restart PowerShell/Terminal!" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "After restart, you can build the GUI with:" -ForegroundColor Yellow
        Write-Host "  cd gui" -ForegroundColor Cyan
        Write-Host "  wails build -skipbindings" -ForegroundColor Cyan
        Write-Host ""
    } else {
        Write-Host "Installation completed but verification failed." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Please:" -ForegroundColor Yellow
        Write-Host "  1. Close this terminal" -ForegroundColor Cyan
        Write-Host "  2. Open a new terminal" -ForegroundColor Cyan
        Write-Host "  3. Run: node --version" -ForegroundColor Cyan
        Write-Host ""
    }
    
    # Clean up
    if (Test-Path $installerPath) {
        Remove-Item $installerPath -Force
        Write-Host "Cleaned up temporary files." -ForegroundColor Gray
    }
    
} catch {
    Write-Host "ERROR: Failed to download or install Node.js" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    Write-Host "Please download and install manually from:" -ForegroundColor Yellow
    Write-Host "https://nodejs.org/en/download/" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Download the 'Windows Installer (.msi)' for x64" -ForegroundColor Gray
    exit 1
}
