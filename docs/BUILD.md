# Building Instructions

## How to Build (Wails GUI Version)

### Requirements

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

### Build the Application

**Method 1: Using Wails (Recommended)**

```powershell
cd cs2-autodirector-gui
wails build
```

The executable will be created in `cs2-autodirector-gui/build/bin/cs2-better-autodirector.exe`

**Method 2: Skip Frontend Bindings (Faster)**

```powershell
cd cs2-autodirector-gui
wails build -skipbindings
```

**Method 3: Development Mode (with hot-reload)**

```powershell
cd cs2-autodirector-gui
wails dev
```

This opens the app in development mode with automatic reload on code changes.

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