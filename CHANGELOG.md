# CS2 Better Auto Director - Changelog

## Version 0.2.1 (Current)

### New Features
- **Settings Import/Export**: Export and import camera priority settings as JSON files
- Settings reset button now uses backend method for consistency

### Changes
- Documentation cleanup and optimization
- Removed obsolete root Go files after Wails migration
- Updated GitHub Actions workflow (builds only on tags)
- Improved installer and build documentation
- Installer default language set to English

---

## Version 0.1.0 - Professional Installer

### 🚀 Windows Installer Package

**New Installation Method:**
- **Professional Windows Installer**: `CS2BetterAutoDirector-Setup.exe` (~5-7 MB)
- **Automated Build Pipeline**: One-click compilation with automatic dependency installation
- **Dual Installation Modes**: User-only (AppData) or system-wide (Program Files)

**Installer Features:**
- ✅ **Automatic Dependency Installation**: Auto-downloads Inno Setup 6 and Node.js if missing
- ✅ **Smart CS2 Detection**: Finds CS2 installation via Steam registry
- ✅ **Automatic Config Copy**: Optionally copies GSI config to CS2 folder (enabled by default)
- ✅ **Manual Path Override**: Browse/enter CS2 path manually if auto-detection fails
- ✅ **Admin Elevation**: Automatically requests admin rights when installed to Program Files
- ✅ **Start Menu Integration**: Creates shortcuts in Start Menu and Programs folder
- ✅ **Optional Desktop Icon**: Choose to create desktop shortcut during installation
- ✅ **GitHub Repository Link**: Quick access to project repository from Start Menu
- ✅ **Clean Uninstallation**: Removes all files including logs folder
- ✅ **Multi-Language Support**: English and German interface
- ✅ **Command-Line Installation**: Full silent/unattended installation support
  - Silent install: `/VERYSILENT`
  - Custom directory: `/DIR="path"`
  - Task selection: `/TASKS="desktopicon,copygsiconfig"`
  - System-wide: `/ALLUSERS` or User-only: `/CURRENTUSER`
  - Deployment scripts for PowerShell and Batch included

**Installation Modes:**
1. **Install for me only** (Default):
   - Location: `%LOCALAPPDATA%\Programs\CS2BetterAutoDirector`
   - No admin rights required
   - App runs without elevation

2. **Install for all users** (Admin):
   - Location: `C:\Program Files\CS2BetterAutoDirector`
   - Requires admin installation
   - App automatically elevates with admin rights

**Build Scripts:**
- `scripts/build-installer.ps1` - Automated installer creation
- `scripts/build-installer.bat` - Batch wrapper
- `scripts/install-innosetup.ps1` - Auto-install Inno Setup 6
- `scripts/install-nodejs.ps1` - Auto-install Node.js 20+
- `scripts/installer.iss` - Inno Setup configuration (176 lines)

**GitHub Actions Integration:**
- `.github/workflows/build-installer.yml` - Automated CI/CD pipeline
- **Automatic builds on tags**: `git tag v3.1.0 && git push origin v3.1.0`
- **Manual workflow dispatch**: Build from GitHub UI
- **Release automation**: Installer automatically attached to GitHub Releases
- **Build artifacts**: 90-day retention for manual builds
- **Build time**: ~8 minutes on GitHub servers
- **No local dependencies**: Builds without Inno Setup/Node.js installed locally

**Documentation Updates:**
- Updated [README.md](README.md) with installation methods
- Updated [docs/BUILD.md](docs/BUILD.md) with detailed installer guide
- Added [.github/workflows/README.md](.github/workflows/README.md) with CI/CD documentation
- Added troubleshooting section
- Added customization guide


### 🎨 Major UI Update - Wails Dashboard

**Complete GUI Rewrite:**
- **Modern Dashboard Interface**: Built with Wails framework (native Windows app)
- **Live Status Widget**: Current player, uptime, total switches, alive players count
- **Real-time Player Table**: HP, Armor, K/D, Weapons, Money - updates every 500ms
- **Top Encounters Rankings**: Shows top 5 encounters with distance and priority scores
- **Live Event Log Stream**: Scrolling log of all switches, kills, detections
- **Statistics Dashboard**: Sniper kills, upset victories, damage detections, switches/min
- **Dark Gaming Theme**: Professional CS2-inspired design with gradients and animations
- **Real-time Updates**: WebSocket-like events using Wails Events system

**Dual-Mode Operation:**
```bash
cs2-better-autodirector.exe           # Starts GUI (default)
cs2-better-autodirector.exe -nogui    # Starts CLI mode
cs2-better-autodirector.exe -nogui -v # CLI with verbose logging
```

**Technical Stack:**
- Backend: Go with Wails v2.11.0
- Frontend: Vanilla JS + Vite + Custom CSS
- Size: ~12 MB (includes embedded frontend)
- No dependencies: Fully standalone .exe

**New Features:**
- ✅ Start/Stop controls in GUI
- ✅ Live player monitoring with team colors (CT/T)
- ✅ Encounter priority visualization
- ✅ Event timeline with timestamps
- ✅ Statistics tracking and display
- ✅ One executable for both GUI and CLI modes

**Migration:**
- Old CLI functionality preserved with `-nogui` flag
- Same configuration files (gamestate_integration_autodirector.cfg)
- Same port (3000) and GSI integration
- All existing features work in both modes

### 🎯 Major Features

#### Sniper System
- **Sniper Duel Detection**: Sniper vs Sniper encounters get +100 priority bonus
- **AWP vs Scout Differentiation**: AWP duels get +30 additional bonus
- **Kill Tracking**: 
  - AWP kills: +200 bonus for 8 seconds
  - Scout kills: +150 bonus for 8 seconds
- **Smart Immediate Switching**:
  - AWP kills → Immediate switch (unless watching another sniper)
  - Scout kills → Uses bonus system (avoids excessive jumping)
  - Respects ongoing sniper action (won't interrupt)
- **Weapon Priority**: AWP +30, Scout +15 in player selection

#### Damage Detection System
- **Real-time HP Tracking**: Monitors all players' health every tick
- **Damage Dealer Identification**:
  - Detects >20 HP loss
  - Finds nearest enemy with shooting weapon
  - Range limits: Normal weapons 1500 units, Snipers 3000 units
- **Smart Filtering**: Excludes grenades, molotovs, knives, C4
- **Immediate Switching**: Switches to damage dealer (+40 bonus for 5 seconds)

#### Upset Victory System
- **Underdog Detection**: Recognizes when non-spectated player wins encounter
- **Immediate Priority**: +100 bonus for 10 seconds
- **Smart Switching**: Forces immediate switch to winner

### 🔄 Project Restructure
- **New Name**: "CS2 Better Auto Director" (renamed from Better Auto Observer)
- **Updated Configs**: All configuration files renamed to `autodirector`
- **Build Output**: Now builds as `cs2-better-autodirector.exe`
- **Log Files**: Updated to `autodirector_[timestamp].log`

### 📊 Enhanced Priority System
- **Grenade Kill Filtering**: Ignores explosive/fire kills for prioritization
- **Weapon-Specific Bonuses**: Different bonuses for different weapon types
- **Time-based Decay**: Bonuses expire after set duration (5-10 seconds)

### 🎮 Improved Viewer Experience
- **Less Hektisch**: Smart switching prevents excessive camera jumping
- **Action-Focused**: Catches damage moments and kills in real-time
- **Sniper-Aware**: Understands high-stakes sniper gameplay

### 🐛 Bug Fixes
- Fixed immediate switching logic to respect active encounters
- Improved kill detection to ignore grenade/molotov kills
- Better weapon type detection and filtering

### New Features
- **Verbose logging mode**: Use `-v` flag for detailed console output
- **Log files**: Normal mode now writes detailed logs to `logs/` directory
- **Improved console output**: Cleaner, more informative messages
- **Position debugging**: Verbose mode shows player positions to diagnose distance calculation issues

### Logging Improvements
- **Normal mode**: Shows only essential info (data received, switches)
- **Verbose mode**: Shows all details (GSI data, analysis, positions, decisions)
- **Log files**: Automatic log files in `logs/autodirector_[timestamp].log`

### Console Output Examples

**Normal Mode:**
```
[INFO] Data received: 10 players
⚔️  ENCOUNTER: Player1 (CT) vs Player2 (T) | Distance: 450 units | Priority: 125.5
➡️  Switching to: Player1 (Slot 3)
```

**Verbose Mode (-v):**
```
[GSI] Round Phase: live
[ANALYZER] Found 10 players
[ANALYZER] Player Player1 position: X=1234.5 Y=567.8 Z=90.1
[ANALYZER] Detected 2 potential encounters
⚔️  ENCOUNTER: Player1 (CT) vs Player2 (T) | Distance: 450 units | Priority: 125.5
➡️  Switching to: Player1 (Slot 3)
```

### Bug Fixes
- Improved error messages for missing GSI data
- Better handling of missing position data

### Usage
- Run with verbose mode: `run.bat -v` or `cs2-better-autodirector.exe -v`
- Normal mode: `run.bat` or `cs2-better-autodirector.exe`
- Check logs: `logs/autodirector_[timestamp].log`

### Initial Release
- Intelligent encounter prediction based on player distance
- Priority-based player switching
- Game State Integration (GSI) support
- Standalone .exe build
- PowerShell build scripts
