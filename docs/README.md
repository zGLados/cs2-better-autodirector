# 📚 CS2 Better Auto Director - Documentation

Welcome to the documentation hub for CS2 Better Auto Director - an intelligent automatic spectating system for Counter-Strike 2.

---

## 🚀 Getting Started

### Quick Start
- **[Main README](../README.md)** - Installation and quick start guide
- **[Build Instructions](BUILD.md)** - Compile from source (Windows/Linux)

---

## ⚙️ Configuration

### Essential Setup
- **[Secrets Configuration](SECRETS.md)** - API keys and credentials setup
- **[Camera Priority System](CAMERA_PRIORITY.md)** - Understanding priority bonuses and camera switching logic

---

## 🎯 Features & Integration

### FACEIT Integration
- **[FACEIT Match Integration](FACEIT.md)** - Connect to FACEIT matches and stream GOTV
  - Automatic match data fetching
  - Team names and logos
  - GOTV server connection
  - **[Testing Guide](FACEIT_TESTING.md)** - Test API integration standalone

### Auto Director System
- **Camera Priority** - Intelligent player selection based on:
  - Weapon bonuses (AWP, Scout, AK-47)
  - Damage detection
  - Upset victories
  - Sniper kills
- **Real-time Analysis** - Live game state monitoring via GSI
- **Automatic Switching** - Smart camera transitions to exciting encounters

---

## 🏗️ Architecture

### Core Components

```
┌─────────────────────────────────────────────┐
│           GUI Dashboard (Wails)              │
│  ┌────────────┐  ┌──────────────────────┐  │
│  │  Frontend  │  │  Go Backend (App)    │  │
│  │ HTML/JS/CSS│◄─┤  - AutoDirector      │  │
│  └────────────┘  │  - FACEIT Client     │  │
│                  │  - PlayerAnalyzer    │  │
│                  └──────────┬───────────┘  │
└─────────────────────────────┼───────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
         ┌────▼────┐    ┌────▼─────┐   ┌────▼─────┐
         │   GSI   │    │ FACEIT   │   │Spectator │
         │ Server  │    │   API    │   │Controller│
         │ :3000   │    │          │   │  (Keys)  │
         └────┬────┘    └──────────┘   └──────────┘
              │
         ┌────▼────┐
         │   CS2   │
         │  Game   │
         └─────────┘
```

### Key Files
- `gui/main.go` - Entry point (GUI + CLI mode)
- `gui/app.go` - Main application logic
- `gui/autodirector.go` - Camera switching algorithm
- `gui/faceit_client.go` - FACEIT API integration
- `gui/player_analyzer.go` - Priority calculation
- `gui/spectator_controller.go` - CS2 keyboard control

---

## 🔧 Development

### Requirements
- Go 1.22+
- Wails v2
- Node.js (for frontend)

### Build Commands
```bash
# Development mode with hot reload
wails dev

# Production build
wails build

# Build installer (Windows)
./scripts/build-installer.bat
```

See [BUILD.md](BUILD.md) for detailed instructions.

### Testing

**Test FACEIT Integration:**
```bash
cd scripts
go run test_faceit.go -url "https://www.faceit.com/en/cs2/room/1-..."
```

Verifies API connection and match data fetching. See [FACEIT Testing Guide](FACEIT_TESTING.md).

---

## 📖 Configuration Files

### User Configuration
- `config/settings.json` - Camera priority and behavior settings
- `config/secrets.json` - API keys (not in version control)
- `config/gamestate_integration_autodirector.cfg` - CS2 GSI config
- `config/spectator_bindings.cfg` - CS2 keybindings

### Templates
- `config/secrets.example.json` - Template for secrets file

---

## 🎮 Usage Workflow

### Basic Usage
1. Start CS2 and join as spectator
2. Execute `spectator_bindings` in console
3. Launch Auto Director GUI
4. Click "Start Auto Director"

### With FACEIT Integration
1. Configure FACEIT API key in `config/secrets.json`
2. Launch Auto Director GUI
3. Paste FACEIT match room URL
4. Click "Fetch Match Data"
5. Copy GOTV connect command
6. Connect to GOTV in CS2
7. Start Auto Director

### Testing FACEIT Integration
Before using in production, test the API connection:
```bash
cd scripts
go run test_faceit.go -url "FACEIT_MATCH_URL"
```
See [FACEIT Testing Guide](FACEIT_TESTING.md) for detailed testing instructions.

---

## 🐛 Troubleshooting

### Common Issues

**"No game state data received"**
- Verify `gamestate_integration_autodirector.cfg` is in CS2 cfg folder
- Restart CS2 after config installation
- Check logs for GSI server status

**"FACEIT client not initialized"**
- Ensure `config/secrets.json` exists
- Verify API key is correct
- Restart application

**"Spectator controls not working"**
- Execute `exec spectator_bindings` in CS2 console
- Check keybindings in spectator mode

---

## 📝 Contributing

### Project Structure
```
cs2-better-autodirector/
├── config/              # Configuration files
├── docs/                # Documentation (you are here)
├── gui/                 # Main application
│   ├── frontend/        # Web UI
│   └── *.go            # Go backend
├── scripts/             # Build scripts
└── README.md           # Main entry point
```

### Adding New Features
1. Backend: Add methods to `gui/app.go`
2. Frontend: Update `gui/frontend/src/main.js`
3. Document: Update relevant docs
4. Test: Build and verify functionality

---

## 📄 License & Credits

See [LICENSE](../LICENSE) for details.

**Built with:**
- [Wails](https://wails.io/) - Go + Web GUI framework
- [FACEIT API](https://developers.faceit.com/) - Match data integration
- CS2 Game State Integration

---

## 🔗 Links

- [GitHub Repository](https://github.com/zGLados/cs2-better-autodirector)
- [Latest Release](https://github.com/zGLados/cs2-better-autodirector/releases/latest)
- [Issue Tracker](https://github.com/zGLados/cs2-better-autodirector/issues)
