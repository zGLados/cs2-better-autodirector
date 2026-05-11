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
- **[Camera Priority System](CAMERA_PRIORITY.md)** - Understanding priority bonuses and camera switching logic

---

## 🎯 Features & Integration

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
│  └────────────┘  │  - PlayerAnalyzer    │  │
│                  └──────────┬───────────┘  │
└─────────────────────────────┼───────────────┘
                              │
              ┌───────────────┼───────────────┐
              │               │               │
         ┌────▼────┐                    ┌────▼─────┐
         │   GSI   │                    │Spectator │
         │ Server  │                    │Controller│
         │ :3000   │                    │  (Keys)  │
         └────┬────┘                    └──────────┘
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

---

## 📖 Configuration Files

### User Configuration
- `config/settings.json` - Camera priority and behavior settings
- `config/gamestate_integration_autodirector.cfg` - CS2 GSI config
- `config/spectator_bindings.cfg` - CS2 keybindings

---

## 🎮 Usage Workflow

### Basic Usage
1. Start CS2 and join as spectator
2. Execute `spectator_bindings` in console
3. Launch Auto Director GUI
4. Click "Start Auto Director"

---

## 🐛 Troubleshooting

### Common Issues

**"No game state data received"**
- Verify `gamestate_integration_autodirector.cfg` is in CS2 cfg folder
- Restart CS2 after config installation
- Check logs for GSI server status

**"Spectator controls not working"****
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
- CS2 Game State Integration

---

## 🔗 Links

- [GitHub Repository](https://github.com/zGLados/cs2-better-autodirector)
- [Latest Release](https://github.com/zGLados/cs2-better-autodirector/releases/latest)
- [Issue Tracker](https://github.com/zGLados/cs2-better-autodirector/issues)
