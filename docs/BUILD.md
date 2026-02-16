# Building Instructions

## How to Build (Wails GUI Version)

### Requirements

#### Windows

Before building, make sure you have installed:

1. **Go 1.22 or higher**: https://go.dev/dl/
2. **Node.js LTS (for frontend)**: https://nodejs.org/
3. **Wails CLI v2.11+**:
   ```powershell
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

After installation, restart your terminal/PowerShell.

**Optional (but recommended):**
- **UPX** (for compression): https://upx.github.io/
- **NSIS** (for installer): https://nsis.sourceforge.io/

#### Linux

Before building, make sure you have installed:

1. **Go 1.22 or higher**: https://go.dev/dl/
   ```bash
   # Ubuntu/Debian
   sudo apt install golang
   
   # Fedora
   sudo dnf install golang
   
   # Arch
   sudo pacman -S go
   ```

2. **Node.js LTS**: https://nodejs.org/
   ```bash
   # Ubuntu/Debian
   sudo apt install nodejs npm
   
   # Fedora
   sudo dnf install nodejs npm
   
   # Arch
   sudo pacman -S nodejs npm
   ```

3. **Build Dependencies**:
   ```bash
   # Ubuntu/Debian
   sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev
   
   # Fedora
   sudo dnf install gtk3-devel webkit2gtk3-devel
   
   # Arch
   sudo pacman -S gtk3 webkit2gtk
   ```

4. **Wails CLI v2.11+**:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   
   # Add Go bin to PATH
   export PATH="$PATH:$(go env GOPATH)/bin"
   # Add this line to ~/.bashrc or ~/.zshrc
   ```

After installation, restart your terminal

### Build the Application

#### Windows

**Method 1: Using Automated Script (Easiest)**

```powershell
.\scripts\build.ps1
```

**Method 2: Using Wails Directly**

```powershell
cd gui
wails build
```

The executable will be created in `gui/build/bin/cs2-better-autodirector.exe`

**Method 3: Skip Frontend Bindings (Faster)**

```powershell
cd gui
wails build -skipbindings
```

**Method 4: Development Mode (with hot-reload)**

```powershell
cd gui
wails dev
```

This opens the app in development mode with automatic reload on code changes.

#### Linux

**Method 1: Using Automated Script (Easiest)**

```bash
chmod +x ./scripts/build.sh
./scripts/build.sh
```

**Method 2: Using Wails Directly**

```bash
cd gui
wails build
```

The executable will be created in `gui/build/bin/cs2-better-autodirector`

**Method 3: Skip Frontend Bindings (Faster)**

```bash
cd gui
wails build -skipbindings
```

**Method 4: Development Mode (with hot-reload)**

```bash
cd gui
wails dev
```

**Method 5: Create Distributable Package**

```bash
# First build the application
./scripts/build.sh

# Then create tar.gz package
chmod +x ./scripts/build-package.sh
./scripts/build-package.sh
```

This creates `build/CS2BetterAutoDirector-Linux-x64.tar.gz` with installer scripts.

### Build Output

The compiled executable:
- **Location**: `cs2-autodirector-gui/build/bin/cs2-better-autodirector.exe`
- **Size**: ~12 MB (includes embedded frontend)
- **Runs in**: 
  - GUI mode (default): Opens dashboard window
  - CLI mode: `cs2-better-autodirector.exe -nogui`

### Verify Installation

Check if all requirements are installed:

```powershell
wails doctor
```

This shows:
- ✅ Go version
- ✅ Node.js version
- ✅ WebView2 status
- ⚠️ Optional dependencies (UPX, NSIS)

---

## Building a Windows Installer

For professional distribution, you can create a complete Windows installer with automatic dependency installation and configuration.

### Installer Features:

The installer (`CS2BetterAutoDirector-Setup.exe`) provides:
- ✅ **Dual Installation Mode**: Choose between user-only or system-wide installation
- ✅ **Automatic Dependencies**: Auto-downloads and installs Inno Setup and Node.js if needed
- ✅ **Smart CS2 Detection**: Automatically finds CS2 installation via Steam registry
- ✅ **Auto Config Copy**: Optionally copies GSI config to CS2 folder (enabled by default)
- ✅ **Manual Path Override**: Browse/enter CS2 path manually if auto-detection fails
- ✅ **Admin Elevation**: Automatically runs with admin rights when installed to Program Files
- ✅ **Start Menu Shortcuts**: Creates shortcuts in Start Menu and Programs folder
- ✅ **Optional Desktop Icon**: Choose to create desktop shortcut during installation
- ✅ **GitHub Repository Link**: Quick access to project repository
- ✅ **Clean Uninstallation**: Removes all files including logs folder
- ✅ **Multi-Language**: Supports English and German

### Installation Modes:

**User-Only Installation** (Default, no admin required):
- Installs to: `%LOCALAPPDATA%\Programs\CS2BetterAutoDirector`
- App runs without admin rights
- Perfect for single-user systems

**System-Wide Installation** (Requires admin):
- Installs to: `C:\Program Files\CS2BetterAutoDirector`
- App automatically requests admin rights on every launch
- Shortcuts configured for admin execution
- Perfect for multi-user systems

### Requirements:

The build script will automatically install missing dependencies:
1. **Inno Setup 6** (auto-downloaded if missing)
2. **Node.js 20+** (auto-downloaded if missing)
3. **Go 1.21+** (must be pre-installed)
4. **Wails CLI** (must be pre-installed)

### Create Installer (Automated):

**Option 1: Using PowerShell Script (Recommended)**
```powershell
.\scripts\build-installer.ps1
```

**Option 2: Using Batch File**
```cmd
.\scripts\build-installer.bat
```

**What the script does:**
1. Checks for Inno Setup (auto-installs if missing)
2. Checks for Node.js (auto-installs if missing)
3. Refreshes environment PATH variables
4. Builds the GUI application with Wails (`wails build -skipbindings`)
5. Compiles the installer with Inno Setup
6. Creates `build\CS2BetterAutoDirector-Setup.exe`

**First-time build:**
- May take 5-10 minutes (downloads dependencies ~100MB)
- Subsequent builds: ~30 seconds

### Create Installer (Manual):

1. **Install Inno Setup 6** manually:
   ```powershell
   .\scripts\install-innosetup.ps1
   ```

2. **Install Node.js** (if needed):
   ```powershell
   .\scripts\install-nodejs.ps1
   ```

3. **Build the GUI**:
   ```powershell
   cd gui
   wails build -skipbindings
   ```

4. **Compile the installer**:
   ```powershell
   & "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" .\scripts\installer.iss
   ```

### Installer Output:

The created installer:
- **Filename**: `CS2BetterAutoDirector-Setup.exe`
- **Location**: `build\` folder
- **Size**: ~5-7 MB (LZMA2/max compression)
- **Includes**: Application exe, config files, uninstaller

### Testing the Installer:

**Test user-only installation:**
1. Run `CS2BetterAutoDirector-Setup.exe`
2. Choose "Install for me only" in the installation dialog
3. Follow the wizard (CS2 path detection, config copy, etc.)
4. App should start without admin prompt

**Test system-wide installation:**
1. Run `CS2BetterAutoDirector-Setup.exe` as administrator
2. Choose "Install for all users" in the installation dialog
3. Follow the wizard
4. App should request admin rights on every launch

**Verify uninstallation:**
1. Use Windows "Add or Remove Programs"
2. Uninstall "CS2 Better Auto Director"
3. Check that installation folder is completely removed
4. Verify Start Menu shortcuts are deleted

---

## Command-Line Installation (Silent/Unattended)

The installer supports full command-line installation for automation, deployment scripts, or enterprise environments.

### Silent Installation

**Basic silent install (user-only, default settings):**
```cmd
CS2BetterAutoDirector-Setup.exe /VERYSILENT
```

**Silent install with log file:**
```cmd
CS2BetterAutoDirector-Setup.exe /VERYSILENT /LOG="install.log"
```

**System-wide installation (requires admin):**
```cmd
CS2BetterAutoDirector-Setup.exe /VERYSILENT /ALLUSERS
```

**Custom installation directory:**
```cmd
CS2BetterAutoDirector-Setup.exe /VERYSILENT /DIR="D:\Games\CS2AutoDirector"
```

### Available Command-Line Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `/VERYSILENT` | Completely silent (no UI) | `/VERYSILENT` |
| `/SILENT` | Silent with progress bar | `/SILENT` |
| `/SUPPRESSMSGBOXES` | No message boxes | `/SUPPRESSMSGBOXES` |
| `/LOG="file"` | Create installation log | `/LOG="C:\install.log"` |
| `/DIR="path"` | Installation directory | `/DIR="C:\MyApps\AutoDirector"` |
| `/GROUP="name"` | Start Menu folder | `/GROUP="CS2 Tools"` |
| `/NOICONS` | Don't create shortcuts | `/NOICONS` |
| `/TASKS="task1,task2"` | Select tasks | `/TASKS="desktopicon,copygsiconfig"` |
| `/ALLUSERS` | Install for all users (admin) | `/ALLUSERS` |
| `/CURRENTUSER` | Install for current user only | `/CURRENTUSER` |
| `/NORESTART` | Don't restart PC | `/NORESTART` |

### Task Names

Use with `/TASKS="task1,task2"`:
- `desktopicon` - Create desktop shortcut
- `copygsiconfig` - Copy GSI config to CS2 folder

**Examples:**
```cmd
# Desktop icon + config copy
CS2BetterAutoDirector-Setup.exe /VERYSILENT /TASKS="desktopicon,copygsiconfig"

# No desktop icon, no config copy
CS2BetterAutoDirector-Setup.exe /VERYSILENT /TASKS=""

# Only config copy (no desktop icon)
CS2BetterAutoDirector-Setup.exe /VERYSILENT /TASKS="copygsiconfig"
```

### Complete Examples

**Enterprise deployment (silent, system-wide, custom path, with log):**
```cmd
CS2BetterAutoDirector-Setup.exe /VERYSILENT /ALLUSERS /DIR="C:\Program Files\CS2Tools\AutoDirector" /LOG="C:\Logs\autodirector-install.log" /SUPPRESSMSGBOXES
```

**User deployment (silent, default location, desktop icon, config copy):**
```cmd
CS2BetterAutoDirector-Setup.exe /VERYSILENT /CURRENTUSER /TASKS="desktopicon,copygsiconfig"
```

**Testing installation (progress bar visible, custom location):**
```cmd
CS2BetterAutoDirector-Setup.exe /SILENT /DIR="D:\Test\AutoDirector" /LOG="test-install.log"
```

### Silent Uninstallation

**Uninstall silently:**
```cmd
# Find uninstaller in installation directory
"C:\Program Files\CS2BetterAutoDirector\unins000.exe" /VERYSILENT
```

**Or via registry:**
```powershell
# Get uninstall command from registry
$uninstall = Get-ItemProperty "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*" | Where-Object { $_.DisplayName -eq "CS2 Better Auto Director" }
$uninstallString = $uninstall.UninstallString
# Run silently
Start-Process $uninstallString -ArgumentList "/VERYSILENT" -Wait
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Setup failed |
| 2 | User cancelled |
| 3 | Fatal error |

### Deployment Scripts

**PowerShell deployment script:**
```powershell
# Download and install
$url = "https://github.com/zGLados/cs2-better-autodirector/releases/latest/download/CS2BetterAutoDirector-Setup.exe"
$installer = "$env:TEMP\CS2AutoDirector-Setup.exe"

# Download
Invoke-WebRequest -Uri $url -OutFile $installer

# Install silently for current user
$process = Start-Process -FilePath $installer -ArgumentList "/VERYSILENT /CURRENTUSER /TASKS=copygsiconfig /LOG=`"$env:TEMP\install.log`"" -PassThru -Wait

# Check exit code
if ($process.ExitCode -eq 0) {
    Write-Host "Installation successful!"
} else {
    Write-Host "Installation failed with exit code: $($process.ExitCode)"
    Get-Content "$env:TEMP\install.log"
}

# Cleanup
Remove-Item $installer
```

**Batch script for mass deployment:**
```batch
@echo off
REM Download installer (use your actual URL)
echo Downloading CS2 Better Auto Director...
curl -L -o "%TEMP%\CS2AutoDirector-Setup.exe" "https://github.com/USER/REPO/releases/latest/download/CS2BetterAutoDirector-Setup.exe"

REM Install silently
echo Installing...
"%TEMP%\CS2AutoDirector-Setup.exe" /VERYSILENT /CURRENTUSER /TASKS="copygsiconfig" /SUPPRESSMSGBOXES /NORESTART

REM Wait for installation
timeout /t 30 /nobreak

REM Cleanup
del "%TEMP%\CS2AutoDirector-Setup.exe"

echo Installation complete!
pause
```

### Checking Installation Status

**PowerShell - Check if installed:**
```powershell
# Check via registry
$installed = Get-ItemProperty "HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*" | Where-Object { $_.DisplayName -eq "CS2 Better Auto Director" }
if ($installed) {
    Write-Host "Installed at: $($installed.InstallLocation)"
    Write-Host "Version: $($installed.DisplayVersion)"
} else {
    Write-Host "Not installed"
}

# Or check if exe exists
if (Test-Path "$env:LOCALAPPDATA\Programs\CS2BetterAutoDirector\cs2-better-autodirector.exe") {
    Write-Host "Found user installation"
}
if (Test-Path "C:\Program Files\CS2BetterAutoDirector\cs2-better-autodirector.exe") {
    Write-Host "Found system-wide installation"
}
```

---

## Distribution Comparison:

| Method | Size | Installation | Config Setup | Admin Rights | Best For |
|--------|------|--------------|--------------|--------------|----------|
| **Windows Installer** | ~5-7 MB | One-click, dual-mode | Auto-detected & copied | Smart (only when needed) | **End users, distribution** |
| **Portable .exe** | ~12 MB | None (just run) | Manual copy | Only if in Program Files | Quick testing, portable use |
| **Build from Source** | N/A | Manual setup | Manual copy | No | Developers, customization |

### Installer Advantages:

**Windows Installer (`CS2BetterAutoDirector-Setup.exe`):**
- ✅ Professional installation wizard
- ✅ Automatic dependency installation (Inno Setup, Node.js)
- ✅ Steam/CS2 auto-detection via registry
- ✅ Automatic config file deployment
- ✅ User choice: AppData (no admin) vs Program Files (admin)
- ✅ Start Menu integration with shortcuts
- ✅ Clean uninstallation (removes logs)
- ✅ Multi-language support (EN/DE)

**Portable .exe (`cs2-better-autodirector.exe`):**
- ✅ No installation needed
- ✅ Run from any location (USB drive, Downloads, etc.)
- ✅ Perfect for testing
- ⚠️ Requires manual config copy
- ⚠️ No Start Menu integration
- ⚠️ Logs stay in program folder

### Recommendations:

**For distribution to others:**
→ Use the **Windows Installer** (`CS2BetterAutoDirector-Setup.exe`)

**For personal development/testing:**
→ Use the **Portable .exe** from `gui/build/bin/`

**For advanced users/developers:**
→ **Build from source** for latest features and customization

---

## Troubleshooting

### Common Build Issues:

**"Node.js not found" error:**
```powershell
# Install Node.js automatically
.\scripts\install-nodejs.ps1
# Or download manually from https://nodejs.org/
```

**"Inno Setup not found" error:**
```powershell
# Install Inno Setup automatically
.\scripts\install-innosetup.ps1
# Or download manually from https://jrsoftware.org/isdl.php
```

**"wails: command not found":**
```powershell
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
# Restart terminal/PowerShell
```

**Frontend build fails:**
```powershell
# Clean and rebuild
cd gui/frontend
npm install
cd ..
wails build
```

**Installer can't find exe:**
```powershell
# Ensure exe is built first
cd gui
wails build -skipbindings
# Check that gui/build/bin/cs2-better-autodirector.exe exists
```

### Installer Issues:

**"App won't start after installation to Program Files":**
- This is expected - the app requires admin rights in Program Files
- Shortcuts are automatically configured to request admin
- Or install "for me only" to avoid admin requirement

**"Config not copied automatically":**
- Ensure you checked "Copy Game State Integration config" during installation
- Or copy manually from installation folder to CS2 cfg folder

**"Uninstaller leaves folder behind":**
- Should not happen with current version
- If it does, manually delete the installation folder
- Report as bug on GitHub

---

## Advanced: Customizing the Installer

The installer script is located at `scripts/installer.iss` and can be customized:

**Change installation path:**
```inno
DefaultDirName={autopf}\CS2BetterAutoDirector  ; Auto-selects based on mode
; Or use fixed paths:
; {pf}\YourFolder     = C:\Program Files\YourFolder
; {localappdata}     = %LOCALAPPDATA%\YourFolder
```

**Add/remove shortcuts:**
```inno
[Icons]
Name: "{group}\YourApp"; Filename: "{app}\yourapp.exe"
```

**Modify CS2 detection logic:**
Edit the `GetCS2ConfigPath()` function in the `[Code]` section to change detection logic.

**Change compression:**
```inno
Compression=lzma2/max  ; Best compression (slower)
; or
Compression=lzma/fast  ; Faster compilation
```

For detailed Inno Setup documentation, see: https://jrsoftware.org/ishelp/

---

## Automated Builds with GitHub Actions

GitHub Actions automatically builds installers on version tags and creates releases.

### Quick Start

**Create a release:**
```bash
# 1. Update version, commit changes
git add .
git commit -m "Release v3.2.0"
git push origin dev

# 2. Create and push tag
git tag -a v3.2.0 -m "Version 3.2.0 - Description"
git push origin v3.2.0

# → Installer automatically attached to GitHub Release
```

**Manual build (testing):**
1. Go to: Repository → Actions → "Build Windows Installer"
2. Click "Run workflow" → Select branch → Run
3. Download from Artifacts section (~5-7 min)

### Workflow Behavior

| Trigger | Builds? | Creates Release? |
|---------|---------|------------------|
| Tag `v*.*.*` | ✅ Yes | ✅ Yes |
| Manual Run | ✅ Yes | ❌ No |
| Push to `dev` | ❌ No | ❌ No |

**Note:** Builds only on tags to save Actions minutes. For testing, build locally with `.\scripts\build-installer.bat`.

### Details

For workflow configuration, troubleshooting, and optimization options, see **[.github/workflows/README.md](../.github/workflows/README.md)**.

---