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

## Advanced: Creating an Installer with Inno Setup

For professional distribution, you can create a Windows installer.

### Requirements:
1. Install [Inno Setup](https://jrsoftware.org/isdl.php)
2. Build the .exe first with `scripts\build.bat`

### Create Installer:

1. Open **Inno Setup Compiler**
2. Open the file `scripts\installer.iss`
3. Click **Build → Compile**

This creates `BetterAutoObserver-Setup.exe` in the Output folder.

### The installer includes:
- ✅ Installs better-autoobserver.exe
- ✅ Automatically copies GSI config to CS folder
- ✅ Creates desktop shortcuts
- ✅ Adds to Start Menu
- ✅ Professional uninstaller

---

## Comparison:

| Feature | build.ps1 | Inno Setup Installer |
|---------|-----------|----------------------|
| Quick build | ✅ Yes | ❌ No |
| For development | ✅ Perfect | ❌ Overkill |
| For distribution | ✅ Good | ✅ Professional |
| Uninstaller | ❌ No | ✅ Yes |
| Size | Small | Medium (~10-20 MB) |

**Recommendation:**
- **Development:** Use `build.ps1` 
- **Distribution to users:** Use Inno Setup Installer
