# 🎮 Better Auto Observer

<div align="center">

```
╔══════════════════════════════════════════════════════════════╗
║        Better Auto Observer for Counter-Strike               ║
║                                                              ║
║  Intelligent automatic spectating                            ║
║  Automatically switches to exciting player encounters        ║
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
better-autoobserver/
├── 📄 main.go                    # Main program
├── 📄 gsi_server.go             # Game State Integration Server
├── 📄 player_analyzer.go        # Intelligent encounter detection
├── 📄 spectator_controller.go   # Keyboard simulation
├── 📄 go.mod                    # Go dependencies
├── 📄 README.md                 # This file
│
├── 📁 config/                   # Configuration files
│   └── gamestate_integration_autoobserver.cfg
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
- Rate limiting: 2 seconds between switches (bypassed on player death)

📊 **Sophisticated Priority Algorithm**
- Distance-based (150 to 20 points based on range)
- Equipment value consideration
- Skill level tracking (kills × 3.0 multiplier)  
- Health status awareness
- Defuser bonus for bomb situations
- 👉 **[Full priority system documentation](docs/CAMERA_PRIORITY.md)**

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

3. **Done!** The `better-autoobserver.exe` will be created in the root folder.

### Method 2: Professional Installer

For creating a professional installer with Inno Setup, see [docs/BUILD.md](docs/BUILD.md).

---

## 🎮 Usage

### 1. Install GSI Config

Copy the file `config/gamestate_integration_autoobserver.cfg` to:
```
C:\Program Files (x86)\Steam\steamapps\common\Counter-Strike Global Offensive\game\csgo\cfg\
```

Or use `setup.ps1` from the scripts folder to automatically copy it.

### 2. Start the Program

**Quick Run (for development/testing):**
```cmd
run.bat              # Normal mode with minimal output
run.bat -v           # Verbose mode with detailed logs
```
This runs the program directly without building an .exe (faster for testing).

**Or use the compiled .exe:**
```cmd
better-autoobserver.exe       # Normal mode
better-autoobserver.exe -v    # Verbose mode
```

**Logging modes:**
- **Normal mode**: Shows only essential info (data received, player switches). Detailed logs are written to `logs/autoobserver_[timestamp].log`
- **Verbose mode (-v)**: Shows all detailed logs in the console

### 3. Start CS & Spectate

1. Launch CS:GO or CS2
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

**3. Auto-Switch**: Switches to the player with the highest action priority

---

## 🔍 Debug Output

The program supports two logging modes:

### Normal Mode (Default)
**What you see in console:**
- Data reception confirmations
- Player switches
- Important events

**Detailed logs are written to:** `logs/autoobserver_[timestamp].log`

**Example console output:**
```
[INFO] Starting Better Auto Observer...
[INFO] GSI Server started on port 8000
[INFO] Data received: 10 players
⚔️  ENCOUNTER: Player1 (CT) vs Player2 (T) | Distance: 450 units | Priority: 125.5
➡️  Switching to: Player1 (Slot 3)
```

### Verbose Mode (`-v` flag)
**Shows everything in console:**
- All GSI data reception
- Round phase changes
- Player analysis details
- Encounter detection
- Rate limiting info
- Position data

**Example verbose output:**
```
[GSI] Round Phase: live
[MAIN] Analyzing game state (phase: live)...
[ANALYZER] Found 10 players
[ANALYZER] Player Player1 position: X=1234.5 Y=567.8 Z=90.1
[ANALYZER] Detected 2 potential encounters
⚔️  ENCOUNTER: Player1 (CT) vs Player2 (T) | Distance: 450 units | Priority: 125.5
[ANALYZER] → Switching to Player1 (better equipment/kills)
[CONTROLLER] Updating player slots for 10 players:
[CONTROLLER]   Slot 1: Player1 (CT) - HP:100 Eq:$4750 K:3
➡️  Switching to: Player1 (Slot 3)
```

**Use verbose mode for:**
- Debugging issues
- Understanding why switches happen (or don't)
- Seeing position data to diagnose distance calculation problems

---

## ⚙️ Configuration

### Adjust Switching Interval

In [player_analyzer.go](player_analyzer.go#L25):
```go
minSwitchInterval: 3 * time.Second  // Minimum time between switches
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
- ✅ Check if `config/gamestate_integration_autoobserver.cfg` is in the CS cfg folder
- ✅ Restart CS after copying the config file
- ✅ Check if port 8000 is free

### "Players not switching" or "Distance: 0 units"
- ✅ CS must be in the **foreground**
- ✅ You must be in **Spectator mode** (not as a player)
- ✅ Test if keys 1-9, 0 work manually
- ✅ Run with `-v` flag to see position data: `run.bat -v`
  - If positions are all 0, the GSI might not be sending position data
  - Make sure you're in GOTV/Demo playback, not free camera mode

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
