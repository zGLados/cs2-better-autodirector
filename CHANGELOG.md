# CS2 Better Auto Director - Changelog

## Version 2.0.0 (Current)

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

---

## Version 1.1.0

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

## Version 1.0.0

### Initial Release
- Intelligent encounter prediction based on player distance
- Priority-based player switching
- Game State Integration (GSI) support
- Standalone .exe build
- PowerShell build scripts
