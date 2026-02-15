# Building Instructions

## How to Build

### Requirements

Before building, make sure you have installed:

1. **Go 1.21 or higher**: https://go.dev/dl/
2. **GCC Compiler (TDM-GCC recommended)**: https://jmeubank.github.io/tdm-gcc/download/

After installation, restart your terminal/PowerShell.

### Build the Application

```powershell
cd scripts
.\build.ps1
```

Or double-click `scripts\build.bat`

This will:
- ✅ Check if Go and GCC are installed
- ✅ Download Go dependencies
- ✅ Compile `cs2-better-autodirector.exe`

The executable will be created in the root directory.

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
