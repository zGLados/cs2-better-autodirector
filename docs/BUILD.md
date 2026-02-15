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

For professional distribution, you can create a complete Windows installer that:
- ✅ Installs to `C:\Program Files\CS2BetterAutoDirector`
- ✅ Creates Start Menu shortcuts
- ✅ Optionally creates Desktop icon
- ✅ Automatically copies GSI config to CS2 folder (if detected)
- ✅ Includes professional uninstaller

### Requirements:
1. **Inno Setup 6**: Download from [https://jrsoftware.org/isdl.php](https://jrsoftware.org/isdl.php)
2. Install Inno Setup (default location recommended)

### Create Installer (Automated):

**Option 1: Using PowerShell Script (Recommended)**
```powershell
.\scripts\build-installer.ps1
```

**Option 2: Using Batch File**
```cmd
.\scripts\build-installer.bat
```

This will:
1. Check if Inno Setup is installed
2. Build the GUI application with Wails
3. Compile the installer with Inno Setup

The installer will be created as: `build\CS2BetterAutoDirector-Setup.exe`

### Create Installer (Manual):

1. Build the GUI first:
   ```powershell
   cd gui
   wails build -skipbindings
   ```

2. Open **Inno Setup Compiler**
3. Open the file `scripts\installer.iss`
4. Click **Build → Compile**

### Installer Features:

The created installer (`CS2BetterAutoDirector-Setup.exe`):
- **Size**: ~15-20 MB (compressed)
- **Installs to**: `C:\Program Files\CS2BetterAutoDirector\`
- **Shortcuts**: Start Menu + optional Desktop
- **Smart GSI Config**: Automatically detects Steam installation and copies GSI config
- **Clean Uninstall**: Professional uninstaller included

---

## Distribution Comparison:

| Method | Use Case | Pros | Cons |
|--------|----------|------|------|
| **Portable .exe** | Quick testing / Development | Fast, simple | Manual config copy |
| **Installer .exe** | Distribution to users | Professional, auto-config | Requires Inno Setup |
| **Shortcut** | Local development | Convenient access | Not for distribution |

**Recommendation:**
- **For yourself:** Use shortcut in root folder (`cs2-better-autodirector.lnk`)
- **For others:** Build and distribute the installer (`CS2BetterAutoDirector-Setup.exe`)
