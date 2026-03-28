# Building Instructions

## Quick Start

### 1. Install Requirements

**Windows:**
```powershell
# Go (https://go.dev/dl/), Node.js (https://nodejs.org/), then:
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**Linux (Ubuntu/Debian 24.04+):**
```bash
sudo apt install golang nodejs npm build-essential libgtk-3-dev libwebkit2gtk-4.1-dev
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 2. Build

**Build App:**
```powershell
# Windows:
.\scripts\build.ps1

# Linux:
./scripts/build.sh
```

**Installer/Package erstellen:**
```powershell
# Windows Installer:
.\scripts\build-installer.ps1
```

### 3. Fertig!

**Ausgabe:**
- App: `gui/build/bin/cs2-better-autodirector[.exe]`
- Installer: `build/CS2BetterAutoDirector-Setup.exe`
- Linux Package: `build/CS2BetterAutoDirector-Linux-x64.tar.gz`

---

## Detailed Instructions

### Requirements

#### Windows

1. **Go 1.22+**: https://go.dev/dl/
2. **Node.js LTS**: https://nodejs.org/
3. **Wails CLI v2.11+**:
   ```powershell
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

Restart terminal after installation.

**Optional:**
- **Inno Setup 6** (für Installer): https://jrsoftware.org/isdl.php

#### Linux

1. **Go 1.22+**: https://go.dev/dl/
2. **Node.js LTS**: https://nodejs.org/
3. **Build Dependencies**:
   ```bash
   # Ubuntu/Debian (24.04+)
   sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.1-dev
   
   # Symlink für Wails-Kompatibilität:
   sudo ln -sf /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.1.pc \
               /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
   
   # Ubuntu/Debian (22.04 und älter)
   sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev
   ```

4. **Wails CLI v2.11+**:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   export PATH="$PATH:$(go env GOPATH)/bin"
   ```

**Check:**
```bash
wails doctor
```

### Manual Build

Without using the scripts:

**Windows:**
```powershell
cd gui
wails build -skipbindings
# Output: gui/build/bin/cs2-better-autodirector.exe
```

**Linux:**
```bash
cd gui
wails build -skipbindings
# Output: gui/build/bin/cs2-better-autodirector
```

**Development mode (hot-reload):**
```bash
cd gui
wails dev
```

---

## Windows Installer

### Erstellen

**Automated (empfohlen):**
```powershell
.\scripts\build-installer.ps1
```

**Manual:**
1. Inno Setup 6 installieren: https://jrsoftware.org/isdl.php
2. GUI builden: `cd gui; wails build -skipbindings`
3. Kompilieren: `& "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" .\scripts\installer.iss`

### Features

- ✅ User-only (kein Admin) oder system-wide Installation
- ✅ Automatische CS2-Pfad-Erkennung (Steam Registry)
- ✅ Auto-Copy von GSI-Config zum CS2-Ordner
- ✅ Debug Mode Shortcut (startet mit `-v` für Fehlerdiagnose)
- ✅ Start Menu Shortcuts + optionales Desktop Icon
- ✅ Saubere Deinstallation
- ✅ Multi-language (EN/DE)

### Silent Installation

Für Automatisierung/Enterprise:

```cmd
# Basic:
CS2BetterAutoDirector-Setup.exe /VERYSILENT

# System-wide mit Log:
CS2BetterAutoDirector-Setup.exe /VERYSILENT /ALLUSERS /LOG="install.log"

# Custom path:
CS2BetterAutoDirector-Setup.exe /VERYSILENT /DIR="D:\Games\CS2AutoDirector"
```

<details>
<summary>Alle Silent-Installation Parameter</summary>

| Parameter | Beschreibung | Beispiel |
|-----------|-------------|----------|
| `/VERYSILENT` | Komplett lautlos (keine UI) | `/VERYSILENT` |
| `/SILENT` | Lautlos mit Progress Bar | `/SILENT` |
| `/LOG="file"` | Installations-Log erstellen | `/LOG="install.log"` |
| `/DIR="path"` | Installations-Verzeichnis | `/DIR="C:\MyApps"` |
| `/TASKS="tasks"` | Tasks auswählen | `/TASKS="desktopicon,copygsiconfig"` |
| `/ALLUSERS` | Für alle User (Admin) | `/ALLUSERS` |
| `/CURRENTUSER` | Nur für aktuellen User | `/CURRENTUSER` |

**Task Namen:** `desktopicon`, `copygsiconfig`

**Exit Codes:** 0=Erfolg, 1=Fehlgeschlagen, 2=Abgebrochen, 3=Fatal Error

</details>

---

## Linux Package

### Erstellen

```bash
# App builden
./scripts/build.sh

# Tar.gz Package erstellen
./scripts/build-package.sh
```

**Output:** `build/CS2BetterAutoDirector-Linux-x64.tar.gz`

**Package enthält:**
- Binary + Configs + Install/Uninstall Scripts
- Desktop Entry Erstellung
- Auto-Erkennung des CS2 Config-Ordners

---

## Distribution

| Methode | Größe | Installation | Best For |
|---------|-------|--------------|----------|
| **Windows Installer** | ~5-7 MB | One-Click Wizard | End Users |
| **Portable .exe** | ~12 MB | Einfach starten | Quick Testing |
| **Linux Package** | ~12 MB | Install Script | Linux Users |
| **Source Build** | N/A | Manual | Developers |

---

## Troubleshooting

**"wails: command not found":**
```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
# Terminal neustarten
```

**Frontend Build schlägt fehl:**
```powershell
cd gui/frontend
npm install
cd ..
wails build
```

**Linux: "webkit2gtk-4.0.pc not found" (Ubuntu 24.04+):**
```bash
sudo ln -sf /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.1.pc \
            /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc
```

**Installer: "Inno Setup not found":**
```powershell
# Auto-install:
.\scripts\install-innosetup.ps1
# Oder manuell: https://jrsoftware.org/isdl.php
```

**Installer kann exe nicht finden:**
```powershell
# GUI erst builden:
cd gui
wails build -skipbindings
# Check: gui/build/bin/cs2-better-autodirector.exe muss existieren
```

---

## GitHub Actions

Automatische Builds bei Version Tags.

**Release erstellen:**
```bash
git tag -a v0.3.1 -m "Version 0.3.1"
git push origin v0.3.1
# → Builds Windows + Linux, erstellt Release automatisch
```

**Manual Build (Testing):** 
Actions → "Build Windows Installer" / "Build Linux Package" → Run workflow

Details: [.github/workflows/README.md](../.github/workflows/README.md)

---

<details>
<summary>Advanced: Installer Customization</summary>

Das Installer-Script liegt in `scripts/installer.iss` und kann angepasst werden:

**Installation Path ändern:**
```inno
DefaultDirName={autopf}\CS2BetterAutoDirector
```

**Shortcuts hinzufügen/entfernen:**
```inno
[Icons]
Name: "{group}\YourApp"; Filename: "{app}\yourapp.exe"
```

**CS2 Detection Logic anpassen:**
Edit `GetCS2ConfigPath()` function in `[Code]` section.

**Komprimierung ändern:**
```inno
Compression=lzma2/max  ; Best compression (langsamer)
; oder
Compression=lzma/fast  ; Schnelleres Builden
```

Inno Setup Docs: https://jrsoftware.org/ishelp/

</details>
