# 🎮 CS2 Better Auto Director

<div align="center">

```
╔══════════════════════════════════════════════════════════════╗
║        CS2 Better Auto Director                              ║
║                                                              ║
║  Intelligent automatic spectating                            ║
║  Automatically switches to exciting player encounters       ║
╚══════════════════════════════════════════════════════════════╝
```

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?style=flat&logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-Free-green?style=flat)](LICENSE)

</div>

---

## ⚡ Quick Start

### Option 1: Direct Build

```cmd
cd scripts
build.bat
```

Requires: [Go](https://go.dev/dl/) + [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/)

### Option 2: Professional Installer

See [docs/BUILD.md](docs/BUILD.md) for instructions on creating a professional installer using Inno Setup.

---

## 📂 Project Structure

```
cs2-better-autodirector/
├── 📄 main.go                    # Main program
├── 📄 gsi_server.go             # Game State Integration Server
├── 📄 player_analyzer.go        # Intelligent encounter detection
├── 📄 spectator_controller.go   # Keyboard simulation
├── 📄 go.mod                    # Go dependencies
├── 📄 README.md                 # This file
│
├── 📁 config/                   # Configuration files
│   └── gamestate_integration_autodirector.cfg
│
├── 📁 scripts/                  # Build & Setup Scripts
│   ├── build.bat               # Wrapper to run build.ps1
│   ├── build.ps1               # Main build script
│   ├── setup.ps1               # Setup with config copy
│   └── installer.iss           # Inno Setup Script
│
└── 📁 docs/                     # Documentation
    └── BUILD.md                # Build instructions
```

---

## ✨ Features

🎯 **Intelligent Encounter Prediction**
- Automatically detects when players from different teams are approaching each other
- Advanced priority system: distance (most important!), equipment value, kills, HP
- **Current player bonus:** Camera stays with players in active fights (+100 priority)
- Smart distance thresholds: Focuses on 0-2000 units (where kills actually happen)

⚡ **Smart Switching**
- Automatically switches to the most exciting encounters
- Prioritizes close-range fights (< 300 units = +150 priority!)
- **Camera stability:** Stays with current player if they're in action
- **Dead player detection:** Immediately switches away when spectated player dies
- **Phase-aware switching:**
  - 🔄 **Freezetime/Warmup**: 5 seconds max per player with team alternation for dynamic viewing
  - ⏸️ **Timeout**: 10 seconds max per player
  - 🎮 **Live rounds**: 15 seconds max (25s during active encounters)
  - ⚡ **Clutch situations** (≤4 players): 1 second rate limit for faster action
- Rate limiting: Adaptive (2s in freezetime, 1s in clutch, 2s normal)

📊 **Sophisticated Priority Algorithm**
- Distance-based (150 to 20 points based on range)
- Equipment value consideration
- Skill level tracking (kills × 3.0 multiplier)  
- Health status awareness
- Defuser bonus for bomb situations
- 👉 **[Full priority system documentation](docs/CAMERA_PRIORITY.md)**

🎯 **Advanced Event Detection**
- **Sniper System:**
  - Sniper duel detection (+100 bonus, +30 for AWP duels)
  - AWP kill tracking (+200 bonus, immediate switch)
  - Scout kill tracking (+150 bonus)
  - Smart switching (respects ongoing sniper action)
- **Damage Detection:**
  - Real-time HP tracking
  - Identifies damage dealers (20+ HP loss)
  - Smart filtering (excludes grenades, molotovs)
  - Immediate switch to shooter (+40 bonus for 5s)
- **Upset Victory:**
  - Detects underdog wins (+100 bonus)
  - Immediate switch to winner

🚀 **Standalone EXE**
- No installation required
- Single .exe file
- Small and performant (~8-15 MB)

---

## 🔧 Installation

### Method 1: Build from Source

For building from source, see [docs/BUILD.md](docs/BUILD.md) for detailed instructions.

**Quick version:**

1. **Install dependencies:**
   - [Go 1.21+](https://go.dev/dl/)
   - [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/)

2. **Build:**
   ```cmd
   cd scripts
   build.bat
   ```

3. **Done!** The `cs2-better-autodirector.exe` will be created in the root folder.

### Method 2: Professional Installer

For creating a professional installer with Inno Setup, see [docs/BUILD.md](docs/BUILD.md).

---

## 🎮 Usage

### 1. Install GSI Config

Copy the file `config/gamestate_integration_autodirector.cfg` to:
```
C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\
```

Or use `setup.ps1` from the scripts folder to automatically copy it.

### 2. Setup Spectator Keybinds ⚠️ IMPORTANT

Copy `config/spectator_bindings.cfg` to the same folder as above, then in CS2 console:
```
exec spectator_bindings
bind F9 spec_mode_toggle
```

**Why is this needed?**
- Auto Director uses keys 1-0 to switch between players
- Keys 1-0 are normally used for weapons, so we use a toggle system
- Press **F9** to switch between normal mode (weapons) and spectator mode (player slots)

**Usage:**
1. Join a match as spectator
2. Run in console: `exec spectator_bindings` (only needed once per CS2 session)
3. Bind toggle key: `bind F9 spec_mode_toggle` (you can use any key instead of F9)
4. Press **F9** to enable spectator mode → Keys 1-0 now select players
5. When done spectating, press **F9** again → Keys 1-0 back to weapons

**To verify bindings work:**
1. Press F9 to enable spectator mode (console shows: "SPECTATOR MODE ON")
2. Manually press keys 1-9, 0 on your keyboard
3. If the camera switches to different players → bindings work ✅
4. If nothing happens → re-run `exec spectator_bindings` ❌

### 3. Start the Program

**Quick Run (for development/testing):**
```cmd
run.bat              # Normal mode with minimal output
run.bat -v           # Verbose mode with detailed logs
```
This runs the program directly without building an .exe (faster for testing).

**Or use the compiled .exe:**
```cmd
cs2-better-autodirector.exe       # Normal mode
cs2-better-autodirector.exe -v    # Verbose mode
```

**Logging modes:**
- **Normal mode**: Shows only essential info (data received, player switches). Detailed logs are written to `logs/autodirector_[timestamp].log`
- **Verbose mode (-v)**: Shows all detailed logs in the console

### 4. Start CS & Spectate

1. Launch CS2
2. Enter **Spectator mode** (GOTV, demo, or as spectator on a server)
3. The program takes over automatically!

**Important:** The CS window must be in the foreground!

---

## 📊 How It Works

### Game State Integration (GSI)
CS automatically sends game data to `http://localhost:8000`:
- Player positions (3D coordinates)
- Health, armor, equipment
- Team affiliation
- Kills, deaths
- Round status

### Intelligent Analysis

**1. Distance Check**: Are players close enough?
   - < 500 Units: Immediate danger
   - < 1500 Units: Likely encounter
   - < 3000 Units: Potential encounter

**2. Priority Calculation**:
   - 🔴 **Very close** (< 500): +100 points
   - 🟡 **Close** (< 1500): +50 points
   - 🟢 **Medium** (< 3000): +10 points
   - 💰 **Equipment value**: +0-50 points
   - 🎯 **Kills**: +3 points per kill
   - ❤️ **Low HP**: +15-30 points
   - 🛠️ **Defuser**: +20 points

**3. Phase-Aware Behavior**:
   - 🔄 **Freezetime/Warmup**: Max 5 seconds per player, alternates between CT/T teams
   - ⏸️ **Timeout**: Max 10 seconds per player for variety
   - 🎮 **Live rounds**: Smart encounter tracking with extended viewing (15-25s)
   - ⚡ **Clutch mode**: Faster switching when ≤4 players alive

**4. Auto-Switch**: Switches to the player with the highest action priority

---

## 🔍 Debug Output & Logging

The program has **two output destinations** with smart behavior:

### 📺 Console Output

**Normal Mode (Default)** - Clean & Focused
```cmd
cs2-better-autodirector.exe
# or
.\scripts\run.bat
```

**Shows only important events:**
- ✅ Starting/stopping messages
- ✅ Data reception confirmations  
- ✅ Player switches (`➡️  Switching to: ...`)
- ✅ Encounters detected (`⚔️  ENCOUNTER: ...`)
- ✅ Critical errors or warnings

**Example console output:**
```
📋 Normal Mode
  → Console: Shows IMPORTANT events only
  → File:    Shows ALL detailed logs
  → Log file: logs/autodirector_2026-02-15_14-30-45.log
=====================================
[INFO] Starting CS2 Better Auto Director...
[INFO] GSI Server started on port 8000
[INFO] Data received: 10 players
⚔️  ENCOUNTER: Player1 (CT) vs Player2 (T) | Distance: 450 units | Priority: 125.5
➡️  Switching to: Player1 (Slot 3)
```

---

**Verbose Mode (`-v` flag)** - Everything Visible
```cmd
cs2-better-autodirector.exe -v
# or
.\scripts\run.bat -v
```

**Shows ALL events in console:**
- ✅ All GSI data reception
- ✅ Round phase changes
- ✅ Player position updates
- ✅ Encounter detection details
- ✅ Priority calculations
- ✅ Rate limiting info
- ✅ Debug messages

**Example verbose output:**
```
📊 Verbose Mode Enabled
  → Console: Shows ALL logs
  → File:    Shows ALL logs
=====================================
[GSI] Round Phase: live
[MAIN] Analyzing game state (phase: live)...
[ANALYZER] Found 10 alive players
[ANALYZER] ✓ Player Player1 position: X=1234.5 Y=567.8 Z=90.1
[ANALYZER] Detected 2 potential encounters
[ANALYZER] 🎯 Clutch situation (3 players alive) - sticky time limit disabled
⚔️  ENCOUNTER: Player1 (CT) vs Player2 (T) | Distance: 450 units | Priority: 125.5
[ANALYZER] → Switching to Player1 (score: 125.5 vs 98.3)
[CONTROLLER] Updating player slots for 10 players:
[CONTROLLER]   Slot 1: Player1 (CT) - HP:100 Eq:$4750 K:3
➡️  Switching to: Player1 (Slot 3)
[CONTROLLER] Pressing key '3' (hold method)...
[CONTROLLER] ✓ Key sequence completed (3x repetition)
```

---

### 📄 Log File Output

**⚡ IMPORTANT: Log files ALWAYS contain ALL details, regardless of console mode!**

```
Location: logs/autodirector_[timestamp].log
Content:  Complete detailed logs (same as verbose mode)
```

**This means:**
- 🎯 **Run without `-v`** = Clean console, detailed log file ✅ **RECOMMENDED**
- 🔍 **Run with `-v`** = Detailed console + detailed log file (for active debugging)

**Use the log file to:**
- Diagnose why switches didn't happen
- See position data and distance calculations
- Understand priority decisions
- Debug encounter detection
- Review complete game state history

---

## ⚙️ Configuration

### Adjust Switching Intervals

In [player_analyzer.go](player_analyzer.go):
```go
// Phase-specific switching timers
case "warmup", "freezetime":
    maxTimeOnPlayer = 5.0              // 5 seconds max in buy phase
    switchInterval = 2 * time.Second   // 2 second rate limit for better pacing
case "timeout":
    maxTimeOnPlayer = 10.0             // 10 seconds max during timeout
    switchInterval = 2 * time.Second
default:  // Live rounds
    maxTimeOnPlayer = 15.0             // 15 seconds max (25s during fights)
    switchInterval = 2 * time.Second   // Normal rate limit
```

### Adjust Encounter Distances

In [player_analyzer.go](player_analyzer.go#L126-L136):
```go
if distance < 500 {
    priority += 100  // Immediate encounter
} else if distance < 1000 {
    priority += 70
} // ...
```

### Update Frequency

In [main.go](main.go#L52):
```go
updateInterval := 500 * time.Millisecond  // Game state analysis frequency
```

After making changes, simply recompile with `scripts\build.bat`.

---

## 🐛 Troubleshooting

### "No data received"
- ✅ Check if `config/gamestate_integration_autodirector.cfg` is in the CS cfg folder
- ✅ Restart CS after copying the config file
- ✅ Check if port 8000 is free

### **"Switches not working" or "Switches only sometimes work"** ⚠️ IMPORTANT
This is the **most common issue**. The switches are being triggered but not reaching CS2.

**Required conditions:**
1. ✅ **CS2 window MUST be in FOREGROUND** (active/focused window)
   - Switches use keyboard simulation - only works on the active window
   - If you Alt+Tab away, switches won't work
   - Keep CS in focus while auto-director is running

2. ✅ **Run as Administrator** (recommended)
   - Right-click `cs2-better-autodirector.exe` → "Run as Administrator"
   - This improves keyboard input reliability

3. ✅ **Verify spectator keybinds work** (CRITICAL!)
   - Make sure you ran `exec spectator_bindings` in CS2 console
   - Make sure you pressed **F9** to enable spectator mode (console shows "SPECTATOR MODE ON")
   - In spectator mode, manually press keys 1, 2, 3, etc.
   - The camera should switch to different players
   - If manual keys don't work → **spectator mode not enabled or keybinds broken**
   - Solution: Re-run `exec spectator_bindings` and press F9
   - Remember: After each CS2 restart, you need to run `exec spectator_bindings` again

4. ✅ **Check verbose logs**

**Diagnostic steps:**
```cmd
# Run with verbose logging to see when switches are attempted
cs2-better-autodirector.exe -v
```

Look for these messages:
- `[INFO] ➡️  Switching to: PlayerName (Slot X)` - Switch was triggered
- `[CONTROLLER] ✓ Key sequence completed` - Keyboard simulation finished

**If switches work sometimes but not always:**
- This means CS loses focus intermittently
- Keep CS window in foreground consistently
- Don't click on other windows while program is running
- Consider using a second monitor to view logs

### "Distance: 0 units" or "No encounters detected"
- ✅ You must be in **GOTV/Demo playback mode**
- ✅ Won't work in live spectator mode (position data not sent)
- ✅ Run with `-v` flag to see position data: `run.bat -v`
  - If positions are all 0, GSI might not be sending position data
  - Make sure you're not in free camera mode

### Build error
- ✅ Install GCC if you see "GCC not found" message
- ✅ After Go installation: Open a **new terminal**
- ✅ The build script will guide you to download requirements

### "robotgo error" or "GCC not found"
1. Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/download/)
2. Install with default settings
3. Open a new terminal and run `scripts\build.bat` again

---

## 🎹 Key Bindings

The program uses the standard CS spectator keys:
- **1-9, 0**: Switches to player 1-10

Make sure these keys are not bound differently in CS!

---

## 🔧 Technical Details

- **Language**: Go 1.22+
- **Dependencies**:
  - `github.com/go-vgo/robotgo` - Keyboard simulation
  - Standard library for HTTP/JSON
- **Size**: ~8-15 MB (compiled with `-ldflags="-s -w"`)
- **Performance**: < 1% CPU usage

---

## 📄 License

Free to use for personal and commercial purposes.

---

## 🤝 Contributing

Forks and pull requests are welcome!

---

<div align="center">

**Enjoy intelligent spectating! 🎮**

Made with ❤️ for the CS Community

</div>
