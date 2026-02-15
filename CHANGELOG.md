# Better Auto Observer - Changelog

## Version 1.1.0 (Current)

### New Features
- **Verbose logging mode**: Use `-v` flag for detailed console output
- **Log files**: Normal mode now writes detailed logs to `logs/` directory
- **Improved console output**: Cleaner, more informative messages
- **Position debugging**: Verbose mode shows player positions to diagnose distance calculation issues

### Logging Improvements
- **Normal mode**: Shows only essential info (data received, switches)
- **Verbose mode**: Shows all details (GSI data, analysis, positions, decisions)
- **Log files**: Automatic log files in `logs/autoobserver_[timestamp].log`

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
- Run with verbose mode: `run.bat -v` or `better-autoobserver.exe -v`
- Normal mode: `run.bat` or `better-autoobserver.exe`
- Check logs: `logs/autoobserver_[timestamp].log`

## Version 1.0.0

### Initial Release
- Intelligent encounter prediction based on player distance
- Priority-based player switching
- Game State Integration (GSI) support
- Standalone .exe build
- PowerShell build scripts
